package main

import (
	logger "Distributed_Web_Crawler/Logger"
	"net/http"

	"github.com/gorilla/websocket"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024, //requests are usually small
	WriteBufferSize: 1024 * 16, //response is usually pretty large
}

//each client's write goroutine is assigned a struct
type Client struct{
	id string						//unique id
	jobResults chan string  		//channel to receive job results from server
	jobAssigned chan string			//channel to send job to server
}


func main(){
	mux := http.NewServeMux()
	mux.HandleFunc("/job", serveWS)
	logger.LogInfo(logger.SERVER, "Listening on %v:%v", myHost, mqPort)
	err := http.ListenAndServe(myHost + ":" + myPort, &LoggingMiddleware{mux})
	if err != nil{
		logger.FailOnError(logger.SERVER, "Failed in listening on port with error %v", err)
	}

}

func serveWS(w http.ResponseWriter, r *http.Request) {
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

}
