package Client

import (
	logger "Distributed_Web_Crawler/Logger"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//each client is is assigned a struct

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

