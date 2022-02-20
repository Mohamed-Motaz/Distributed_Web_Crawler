package main

import (
	logger "Distributed_Web_Crawler/Logger"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const MY_PORT string = 				"MY_PORT"
const MY_HOST string =				"MY_HOST"
const MQ_PORT string = 				"MQ_PORT"
const MQ_HOST string = 				"MQ_HOST"
const LOCAL_HOST string = 			"127.0.0.1"

var myHost string = 				getEnv(MY_HOST, LOCAL_HOST) 
var mqHost string = 				getEnv(MQ_HOST, LOCAL_HOST)

var myPort string =  				os.Getenv(MY_PORT)
var mqPort string =  				os.Getenv(MQ_PORT)

const MAX_IDLE_TIME time.Duration = 10 * time.Minute





type LoggingMiddleware struct{
	handler http.Handler
}

//each client's write goroutine is assigned a struct
type Client struct{
	id string						//unique id
	jobResults chan string  		//channel to receive job results from server
	conn *websocket.Conn			//websocket connection associated with client
	connTime int64					//epoch seconds at which conn created, all idle connections 
									//are terminated after MAX_IDLE_TIME
	killChan chan struct{}

}


func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request){
	l.handler.ServeHTTP(w, r)
	logger.LogRequest(logger.SERVER, "Request received from %v", r.RemoteAddr)
}

func NewClient(conn *websocket.Conn) (*Client, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.SERVER, "Error generationg uuid for new ws connection: %v", err)
		return nil, err
	}
	return &Client{
		id: guid.String(),
		jobResults: make(chan string),
		conn: conn,
		connTime: time.Now().Unix(),
		killChan: make(chan struct{}),
	}, nil
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}