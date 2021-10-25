package master

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"

	crawling "github.com/mohamed247/Distributed_Web_Crawler/Functionality"
)
type Master struct{

}



func MakeMaster() *Master{
	master := &Master{}
	master.server()

	

	links, err := crawling.GetURLs("https://google.com")
	fmt.Printf("%v %v", links ,err)

	return master;

}

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
func (master *Master) server() {
	rpc.Register(master)
	rpc.HandleHTTP()
	portNum, listener, err := generatePortNum()

	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("Serving on port num %v \n", portNum)

	go http.Serve(*listener, nil)
}