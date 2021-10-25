package master

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"

	"github.com/google/uuid"

	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
)
type Master struct{
	id string
	workersAlive []bool
}



func MakeMaster() (*Master, error){
	master := &Master{}
	guid, err := uuid.NewRandom()

	if err != nil{
		logger.LogError("Error generationg uuid: %v", err)
		return master, err
	}
	
	master.id = guid.String()

	
	

	master.server()
	return master, nil;

}

//
// find an empty port between 8000 and 9000 to listen on
//
func generatePortNum() (int, *net.Listener, error){
	for i := 8000; i <= 9000; i++{
		listener, err := net.Listen("tcp", ":" + strconv.Itoa(i))
		if err != nil{
			continue
		}
		//successfully found a free port
		return i, &listener, nil
	}
	return -1,nil, fmt.Errorf("unable to find an empty port")
}

//
// start a thread that listens for RPCs
//
func (master *Master) server() error{
	rpc.Register(master)
	rpc.HandleHTTP()
	portNum, listener, err := generatePortNum()

	if err != nil {
		logger.LogError("Listener error: %v", err)
		return err
	}

	logger.LogInfo("Serving on port num %v \n", portNum)

	go http.Serve(*listener, nil)

	return nil;
}


// links, err := crawling.GetURLs("https://google.com")

// 	if err != nil{
// 		logger.LogError("Error: %v", err)
// 	}

// 	logger.LogInfo("Links: %v", links)