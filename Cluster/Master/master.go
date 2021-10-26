package main

import (
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"

	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)

const localhost string = "127.0.0.1"

func main(){
	port1 :=  os.Args[1]
	port2 := os.Args[2]


	master, err := MakeMaster([]string{localhost + ":" + port1, localhost + ":" + port2})
	
	if err != nil{
		logger.LogError(logger.CLUSTER, "Exiting becuase of error creating a master: %v", err)
		os.Exit(1)
	}


	//later on, will be using rabbit mq
	testUrl := "https://www.google.com/"
	websitesNum := 1000
	master.DoCrawl(testUrl, websitesNum)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait() 
}



type Master struct{
	id string


	workersAlive []bool  //map to keep track of live workers
	workerPorts []string  //map to keep track of socketName for each worker
	URLsMap map[string]bool //map of whether a url has been visited or not    
	workersTimers []time.Time  //keep track of last time a task was issued

	masterMu sync.Mutex
}


func MakeMaster(worker_ports []string) (*Master, error){
	guid, err := uuid.NewRandom()

	if err != nil{
		logger.LogError(logger.MASTER, "Error generationg uuid: %v", err)
		return nil, err
	}

	master := &Master{
		id: guid.String(),
		workersAlive: make([]bool, len(worker_ports)),
		workerPorts: worker_ports,
		URLsMap: make(map[string]bool),
		workersTimers: make([]time.Time, len(worker_ports)),
		masterMu: sync.Mutex{},
	}


	for k := range master.workerPorts{
		master.callWorkerWithTask(k, &RPC.TaskArgs{}, &RPC.TaskReply{})
	}


	return master, nil;

}

//in the future, will take the work from rabbitmq
func (master *Master) DoCrawl(url string, urlsNum int){

}

func (master *Master) callWorkerWithTask(workerIdx int, args *RPC.TaskArgs, reply *RPC.TaskReply) bool {

	client, err := rpc.DialHTTP("tcp", master.workerPorts[workerIdx])  //blocking
	if err != nil{
		logger.LogError(logger.MASTER, "Error dialing http: %v", err)
		return false
	}
	defer client.Close()

	err = client.Call("Worker.HandleTasks", args, reply)

	if err != nil{
		logger.LogError(logger.MASTER, "Unable to call worker with RPC with error: %v", err)
		return false
	}

	return true
}



// links, err := crawling.GetURLs("https://google.com")

// 	if err != nil{
// 		logger.LogError(logger.MASTER, "Error: %v", err)
// 	}

// 	logger.LogInfo("Links: %v", links)