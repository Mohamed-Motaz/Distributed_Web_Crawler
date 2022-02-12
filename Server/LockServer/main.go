package main //LockServer

import (
	logger "Server/Cluster/Logger"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

var domain string = "127.0.0.1"
const portEnv string = "PORT"

func main(){
	
	port :=  os.Getenv(portEnv)

	_, err := New(domain + ":" + port)
	
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
