package main //LockServer

import (
	logger "Distributed_Web_Crawler/Logger"
	"Distributed_Web_Crawler/Server/Cluster/RPC"
	database "Distributed_Web_Crawler/Server/LockServer/Database"
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

const MY_PORT string = 			"MY_PORT"
const MY_HOST string =			"MY_HOST"
const DB_PORT string = 			"DB_PORT"
const DB_HOST string = 			"DB_HOST"
const LOCAL_HOST string = 		"127.0.0.1"

var myHost string = 			getEnv(MY_HOST, LOCAL_HOST) 
var dbHost string = 			getEnv(DB_HOST, LOCAL_HOST)

var myPort string =  			os.Getenv(MY_PORT)
var dbPort string = 			os.Getenv(DB_PORT)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func main(){

	p, err := New(myHost, myPort, dbHost, dbPort)
	p.dbWrapper.DeleteAllRecords()
	
	if err != nil{
		logger.FailOnError(logger.LOCK_SERVER, logger.ESSENTIAL, "Exiting becuase of error creating a lockServer: %v", err)
	}


	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	
	sig := <- signalCh //block until user exits
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Received a quit sig %+v\nNow cleaning up and closing resources", sig)

}


type LockServer struct{
	id string
	address string
	dbWrapper *database.DBWrapper

	mu sync.Mutex
}


func New(myHost, myPort, dbHost, dbPort string) (*LockServer, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		return nil, err
	}

	lockServer := &LockServer{
		id: guid.String(),
		address: myHost + ":" + myPort,
		dbWrapper: database.New(dbHost, dbPort),

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

	
	os.Remove(lockServer.address)
	listener, err := net.Listen("tcp", lockServer.address)


	if err != nil {
		logger.FailOnError(logger.LOCK_SERVER, logger.ESSENTIAL, "Error while listening on socket: %v", err)

	}

	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Listening on socket: %v", lockServer.address)

	go http.Serve(listener, nil)
	return nil
}


//
// RPC handlers
//

func (lockServer *LockServer) HandleGetJobs(args *RPC.GetJobArgs, reply *RPC.GetJobReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "A master requested to be given a job %+v", args)
	reply.Accepted = false
	reply.AlternateJob = false

	//check if there are any late jobs (60 sec) that need to be reassigned
	secondsPassed := 60
	if lockServer.handleLateJobs(args, reply, secondsPassed){
		return nil
	}

	//no late jobs present


	if args.NoCurrentJob{    
		//this is a master who doesnt have a job from the queue, so is instead asking me for
		//an outstanding job
		return nil
	}


	//check the job isnt taken by any other master
	info := &database.Info{}
	lockServer.dbWrapper.GetRecordByJobId(info, args.JobId)

	if info.Id == 0{
		//no job present, can safely send assign this job to this master
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request accepted %v", args.JobId)
		lockServer.dbWrapper.AddRecord(&database.Info{
			JobId: args.JobId,
			ClientId: args.ClientId,
			MasterId: args.MasterId,
			UrlToCrawl: args.URL,
			DepthToCrawl: args.Depth,
			TimeAssigned: time.Now().Unix(),
		})
		reply.Accepted = true
		return nil
	}

	//job is present and another master has been assigned it
	//job can't be late since we have already handled the case where it is indeed late
	//we have to reject it, and we can't provide an alternative
	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %v", args.JobId)

	return nil
}

func (lockServer *LockServer) HandleFinishedJobs(args *RPC.FinishedJobArgs, reply *RPC.FinishedJobReply) error {
	logger.LogJobDone(logger.LOCK_SERVER, logger.ESSENTIAL, "A master just finished this job: \n%+v", args)

	info := &database.Info{}
	lockServer.dbWrapper.GetRecordByJobId(info, args.JobId)
	if info.Id == 0{
		//job not in the db, may have already been finished
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job already removed from database")
		return nil
	}

	lockServer.dbWrapper.DeleteRecord(info)
	logger.LogJobDone(logger.LOCK_SERVER, logger.ESSENTIAL, "Job finished successfully")
	return nil

}

func (lockServer *LockServer) handleLateJobs(args *RPC.GetJobArgs, reply *RPC.GetJobReply, secondsPassed int) bool{
	infos := []database.Info{}
	lockServer.dbWrapper.GetRecordsThatPassedXSeconds(&infos, secondsPassed)

	if len(infos) == 0{
		return false //no late jobs found
	}
	logger.LogDelay(logger.LOCK_SERVER, logger.ESSENTIAL, "Found a late job that will be reassigned %+v", infos[0])


	//found a job that had been assigned but is late
	info := &infos[0]
	reply.AlternateJob = true
	reply.JobId = info.JobId
	reply.ClientId = info.ClientId
	reply.URL = info.UrlToCrawl
	reply.Depth = info.DepthToCrawl
	
	//update the db with the details of the new master
	info.MasterId = args.MasterId
	info.TimeAssigned = time.Now().Unix()
	lockServer.dbWrapper.UpdateRecord(info)
	return true

}