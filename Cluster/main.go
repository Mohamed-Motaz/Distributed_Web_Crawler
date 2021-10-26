package main

import (
	"os"
	"sync"

	"github.com/google/uuid"
	master "github.com/mohamed247/Distributed_Web_Crawler/Cluster/Master"
	worker "github.com/mohamed247/Distributed_Web_Crawler/Cluster/Worker"
	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)

const localhost = "127.0.0.1";


func main(){
	logger.LogInfo(logger.CLUSTER, "Setting up Cluster")
	
	socketName, err := uniqueSocket()
	
	if err != nil{
		logger.LogError(logger.CLUSTER, "Exiting becuase of error: %v", err)
		os.Exit(1)
	}

	worker.MakeWorker(socketName)
	master.MakeMaster(socketName)

	wg := sync.WaitGroup{}
	wg.Wait() 
}



// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp
func uniqueSocket() (string, error) {
	s := "/var/tmp/distributedWebCrawler"
	uid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.CLUSTER, "Error while creating uuid: %v", err)
		return "", err
	}
	s += uid.String()
	return s, nil
}



