package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"
	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
	utils "github.com/mohamed247/Distributed_Web_Crawler/Utils"
)

const localhost string = "127.0.0.1"

func main(){
	port :=  os.Args[1]


	master, err := MakeMaster(localhost + ":" + port)
	
	if err != nil{
		logger.LogError(logger.CLUSTER, "Exiting becuase of error creating a master: %v", err)
		os.Exit(1)
	}


	//later on, will be using rabbit mq
	testUrl := "https://www.google.com/"
	websitesNum := 2
	master.doCrawl(testUrl, websitesNum)

	// testUrl = "https://www.youtube.com/"
	// websitesNum = 2
	// master.doCrawl(testUrl, websitesNum)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait() 
}

const 
(
	TaskAvailable = iota
	TaskAssigned
	TaskDone
)

type Master struct{
	id string
	port string

	jobNum int       //job number to keep track of current job
	jobRequiredDepth int   //required depth to traverse
						   //min is 1, in which case I just crawl a single page and return the results

	currentJob bool  //whether or not I am currently executing a job
	currentURL string   //currentURL job to crawl
	currentDepth int //currentDepth, should not exceed jobRequiredDepth

	URLsTasks []map[string]int //slice -> for each depth, map of whether a url task has been done, assigned, or is available

	workersTimers []map[string]time.Time  //keep track of last time a task was issued

	mu sync.Mutex
}


func MakeMaster(port string) (*Master, error){
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
		mu: sync.Mutex{},
	}

	go master.server()
	go master.checkLateTasks()
	go master.checkJobDone()
	//go master.debug()


	return master, nil;

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
			"A worker requested to be given a task but we have already  " +
			"finished the job, cur depth is %v and required is %v",
			master.currentDepth, master.jobRequiredDepth )
		
		return nil
	} 

	master.checkJobAvailable(reply)

	return nil
}

func (master *Master) HandleFinishedTasks(args *RPC.FinishedTaskArgs, reply *RPC.FinishedTaskReply) error {
	logger.LogInfo(logger.MASTER, "A worker just finished this task: \n" +
								  "JobNum: %v \nURL: %v \nURLsLen :%v", 
									args.JobNum, args.URL, len(args.URLs))

	master.mu.Lock()
	defer master.mu.Unlock()

	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogDelay(logger.MASTER, 
			"A worker has finished a task but we have already  " +
			"finished the job, cur depth is %v and required is %v",
			master.currentDepth, master.jobRequiredDepth )
		return nil
	} 

	if !master.currentJob {
		logger.LogDelay(logger.MASTER, 
			"A worker finished a late task for job num %v but the job is already finished",
			args.JobNum)
		return nil
	}


	//if not already visited, then can set it as done
	if (!master.isUrlVisited(args)){
		master.URLsTasks[master.currentDepth][args.URL] = TaskDone //set the task as done

		if master.currentDepth + 1 < master.jobRequiredDepth{
			//add all urls to the URLsTasks next depth and set their tasks as available
			for _, v := range args.URLs{
				master.URLsTasks[master.currentDepth + 1][v] = TaskAvailable
			}
		}
		
	}
	logger.LogInfo(logger.MASTER, "Current URLsTasks len after checking worker's respose is %v",len(master.URLsTasks))

	return nil
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
// clear all data from previous crawl and get ready for new one
// hold lock --
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
		logger.LogError(logger.MASTER, "Error while listening on socket: %v", err)
		return err
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
//hold lock
//
func (master *Master) isUrlVisited(args *RPC.FinishedTaskArgs) bool{
	//need to check if url already visited
	for i := 0; i <= master.currentDepth; i++{
		if (master.URLsTasks[i][args.URL] == TaskDone){
			//already finished, do nothing
			logger.LogDelay(logger.MASTER, "Worker has finished this task with jobNum %v and URL %v " +
								"which was already finished", args.JobNum, args.URL)
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
			//now send it over rabbit mq
			logger.LogInfo(logger.MASTER, "Done with job with url %v, depth %v, and jobNum %v", 
			master.currentURL, master.jobRequiredDepth, master.jobNum)
			URLsList := utils.ConvertMapArrayToList(master.URLsTasks)
			logger.LogInfo(logger.MASTER, "Here is the old data len %v %+v", len(URLsList), URLsList)

			logger.LogInfo(logger.MASTER, "Here is the data len %v %+v", len(URLsList), URLsList)
			master.resetMasterJobStatus(0)
		}

		master.mu.Unlock()
		time.Sleep(time.Second)
	}
	
}

func (master *Master) debug(){
	for {
		
			logger.LogInfo(logger.MASTER, "This is the map \n")
			master.mu.Lock()
			logger.LogInfo(logger.MASTER, "URLsTasks: \n %+v \n", master.URLsTasks)
			master.mu.Unlock()
			time.Sleep(time.Second * 2)
	}

}



//
// find an empty port between 8000 and 9000 to listen on
//
func generatePortNum() (int, *net.Listener, error){
	for i := 8000; i <= 9000; i++{
		listener, err := net.Listen("tcp", localhost + ":" + strconv.Itoa(i))
		if err != nil{
			continue
		}
		//successfully found a free port
		return i, &listener, nil
	}
	return -1,nil, fmt.Errorf("unable to find an empty port")
}

