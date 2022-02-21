package Client

import (
	logger "Distributed_Web_Crawler/Logger"
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

var MyHost string = 				getEnv(MY_HOST, LOCAL_HOST)
var MqHost string = 				getEnv(MQ_HOST, LOCAL_HOST)

var MyPort string =  				os.Getenv(MY_PORT)
var MqPort string =  				os.Getenv(MQ_PORT)


func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}


//each client's write goroutine is assigned a struct
type Client struct{
	Id string						//unique id
	JobResults chan string  		//channel to receive job results from server
	Conn *websocket.Conn			//websocket connection associated with client
	ConnTime int64					//epoch seconds at which conn created, all idle connections 
									//are terminated after MAX_IDLE_TIME

}

func NewClient(conn *websocket.Conn) (*Client, error){
	guid, err := uuid.NewRandom()
	if err != nil{
		logger.LogError(logger.SERVER, logger.NON_ESSENTIAL, "Error generationg uuid for new ws connection: %v", err)
		return nil, err
	}
	return &Client{
		Id: guid.String(),
		JobResults: make(chan string),
		Conn: conn,
		ConnTime: time.Now().Unix(),
	}, nil
}

