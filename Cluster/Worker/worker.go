package main

import (
	"net/rpc"
	"os"
	"sync"

	"github.com/google/uuid"

	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"
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
	}
	worker.callMaster("Master.HandleGetTasks", &RPC.GetTaskArgs{}, &RPC.GetTaskReply{})
	return worker, nil
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


