package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"

	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"
	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)

const localhost string = "127.0.0.1"

func main(){
	port :=  os.Args[1]
	_, err := MakeWorker(localhost + ":" + port)

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
	port string
}



func MakeWorker(port string) (*Worker, error) {
	guid, err := uuid.NewRandom()

	if err != nil{
		logger.LogError(logger.MASTER, "Error generationg uuid: %v", err)
		return nil, err
	}
	worker := &Worker{
		Id: guid.String(),
		port: port,
	}
	worker.server(port)

	return worker, nil
}




//
// RPC handlers
//

func (worker *Worker) HandleTasks(args *RPC.TaskArgs, reply *RPC.TaskReply) error {
	logger.LogInfo(logger.WORKER, "Recieved a task from master %v", args)
	return nil
}


//
// start a thread that listens for RPCs
//
func (worker *Worker) server(port string) error{
	rpc.Register(worker)
	rpc.HandleHTTP()

	
	os.Remove(port)
	listener, err := net.Listen("tcp", port)
	
	if err != nil {
		logger.LogError(logger.WORKER, "Error while listening on socket: %v", err)
		return err
	}

	logger.LogInfo(logger.WORKER, "Listening on socket: %v", port)

	go http.Serve(listener, nil)
	return nil
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

