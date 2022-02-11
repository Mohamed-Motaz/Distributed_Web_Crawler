package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	utils "server/Cluster/Utils"
	RPC "server/cluster/rpc"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	logger "server/cluster/logger"
	mq "server/messagequeue"
)

var domain string = "127.0.0.1"
const portEnv string = "PORT"
const 
(
	TaskAvailable = iota
	TaskAssigned
	TaskDone
)


const JOBS_QUEUE = "jobs"
const DONE_JOBS_QUEUE = "doneJobs"


type Job struct{
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
}

type DoneJob struct{
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
	Results [][]string		`json:"results"`
}


func main(){
	port :=  os.Getenv(portEnv)
	// logger.LogDebug(logger.WORKER, "the master port %v", port)

	master, err := New(domain + ":" + port)
	
	if err != nil{
		logger.FailOnError(logger.CLUSTER, "Exiting becuase of error creating a master: %v", err)
	}

	//subsribe to queue
	
	
	// 	mQ.Publish(JOBS_QUEUE, "hiiiiiii there")

	//later on, will be using rabbit mq
	testUrl := "https://www.google.com/"
	websitesNum := 3
	master.doCrawl(testUrl, websitesNum)

	// testUrl = "https://www.youtube.com/"
	// websitesNum = 2
	// master.doCrawl(testUrl, websitesNum)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	
	sig := <- signalCh //block until user exits
	logger.LogInfo(logger.MASTER, "Received a quit sig %+v\nNow cleaning up and closing resources", sig)
	master.q.Close()
}


type Master struct{
	id string
	port string

	jobNum int       						//job number to keep track of current job
	jobRequiredDepth int   					//required depth to traverse
						   					//min is 1, in which case I just crawl a single page and return the results

	currentJob bool  						//whether or not I am currently executing a job
	currentURL string   					//currentURL job to crawl
	currentDepth int 						//currentDepth, should not exceed jobRequiredDepth

	URLsTasks []map[string]int 				//slice -> for each depth, map of whether a url task has been done, assigned, or is available

	workersTimers []map[string]time.Time  	//keep track of last time a task was issued

	q *mq.MQ   								//message queue to publish and consume messages
	publishCh chan bool  					//ch to send finished jobs on
	consumeCh chan bool          			//ch to consume jobs on

	publishChAck chan bool  				//ch to send ack that finished job sent

	mu sync.Mutex
}


func New(port string) (*Master, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.MASTER, "Error generationg uuid: %v", err)
		return nil, err
	}

	master := &Master{
		id: guid.String(),
		port: port,
		jobNum: 0,
		jobRequiredDepth: 0,
		currentJob: false,
		currentURL: "",
		currentDepth: 0,
		URLsTasks: make([]map[string]int, 0),
		workersTimers: make([]map[string]time.Time, 0),
		q: mq.New("amqp://guest:guest@localhost:5672/"),  //os.Getenv("AMQP_URL"))
		publishCh: make(chan bool),
		consumeCh: make(chan bool),
		publishChAck: make(chan bool),
		mu: sync.Mutex{},
	}

	//initialize messageQ
	go master.qPublisher()
	go master.qConsumer()

	go master.server()
	go master.checkLateTasks()
	go master.checkJobDone()
	go master.debug()
	

	return master, nil;
}




//in the future, will take the work from rabbitmq
func (master *Master) doCrawl(url string, depth int) {
	//TODO clear up all the data from the previous crawl

	logger.LogInfo(logger.MASTER, 
		"Recieved a request to crawl this website %v with a depth of %v",
		url, depth)

	master.mu.Lock()
	defer master.mu.Unlock()

	master.resetMasterJobStatus(depth)

	// master.URLsTasks["https://www.google.com/"] = TaskAvailable  //set task as available to be given to servers
	// master.URLsTasks["https://www.youtube.com/"] = TaskAvailable  //set task as available to be given to servers
	// master.URLsTasks["https://www.facebook.com/"] = TaskAvailable  //set task as available to be given to servers



	master.currentURL = url
	master.currentJob = true
	master.currentDepth = 0

	master.jobNum++
	master.jobRequiredDepth = depth

	master.URLsTasks[0][url] = TaskAvailable  //set task as available to be given to servers
	
}

//
// hold lock --
// clear all data from previous crawl and get ready for new one
//
func (master *Master) resetMasterJobStatus(depth int){
	master.jobRequiredDepth = 0

	master.currentJob = false
	master.currentURL = ""
	master.currentDepth = 0

	master.URLsTasks = make([]map[string]int, depth)
	master.workersTimers = make([]map[string]time.Time, depth)

	for i := 0; i < depth; i++{
		master.URLsTasks[i] = make(map[string]int)
		master.workersTimers[i] = make(map[string]time.Time)
	}
}


//
// start a thread that listens for RPCs
//
func (master *Master) server() error{
	rpc.Register(master)
	rpc.HandleHTTP()

	
	os.Remove(master.port)
	listener, err := net.Listen("tcp", master.port)


	if err != nil {
		logger.FailOnError(logger.MASTER, "Error while listening on socket: %v", err)

	}

	logger.LogInfo(logger.MASTER, "Listening on socket: %v", master.port)

	go http.Serve(listener, nil)
	return nil
}

//
// start a thread that keeps an eye on if any tasks are late
//
func (master *Master) checkLateTasks() {

	for {	
		master.mu.Lock()

		if (!master.currentJob || master.currentDepth >= master.jobRequiredDepth){
			master.mu.Unlock()
			time.Sleep(time.Second)
			continue
		}

		for k,v := range master.URLsTasks[master.currentDepth]{
			if v == TaskAssigned{
				if time.Since(master.workersTimers[master.currentDepth][k]) > time.Second * 20 {
					logger.LogDelay(logger.MASTER, "Found a server that was asleep with this url %v", k)
					//a server hasnt replied for 20 seconds
					master.URLsTasks[master.currentDepth][k] = TaskAvailable   //set his task to be available
				}
			}
			
		}		
		master.mu.Unlock()

		time.Sleep(time.Second)
	}
	
}

//
//hold lock --
//
func (master *Master) urlInTasks(url string, parentUrl string) bool{
	//need to check if url already in our tasks
	for i := 0; i <= master.currentDepth; i++{
		if _, ok := master.URLsTasks[i][url]; ok{
			return true
		}	
	}
	return false
}

//
//hold lock --
//check if job is available or not
//if not, increment current depth only after making sure 
//all elements in the current depth are finished
//
func (master *Master) checkJobAvailable(reply *RPC.GetTaskReply){
	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogInfo(logger.MASTER, "The worker was given no tasks as none are available %v", reply)
		return
	} 

	currentDepthFinished := true
	for url, jobStatus := range master.URLsTasks[master.currentDepth] {			
		if jobStatus != TaskDone{
			currentDepthFinished = false
		}	
		if jobStatus == TaskAvailable{
			//send this job and mark it as assigned
			master.URLsTasks[master.currentDepth][url] = TaskAssigned
			master.workersTimers[master.currentDepth][url] = time.Now()
			reply.URL = url
			reply.JobNum = master.jobNum
			logger.LogInfo(logger.MASTER, "The worker was given this task %v", reply)
			return
		}
	}

	//no job found, can try to move on to the next depth
	if currentDepthFinished{
		logger.LogMilestone(logger.MASTER, "Milestone reached, depth %v has been finished", master.currentDepth)
		master.currentDepth++
		master.checkJobAvailable(reply)
		return
	}
}

//
// start a thread that checks if currentJob is done
//
func (master *Master) checkJobDone() {

	for {	
		master.mu.Lock()
		if master.currentJob && master.currentDepth >= master.jobRequiredDepth{
			//job is finished
			logger.LogInfo(logger.MASTER, "Done with job with url %v, depth %v, and jobNum %v", 
			master.currentURL, master.jobRequiredDepth, master.jobNum)

			//now send it over channel to qPublisher rabbit mq
			


			master.publishCh <- true
			<- master.publishChAck  //block till Publisher sends the message to the queue or fails to do so

			master.resetMasterJobStatus(0)
		}

		master.mu.Unlock()
		time.Sleep(time.Second)
	}
	
}

//
// hold lock ---
// start a thread that listens for a finished job
// and then publishes it to the message queue
//
func (master *Master) qPublisher() {

	for {
		select{
		case <- master.publishCh:
			URLsList := utils.ConvertMapArrayToList(master.URLsTasks)
			
			fin := &DoneJob{
				UrlToCrawl: master.currentURL,
				DepthToCrawl: master.jobRequiredDepth,
				Results: URLsList,
			}
			
			res, err := json.Marshal(fin)
			if err != nil{
				logger.LogError(logger.MASTER, "Unable to convert result to string! Discarding...")
			}else{
				master.q.Publish(DONE_JOBS_QUEUE, res)
				if err != nil{
					logger.LogError(logger.MASTER, "DoneJob not published to queue with err %v", err)
				}
			}
			
			master.publishChAck <- true
			
		default:
			time.Sleep(time.Second)
		}


	}

}

//
// start a thread that waits on a job from the message queue
//
func (master *Master) qConsumer() {
	// ch, err := master.q.Consume(JOBS_QUEUE)
	// if err != nil{
	// 	logger.FailOnError(logger.MASTER, "Master can't consume jobs because with this error %v", err)
	// }

	for {
		select{
		// case newJob := <- ch:
		// 	body := string(newJob.Body)
			
			
		default:
			time.Sleep(time.Second)
		}


	}

}

func (master *Master) debug(){
	for {
		
		master.mu.Lock()
		for i,e  := range master.URLsTasks{
			logger.LogDebug(logger.MASTER, "This is the map with len %v \n", len(e))
			taskAvailable := 0
			taskAssigned := 0
			taskDone := 0
			for _,v := range e{
				if v == TaskAvailable{taskAvailable++}
				if v == TaskAssigned{taskAssigned++}
				if v == TaskDone{taskDone++} 
			}
			logger.LogDebug(logger.MASTER, "URLsTasks num %v: \n" + 
			"TasksAvailable:  %v\nTasksAssigned: %v\nTasksDone: %v",i, taskAvailable, taskAssigned, taskDone)
		}
		master.mu.Unlock()
		time.Sleep(time.Second * 5)
	}

}



//
// find an empty port between 8000 and 9000 to listen on
//
func generatePortNum() (int, *net.Listener, error){
	for i := 8000; i <= 9000; i++{
		listener, err := net.Listen("tcp", domain + ":" + strconv.Itoa(i))
		if err != nil{
			continue
		}
		//successfully found a free port
		return i, &listener, nil
	}
	return -1,nil, fmt.Errorf("unable to find an empty port")
}





//
// RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, "A worker requested to be given a task %v", args)
	reply.JobNum = -1
	reply.URL = ""

	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.currentJob{
		logger.LogInfo(logger.MASTER, 
			"A worker requested to be given a task but we have no jobs at the moment")
		return nil
	} 

	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogDelay(logger.MASTER, 
			"A worker requested to be given a task but we have already finished the job")
		return nil
	} 

	master.checkJobAvailable(reply)

	return nil
}

func (master *Master) HandleFinishedTasks(args *RPC.FinishedTaskArgs, reply *RPC.FinishedTaskReply) error {
	master.mu.Lock()
	defer master.mu.Unlock()

	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogDelay(logger.MASTER, 
			"A worker has finished a task %v but we have already finished the job", args.URL)
		return nil
	} 

	if !master.currentJob {
		logger.LogDelay(logger.MASTER, 
			"A worker finished a late task %v for job num %v but there is no current job",
			args.URL, args.JobNum)
		return nil
	}

	if (master.URLsTasks[master.currentDepth][args.URL] == TaskDone){
		//already finished, do nothing
		logger.LogDelay(logger.MASTER, "Worker has finished this task with jobNum %v and URL %v " +
							"which was already finished", args.JobNum, args.URL)
		return nil
	}	

	logger.LogTaskDone(logger.MASTER, "A worker just finished this task: \n" +
		"JobNum: %v \nURL: %v \nURLsLen :%v", 
		args.JobNum, args.URL, len(args.URLs))

	master.URLsTasks[master.currentDepth][args.URL] = TaskDone //set the task as done


	if master.currentDepth + 1 < master.jobRequiredDepth{
		//add all urls to the URLsTasks next depth and set their tasks as available
		for _, v := range args.URLs{
			if (!master.urlInTasks(v, args.URL)){
				master.URLsTasks[master.currentDepth + 1][v] = TaskAvailable
			}	
		}
	}
	
	return nil
}

