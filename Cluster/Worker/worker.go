package worker

import (
	"net"
	"net/http"
	"net/rpc"
	"os"

	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"
	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)



type Worker struct {
	id int64
}



func MakeWorker(socketName string) *Worker{

	worker := &Worker{

	}
	worker.server(socketName)

	return worker
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
func (worker *Worker) server(socketName string) error{
	rpc.Register(worker)
	rpc.HandleHTTP()

	
	os.Remove(socketName)
	listener, err := net.Listen("unix", socketName)
	
	if err != nil {
		logger.LogError(logger.WORKER, "Error while listening on socket: %v", err)
		return err
	}

	logger.LogInfo(logger.WORKER, "Listening on socket: %v", socketName)

	go http.Serve(listener, nil)
	return nil
}

//
// find an empty port between 8000 and 9000 to listen on
//
// func generatePortNum() (int, *net.Listener, error){
// 	for i := 8000; i <= 9000; i++{
// 		listener, err := net.Listen("tcp", localhost + ":" + strconv.Itoa(i))
// 		if err != nil{
// 			continue
// 		}
// 		//successfully found a free port
// 		return i, &listener, nil
// 	}
// 	return -1,nil, fmt.Errorf("unable to find an empty port")
// }
