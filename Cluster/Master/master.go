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
	websitesNum := 1000
	master.doCrawl(testUrl, websitesNum)

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
	jobRequiredUrlsLen int
	currentJob bool  //whether or not I am currently executing a job
	URLsMap map[string]bool     //urls gotten so far
	URLsVisited map[string]bool //map of whether a url has been visited or not    
	URLsTasks map[string]int //map of whether a url task has been done, assigned, or is available

	workersTimers map[string]time.Time  //keep track of last time a task was issued

	masterMu sync.Mutex
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
		URLsVisited: make(map[string]bool),
		workersTimers: make(map[string]time.Time),
		masterMu: sync.Mutex{},
	}

	go master.server()
	go master.checker()


	return master, nil;

}



//
// RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, "A worker requested to be given a task %v", args)

	if len(master.URLsMap) > master.jobRequiredUrlsLen{
		logger.LogInfo(logger.MASTER, 
			"A worker requested to be given a task but we have already  " +
			"finished the job, cur length is %v and required is %v",
			len(master.URLsMap), master.jobRequiredUrlsLen )
		reply.JobNum = -1
		reply.URL = ""
		return nil

	} 

	//the reason behind looping on links from the beginning every time is
	//so since some of the previous tasks given out may have not been actually done

	for k, _ := range master.URLsMap {				
		//a task is indeed available to distribute

		if master.URLsTasks[k] == TaskAvailable{
			//send this job and mark it as assigned
			master.URLsTasks[k] = TaskAssigned
			master.workersTimers[k] = time.Now()
			reply.URL = k
			reply.JobNum = master.jobNum
			logger.LogInfo(logger.MASTER, "The worker was given this task %v", reply)
			return nil
		}
	}
		
	return nil
}

func (master *Master) HandleFinishedTasks(args *RPC.FinishedTaskArgs, reply *RPC.FinishedTaskReply) error {
	logger.LogInfo(logger.MASTER, "A worker just finished this task %v", args)

	reply.JobRecievedSuccessfully = true


	if len(master.URLsMap) > master.jobRequiredUrlsLen{
		logger.LogInfo(logger.MASTER, 
			"A worker has finished a task but we have already  " +
			"finished the job, cur length is %v and required is %v",
			len(master.URLsMap), master.jobRequiredUrlsLen )
		return nil
	} 

	if !master.currentJob {
		logger.LogInfo(logger.MASTER, 
			"A worker requested to be given a task for job num %v but the hob is already finished and we " +
			"have started another on job num %v",
			args.JobNum, master.currentJob )
		return nil
	}



	//need to check if url already visited
	if (master.URLsVisited[args.URL]){
		//already finished, do nothing
		logger.LogInfo(logger.MASTER, "Worker has finished this task %v which was already assigned to another server", args)
	}else{
		master.URLsVisited[args.URL] = true; //assign the page as visited
		
		//add all urls to the URLsMap and set their tasks as available
		for _, v := range args.URLs{
			master.URLsMap[v] = true
			master.URLsTasks[v] = TaskAvailable
		}
	}

	return nil
}




//in the future, will take the work from rabbitmq
func (master *Master) doCrawl(url string, urlsNum int) {
	//TODO clear up all the data from the previous crawl

	logger.LogInfo(logger.MASTER, 
		"Recieved a request to crawl this website %v and get %v links from it",
		url, urlsNum)

	master.currentJob = true
	master.URLsTasks[url] = TaskAvailable  //set task as available to be given to servers
	master.jobNum++
	master.jobRequiredUrlsLen = urlsNum
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
func (master *Master) checker() {
	for k,_ := range master.URLsTasks{
		if master.URLsTasks[k] == TaskAssigned{
			if time.Since(master.workersTimers[k]) > time.Second * 10 {
				//a server hasnt replied for 10 seconds
				master.URLsTasks[k] = TaskAvailable   //set his task to be available
			}
		}
		
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

