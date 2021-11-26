package logger

import (
	"fmt"
	"strconv"
	"time"
)

const debug = 1;
const log = 0;

const MASTER = 0
const WORKER = 1
const CLUSTER = 3
const CRAWLING = 4
const DATABASE = 5

func LogInfo(role int, format string, a ...interface{}){
	additionalInfo := determineRole(role)
	additionalInfo += "INFO: " + strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}

	if debug == 1{
		fmt.Printf(additionalInfo + format, a...)
	}
}

func LogError(role int, format string, a ...interface{}){
	additionalInfo := determineRole(role)
	additionalInfo += "ERROR: " + strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}

	if debug == 1{
		fmt.Printf(additionalInfo + format, a...)
	}
}


func makeTimestamp() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}

func determineRole(role int) string{

	switch role{
	case MASTER:
		return "MASTER-> "
	case WORKER:
		return "WORKER-> "
	case CLUSTER:
		return "CLUSTER-> "
	case CRAWLING:
		return "CRAWLING-> "
	default:
		return "UNKNOWN -> "
	}
}