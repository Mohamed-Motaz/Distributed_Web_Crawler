package main

import (
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"
	crawling "github.com/mohamed247/Distributed_Web_Crawler/Functionality"
	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)

const localhost string = "127.0.0.1"

func main(){
	masterPort :=  os.Args[1]
	_, err := MakeWorker(localhost + ":" + masterPort)

	if err != nil{
		logger.LogError(logger.CLUSTER, "Exiting becuase of error with worker creation: %v", err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait() 
}



type Worker struct {
	Id string
	masterPort string
	currentJob bool   //currently working on a job
	currentJobNum int //num of currentJob, -1 if none
	currentURL string //current string to crawl
	currentFinishedURLs []string //urls finished crawling and not yet sent to master

	mu sync.Mutex
}



func MakeWorker(port string) (*Worker, error) {
	guid, err := uuid.NewRandom()

	if err != nil{
		logger.LogError(logger.MASTER, "Error generationg uuid: %v", err)
		return nil, err
	}
	worker := &Worker{
		Id: guid.String(),
		masterPort: port,
		currentJob: false,
		currentJobNum: -1,
		currentURL: "",
		currentFinishedURLs: make([]string, 0),
		mu: sync.Mutex{},
	}
	go worker.askForJobLooper()

	return worker, nil
}

func (worker *Worker) askForJob(){
	logger.LogInfo(logger.WORKER, "Worker asking for a job from master")
	reply := &RPC.GetTaskReply{}
	ok := worker.callMaster("Master.HandleGetTasks", &RPC.GetTaskArgs{}, reply)
	
	if !ok {
		logger.LogError(logger.WORKER, "Error while sending handleGetTasks to master")
		return
	}
	logger.LogInfo(logger.WORKER, "Worker recieved this reponse after asking for a job %v", reply)

	worker.mu.Lock()

	if reply.JobNum == -1 {
		worker.mu.Unlock()
		worker.currentJob = false
		worker.currentJobNum = -1
		logger.LogInfo(logger.WORKER, "Worker didnt recieve a job from master as there are currently none available")
		return
	}

	//recieved a valid job

	worker.currentJob = true
	worker.currentJobNum = reply.JobNum
	worker.currentURL = reply.URL

	worker.mu.Unlock()

	go worker.doTask(reply.URL)
}

func (worker *Worker) doTask(URL string){

	links, err := worker.doCrawl(URL)

	worker.mu.Lock()
	defer worker.mu.Unlock()
	if err != nil {
		//cant do current job, so discard it
		worker.resetWorkerJobStatus()
	}else{
		//say that u finished the job

		worker.currentFinishedURLs = links   //set those to be finished to send them again if u cant
		
		args := &RPC.FinishedTaskArgs{
			URL: URL,
			JobNum: worker.currentJobNum,
			URLs: links,
		}

		ok := worker.callMaster("Master.HandleFinishedTasks", args, &RPC.FinishedTaskReply{})

		if !ok{
			logger.LogError(logger.WORKER, "Error while sending handleFinishedTasks to master")
			return   //unable to send the rpcs
		}

		//successfully sent jobs to master and finished tasks, now mark currentJob as false

		worker.resetWorkerJobStatus()
	}
}

//
// 	reset job to false for when a job is finished or am unable to complete for some reason
//	hold lock
//
func (worker *Worker) resetWorkerJobStatus(){
	worker.currentJob = false
	worker.currentJobNum = -1
	worker.currentURL = ""
	worker.currentFinishedURLs = []string{}
}

func (worker *Worker) doCrawl(URL string) ([]string, error){
	links, err := crawling.GetURLsSlice(URL)
	if err != nil{
		return nil, err
	}
	return links, nil
}


//
// thread to sleep and wake up every second to ask for a job if there currently is none
//
func (worker *Worker) askForJobLooper(){

	for {

		worker.mu.Lock()
		if !worker.currentJob{
			worker.mu.Unlock()
			worker.askForJob()
		}else{
			worker.mu.Unlock()
			//there is a currentJob running at the moment
			//need to check for the corner case where I finished a job and unfortunately wasnt able
			//to send it or discard after set amount of time
		}

		time.Sleep(time.Second)
	}
}


func (worker *Worker) callMaster(rpcName string, args interface{}, reply interface{}) bool {

	client, err := rpc.DialHTTP("tcp", worker.masterPort)  //blocking
	if err != nil{
		logger.LogError(logger.WORKER, "Error dialing http: %v", err)
		return false
	}
	defer client.Close()

	err = client.Call(rpcName, args, reply)

	if err != nil{
		logger.LogError(logger.WORKER, "Unable to call master with RPC with error: %v", err)
		return false
	}

	return true
}


