package main

import (
	logger "Distributed_Web_Crawler/Logger"
	"net/http"
	"os"
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

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

type LoggingMiddleware struct{
	handler http.Handler
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request){
	l.handler.ServeHTTP(w, r)
	logger.LogRequest(logger.SERVER, "Request received from %v", r.RemoteAddr)
}
