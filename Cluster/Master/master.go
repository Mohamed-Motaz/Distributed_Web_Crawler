package master

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/google/uuid"
	RPC "github.com/mohamed247/Distributed_Web_Crawler/Cluster/RPC"

	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)

var socketName string



type Master struct{
	id string


	workersAlive map[string]bool  //map to keep track of live workers
	URLsMap map[string]bool //map of whether a url has been visited or not    
	workersTimers map[string]time.Time  //keep track of last time a task was issued

	masterMu sync.Mutex
}




func MakeMaster(socketNamee string) (*Master, error){
	socketName = socketNamee
	guid, err := uuid.NewRandom()

	if err != nil{
		logger.LogError(logger.MASTER, "Error generationg uuid: %v", err)
		return nil, err
	}

	master := &Master{
		id: guid.String(),
		workersAlive: make(map[string]bool),
		URLsMap: make(map[string]bool),
		workersTimers: make(map[string]time.Time),
		masterMu: sync.Mutex{},
	}


	
	callWorkerWithTask(&RPC.TaskArgs{}, &RPC.TaskReply{})


	return master, nil;

}

func callWorkerWithTask(args *RPC.TaskArgs, reply *RPC.TaskReply) bool {

	client, err := rpc.DialHTTP("unix", socketName)  //blocking
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