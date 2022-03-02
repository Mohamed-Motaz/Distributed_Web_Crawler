package main

import (
	cl "Distributed_Web_Crawler/ClientFacingServer/Client"
	logger "Distributed_Web_Crawler/Logger"
	mq "Distributed_Web_Crawler/MessageQueue"
	rc "Distributed_Web_Crawler/RedisCache"
	utils "Distributed_Web_Crawler/Utils"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)
const MY_PORT string = 				"MY_PORT"
const MY_HOST string =				"MY_HOST"
const MQ_PORT string = 				"MQ_PORT"
const MQ_HOST string = 				"MQ_HOST"
const CACHE_PORT string = 			"CACHE_PORT"
const CACHE_HOST string = 			"CACHE_HOST"
const LOCAL_HOST string = 			"127.0.0.1"

var MyHost string = 				utils.GetEnv(MY_HOST, LOCAL_HOST) 
var MqHost string = 				utils.GetEnv(MQ_HOST, LOCAL_HOST)
var CacheHost string = 				utils.GetEnv(CACHE_HOST, LOCAL_HOST)

var MyPort string =  				utils.GetEnv(MY_PORT, "5555")
var MqPort string =  				utils.GetEnv(MQ_PORT, "5672")
var CachePort string =  			utils.GetEnv(CACHE_PORT, "6379")

const MAX_IDLE_CONNECTION_TIME time.Duration			= 5 * time.Minute
const MAX_IDLE_CACHE_TIME time.Duration 				= 60 * time.Minute


var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024, //requests are usually small
	WriteBufferSize: 1024 * 16, //response is usually pretty large
}

type Server struct{
	handler http.Handler					//logging requests middleware
	cache *rc.Cache							//wrapper around redis
	cachedJobsCh chan mq.DoneJob				//channel to send along doneJobs that are present in cache
	q *mq.MQ   								//message queue to publish and consume messages
	connsMap map[string]*cl.Client  		//keep track of all clients and their connections, to be able to send on them
	mu sync.RWMutex
}

func main(){
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }  //allow cross-origin requests	
	server, err := New()
	if (err != nil){
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "Error while creating server %v", err)
	}
	
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	
	sig := <- signalCh //block until user exits
	logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Received a quit sig %+v\nNow cleaning up and closing resources", sig)
	
	//close all connections
	server.mu.RLock()
	for _, v := range server.connsMap{
		v.Conn.Close()
	}
	server.mu.RUnlock()
}

func New() (*Server, error) {


	server := &Server{
		q: mq.New("amqp://guest:guest@" + MqHost + ":" + MqPort + "/"), 
		cache: rc.New(CacheHost, CachePort),
		cachedJobsCh: make(chan mq.DoneJob),
		connsMap: make(map[string]*cl.Client),
		mu: sync.RWMutex{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.serveWS)
	server.handler = mux

	go server.serve()
	go server.qConsumer()
	go server.cleaner()
	go server.debug()

	return server, nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request){
	server.serveWS(w, r)
	logger.LogRequest(logger.SERVER, logger.ESSENTIAL, "Request received from %v", r.RemoteAddr)
}

func (server *Server) serve(){
	logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)
	err := http.ListenAndServe(MyHost + ":" + MyPort, server)
	if err != nil{
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "Failed in listening on port with error %v", err)
	}
}

func (server *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.LogError(logger.SERVER, logger.ESSENTIAL, "Unable to upgrade http request to use websockets with error %v", err)
		return
	}
	
	client, err := cl.NewClient(conn)
	if err != nil{
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "Unable to create client with error %v", err)
		return
	}

	server.mu.Lock()
	server.connsMap[client.Id] = client
	server.mu.Unlock()

	go server.reader(client)
}



//write given data to a given connection
func (server *Server) writer(conn *websocket.Conn, data interface{}){
	conn.WriteJSON(data)
}

//remove all connections idle for more than 1 hour
func (server *Server) cleaner(){
	for {
		time.Sleep(5 * time.Second)

		idleClients := make([]string, 0)
		server.mu.RLock()
		for k, v := range server.connsMap{
			if  time.Since(time.Unix(v.ConnTime, 0)) > MAX_IDLE_CONNECTION_TIME{
				logger.LogDelay(logger.SERVER, logger.ESSENTIAL, "Found an idle connection with client %v, about to delete it", v.Conn.RemoteAddr())
				idleClients = append(idleClients, k)
			}
		}
		server.mu.RUnlock()

		if len(idleClients) > 0{
			server.mu.Lock()
			for _, id := range idleClients{
				delete(server.connsMap, id)
			}
			server.mu.Unlock()
		}
		
	}
}

//
// start a thread that waits on a doneJobs from the message queue
//
func (server *Server) qConsumer() {
	ch, err := server.q.Consume(mq.DONE_JOBS_QUEUE)

	if err != nil{
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "Server can't consume doneJobs because with this error %v", err)
	}

	for {		
		select{
		case doneJob := <- ch:  //job has been finished and pushed to queue
			body := doneJob.Body
			data := &mq.DoneJob{}
			
			err := json.Unmarshal(body, data)
			if err != nil {
				logger.LogError(logger.SERVER, logger.ESSENTIAL, "Unable to unMarshal job %v with error %v\nWill discard it", string(body),err) 
				doneJob.Ack(false)
				continue
			}

			//a job has been finished, now need to push it over
			//appropriate connection
			logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Job %+v consumed from message queue", data) 

			server.mu.RLock()
			tmp, ok := server.connsMap[data.ClientId]
			var client cl.Client
			if ok{
				client = *tmp			 		//copy client in memory, rather than change the pointer directly, to avoid a possible data race
			}
			server.mu.RUnlock()


			if ok{
				//send results to client over conn
				go server.writer(client.Conn, data)
				logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Job sent to client %+v\n%+v", client.Conn.RemoteAddr(), data) 

			}else{
				logger.LogError(logger.SERVER, logger.ESSENTIAL, "Connection with client may have been terminated", data) 
				doneJob.Nack(false, true)  
			} 

			doneJob.Ack(false)

			//add  job to cache

			toCache := &rc.CachedObj{
				UrlToCrawl: data.UrlToCrawl,
				DepthToCrawl: data.DepthToCrawl,
				Results: data.Results,
			}

			server.cache.AddIfNoLargerResultPresent(data.UrlToCrawl, toCache, MAX_IDLE_CACHE_TIME)

		case cachedJob := <- server.cachedJobsCh:
				
			//a job has been found in cache, now need to push it over
			//appropriate connection

			server.mu.RLock()
			tmp, ok := server.connsMap[cachedJob.ClientId]
			var client cl.Client
			if ok{
				client = *tmp			 		//copy client in memory, rather than change the pointer directly, to avoid a possible data race
			}
			server.mu.RUnlock()

			if ok{
				//send results to client over conn
				go server.writer(client.Conn, cachedJob)
				logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Job sent to client %+v\n%+v", client.Conn.RemoteAddr(), cachedJob) 

			}else{
				logger.LogError(logger.SERVER, logger.ESSENTIAL, "Job %+v done, but connection with client has been terminated", cachedJob) 
			} 

		default:
			time.Sleep(time.Second * 2)
		}	
	}
}

//read client job requests, and dump them to rabbit mq
func (server *Server) reader(c *cl.Client){
	defer c.Conn.Close()

	for{
		
		c.Conn.SetReadDeadline(time.Now().Add(MAX_IDLE_CONNECTION_TIME))  //can be idle for at most 10 mins
		_, message, err := c.Conn.ReadMessage()
		if err != nil{
			logger.LogError(logger.SERVER, logger.NON_ESSENTIAL, "Error with client %v\n%v", c.Conn.RemoteAddr(), err)
			return
		}

		//update client time
		server.mu.Lock()
		c.ConnTime = time.Now().Unix()  //lock this operation since cleaner is running
										//and may check on c.connTime
		newJob := &mq.Job{}
		newJob.ClientId = c.Id //need to set clientId of the newJob as this client's id
		server.mu.Unlock()

		//read message into json
		err = json.Unmarshal(message, newJob)
		if err != nil{
			logger.LogError(logger.SERVER, logger.NON_ESSENTIAL, "Error with client %v\n%v", c.Conn.RemoteAddr(), err)
			return
		}

		//make sure job not present in cache
		res := server.cache.GetCachedJobIfPresent(newJob.UrlToCrawl, newJob.DepthToCrawl, MAX_IDLE_CACHE_TIME)
		if res != nil{
			//found in cache
			logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "New job found in cache %+v", res)
			doneJob := &mq.DoneJob{
				ClientId: c.Id,
				JobId: newJob.JobId,
				UrlToCrawl: newJob.UrlToCrawl,
				DepthToCrawl: newJob.DepthToCrawl,
				Results: res.Results[:newJob.DepthToCrawl],  //cut the results to the size requested by the user
			}
			//send it over channel to be sent immediatley
			go func(job *mq.DoneJob){
				server.cachedJobsCh <- *job
			}(doneJob)
			continue
		}

		toPublish := new(bytes.Buffer)
		err = json.NewEncoder(toPublish).Encode(newJob)
		if err != nil{
			logger.LogError(logger.SERVER, logger.ESSENTIAL, "Error with client %v\n%v", c.Conn.RemoteAddr(), err)
			return
		}
		//message is viable and isnt present in cache, can now send it over to mq
		err = server.q.Publish(mq.JOBS_ASSIGNED_QUEUE, toPublish.Bytes())
		if err != nil{
			logger.LogError(logger.SERVER, logger.ESSENTIAL, "New job not published to jobs assigned queue with err %v", err)
		}else{
			logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "New job successfully published to jobs assigned queue")
		}
		
	}
}

func (server *Server) debug(){
	for{
		time.Sleep(5 * time.Second)
		server.mu.RLock()
		logger.LogDebug(logger.SERVER, logger.NON_ESSENTIAL, "ConnsMap\n")
		ctr := 1
		for k,v := range server.connsMap{
			logger.LogDebug(logger.SERVER, logger.NON_ESSENTIAL, "%v -- %v -- %v", ctr, k, v.Conn.RemoteAddr())
			ctr++
		}
		server.mu.RUnlock()
	}
}

