package main

import (
	logger "Distributed_Web_Crawler/Logger"
	mq "Distributed_Web_Crawler/MessageQueue"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)
const MY_PORT string = 				"MY_PORT"
const MY_HOST string =				"MY_HOST"
const MQ_PORT string = 				"MQ_PORT"
const MQ_HOST string = 				"MQ_HOST"
const LOCAL_HOST string = 			"127.0.0.1"

var MyHost string = 				getEnv(MY_HOST, LOCAL_HOST) 
var MqHost string = 				getEnv(MQ_HOST, LOCAL_HOST)

var MyPort string =  				os.Getenv(MY_PORT)
var MqPort string =  				os.Getenv(MQ_PORT)

const MAX_IDLE_TIME time.Duration = 10 * time.Minute



func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}




var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024, //requests are usually small
	WriteBufferSize: 1024 * 16, //response is usually pretty large
}


type LoggingMiddleware struct{
	handler http.Handler
}

type Server struct{
	q *mq.MQ   								//message queue to publish and consume messages
	connsMap map[string]*Client  			//keep track of all clients and their connections, to be able to send on them
	mu sync.RWMutex
}

var server Server

func main(){	
	server, err := New()
	if (err != nil){
		logger.FailOnError(logger.SERVER, "Error while creating server %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/job", server.serveWS)

	logger.LogInfo(logger.SERVER, "Listening on %v:%v", MyHost, MqPort)
	
	err = http.ListenAndServe(MyHost + ":" + MyPort, &LoggingMiddleware{mux})
	if err != nil{
		logger.FailOnError(logger.SERVER, "Failed in listening on port with error %v", err)
	}
}

func New() (*Server, error) {
	server := &Server{
		q: mq.New("amqp://guest:guest@" + MqHost + ":" + MqPort + "/"),  //os.Getenv("AMQP_URL"))
		connsMap: make(map[string]*Client),
		mu: sync.RWMutex{},
	}

	go server.qConsumer()
	go server.cleaner()

	return server, nil
}

//logging middleware
func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request){
	l.handler.ServeHTTP(w, r)
	logger.LogRequest(logger.SERVER, "Request received from %v", r.RemoteAddr)
}


func (server *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.LogError(logger.SERVER, "Unable to upgrade http request to use websockets with error %v", err)
		return
	}
	
	client, err := NewClient(conn)
	if err != nil{
		logger.FailOnError(logger.SERVER, "Unable to create client with error %v", err)
		return
	}

	server.mu.Lock()
	server.connsMap[client.Id] = client
	server.mu.Unlock()

	go client.reader()
}



//write given data to a given connection
func (server *Server) writer(conn *websocket.Conn, data interface{}){
	conn.WriteJSON(data)
}

//remove all connections idle for more than 1 hour
func (server *Server) cleaner(){
	for {
		time.Sleep(5 * time.Second)

		now := time.Now().Unix()
		idleClients := make([]string, 0)
		server.mu.RLock()
		for k, v := range server.connsMap{
			if  now - v.connTime > int64(MAX_IDLE_TIME){
				logger.LogDelay(logger.SERVER, "Found an idle connection with client %v, about to delete it", v.conn.RemoteAddr())
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
		logger.FailOnError(logger.SERVER, "Server can't consume doneJobs because with this error %v", err)
	}

	for {		
		select{
		case doneJob := <- ch:  //job has been finished and pushed to queue

			body := doneJob.Body
			data := &mq.DoneJob{}
			
			err := json.Unmarshal(body, data)
			if err != nil {
				logger.LogError(logger.MASTER, "Unable to unMarshal job with error %v\nWill discard it", err) 
				doneJob.Ack(false)
				continue
			}

			//a job has been finished, now need to push it over
			//appropriate connection

			server.mu.RLock()
			client, ok := server.connsMap[data.ClientId]
			server.mu.RUnlock()


			if ok{
				//send results to client over conn
				go server.writer(client.conn, data)
			} 
			//else, connection with client has already been terminated

			doneJob.Ack(false)

			//TODO add  job to cache
			
		default:
			logger.LogInfo(logger.SERVER, "No jobs found, about to sleep") 
			time.Sleep(time.Second * 5)
		}	
	}
}



//CLIENT  

//each client's write goroutine is assigned a struct
type Client struct{
	Id string						//unique id
	jobResults chan string  		//channel to receive job results from server
	conn *websocket.Conn			//websocket connection associated with client
	connTime int64					//epoch seconds at which conn created, all idle connections 
									//are terminated after MAX_IDLE_TIME
	killChan chan struct{}

}

func NewClient(conn *websocket.Conn) (*Client, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.SERVER, "Error generationg uuid for new ws connection: %v", err)
		return nil, err
	}
	return &Client{
		Id: guid.String(),
		jobResults: make(chan string),
		conn: conn,
		connTime: time.Now().Unix(),
		killChan: make(chan struct{}),
	}, nil
}

//read client job requests, and dump them to rabbit mq
func (c *Client) reader(){
	defer c.conn.Close()

	for{
		select{
		case <- c.killChan:
			//die
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(MAX_IDLE_TIME))  //can be idle for at most 10 mins
			_, message, err := c.conn.ReadMessage()
			if err != nil{
				logger.LogError(logger.SERVER, "Error %v with client %v", err, c.conn.RemoteAddr())
				return
			}

			//update client time
			server.mu.Lock()
			c.connTime = time.Now().Unix()  //lock this operation since cleaner is running
											//and may check on c.connTime
			server.mu.Unlock()

			//read message
			newJob := &mq.Job{}
			err = json.Unmarshal(message, newJob)
			if err != nil{
				logger.LogError(logger.SERVER, "Error %v with client %v", err, c.conn.RemoteAddr())
				return
			}

			//TODO
			//make sure job not present in cache
			

			//message is viable, can now send it over to mq
			err = server.q.Publish(mq.JOBS_QUEUE, message)
			if err != nil{
				logger.LogError(logger.SERVER, "New job not published to queue with err %v", err)
			}else{
				logger.LogInfo(logger.SERVER, "New job successfully published to queue")
			}
		}
	}
}