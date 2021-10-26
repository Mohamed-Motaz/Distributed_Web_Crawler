package main



func main(){
	// logger.LogInfo(logger.CLUSTER, "Setting up Cluster")
	
	// socketName1, err := uniqueSocket()
	
	// if err != nil{
	// 	logger.LogError(logger.CLUSTER, "Exiting becuase of error with socketNames: %v", err)
	// 	os.Exit(1)
	// }

	// socketName2, err := uniqueSocket()
	
	// if err != nil{
	// 	logger.LogError(logger.CLUSTER, "Exiting becuase of error with socketNames: %v", err)
	// 	os.Exit(1)
	// }

	// worker1, err := worker.MakeWorker(socketName1)
	// worker2, err := worker.MakeWorker(socketName2)

	// workerSockets := make(map[string]string)
	// workerSockets[worker1.Id] = socketName1
	// workerSockets[worker2.Id] = socketName2

	// master, err := master.MakeMaster(workerSockets)
	
	// if err != nil{
	// 	logger.LogError(logger.CLUSTER, "Exiting becuase of error creating a master: %v", err)
	// 	os.Exit(1)
	// }


	// //later on, will be using rabbit mq
	// testUrl := "https://www.google.com/"
	// websitesNum := 1000
	// master.DoCrawl(testUrl, websitesNum)

	// wg := sync.WaitGroup{}
	// wg.Wait() 
}


// // Cook up a unique-ish UNIX-domain socket name
// // in /var/tmp
// func uniqueSocket() (string, error) {
// 	s := "/var/tmp/distributedWebCrawler"
// 	uid, err := uuid.NewRandom()
// 	if err != nil{
// 		logger.LogError(logger.CLUSTER, "Error while creating uuid: %v", err)
// 		return "", err
// 	}
// 	s += uid.String()
// 	return s, nil
// }



