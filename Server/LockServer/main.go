package main //LockServer

import (
	logger "Server/Cluster/Logger"
	"Server/Cluster/RPC"
	database "Server/LockServer/Database"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

var domain string = "127.0.0.1"
const portEnv string = "PORT"

func main(){
	
	port :=  os.Getenv(portEnv)
	l, err := New(domain + ":" + port)
	
	//database.ManualTesting(l.dbWrapper);
	info := &database.Info{}
	l.dbWrapper.GetRecordByJobId(info, "NNEWWWW")
	fmt.Printf("%+v\n", info)


	infos := []database.Info{}
	l.dbWrapper.GetRecordsThatPassedXSeconds(&infos, 20)
	fmt.Printf("%+v\n", infos)

	if err != nil{
		logger.FailOnError(logger.LOCK_SERVER, "Exiting becuase of error creating a lockServer: %v", err)
	}


	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	
	sig := <- signalCh //block until user exits
	logger.LogInfo(logger.LOCK_SERVER, "Received a quit sig %+v\nNow cleaning up and closing resources", sig)

}


type LockServer struct{
	id string
	port string
	dbWrapper *database.DBWrapper

	mu sync.Mutex
}


func New(port string) (*LockServer, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.LOCK_SERVER, "Error generationg uuid: %v", err)
		return nil, err
	}

	lockServer := &LockServer{
		id: guid.String(),
		port: port,
		dbWrapper: database.New(),

		mu: sync.Mutex{},
	}

	go lockServer.server()

	return lockServer, nil;
}



//
// start a thread that listens for RPCs
//
func (lockServer *LockServer) server() error{
	rpc.Register(lockServer)
	rpc.HandleHTTP()

	
	os.Remove(lockServer.port)
	listener, err := net.Listen("tcp", lockServer.port)


	if err != nil {
		logger.FailOnError(logger.LOCK_SERVER, "Error while listening on socket: %v", err)

	}

	logger.LogInfo(logger.LOCK_SERVER, "Listening on socket: %v", lockServer.port)

	go http.Serve(listener, nil)
	return nil
}


//
// RPC handlers
//

func (lockServer *LockServer) HandleGetJobs(args *RPC.GetJobArgs, reply *RPC.GetJobReply) error {
	logger.LogInfo(logger.LOCK_SERVER, "A master requested to be given a job %v", args)
	reply.Accepted = false
	reply.AlternateJob = false

	//check if there are any late jobs (60 sec) that need to be reassigned
	if lockServer.handleLateJobs(args, reply){
		return nil
	}

	//no late jobs present


	//check the job isnt taken by any other master
	info := &database.Info{}
	lockServer.dbWrapper.GetRecordByJobId(info, args.JobId)

	if info.Id == 0{
		//no job present, can safely send assign this job to this master
		reply.Accepted = true
		return nil
	}

	//job is present and another master has been assigned it
	//job can't be late since we have already handled the case where it is indeed late
	//we have to reject it, and we can't provide an alternative

	return nil
}

func (lockServer *LockServer) HandleFinishedjobs(args *RPC.FinishedJobArgs, reply *RPC.FinishedJobReply) error {
	lockServer.mu.Lock()
	defer lockServer.mu.Unlock()

	// if lockServer.currentDepth >= lockServer.jobRequiredDepth{
	// 	logger.LogDelay(logger.LOCK_SERVER, 
	// 		"A worker has finished a job %v but we have already finished the job", args.URL)
	// 	return nil
	// } 

	// if !lockServer.currentJob {
	// 	logger.LogDelay(logger.LOCK_SERVER, 
	// 		"A worker finished a late job %v for job num %v but there is no current job",
	// 		args.URL, args.JobNum)
	// 	return nil
	// }

	// if (lockServer.URLsjobs[lockServer.currentDepth][args.URL] == jobDone){
	// 	//already finished, do nothing
	// 	logger.LogDelay(logger.LOCK_SERVER, "Worker has finished this job with jobNum %v and URL %v " +
	// 						"which was already finished", args.JobNum, args.URL)
	// 	return nil
	// }	

	// logger.LogjobDone(logger.LOCK_SERVER, "A worker just finished this job: \n" +
	// 	"JobNum: %v \nURL: %v \nURLsLen :%v", 
	// 	args.JobNum, args.URL, len(args.URLs))

	// lockServer.URLsjobs[lockServer.currentDepth][args.URL] = jobDone //set the job as done


	// if lockServer.currentDepth + 1 < lockServer.jobRequiredDepth{
	// 	//add all urls to the URLsjobs next depth and set their jobs as available
	// 	for _, v := range args.URLs{
	// 		if (!lockServer.urlInjobs(v, args.URL)){
	// 			lockServer.URLsjobs[lockServer.currentDepth + 1][v] = jobAvailable
	// 		}	
	// 	}
	// }
	
	return nil
}

func (lockServer *LockServer) handleLateJobs(args *RPC.GetJobArgs, reply *RPC.GetJobReply) bool{
	secondsPassed := 60
	infos := []database.Info{}
	lockServer.dbWrapper.GetRecordsThatPassedXSeconds(&infos, secondsPassed)

	if len(infos) == 0{
		return false //no late jobs found
	}


	//found a job that had been assigned but is late
	info := &infos[0]
	reply.AlternateJob = true
	reply.JobId = info.JobId
	reply.URL = info.UrlToCrawl
	reply.Depth = info.DepthToCrawl
	
	//update the db with the details of the new master
	info.MasterId = args.MasterId
	info.TimeAssigned = time.Now().Unix()
	lockServer.dbWrapper.UpdateRecord(info)
	return true

}