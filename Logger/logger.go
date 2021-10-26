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

func LogInfo(role int, format string, a ...interface{}){
	
	var additionalInfo string

	if role == MASTER{
		additionalInfo = "MASTER-> "
	}else if role == WORKER{
		additionalInfo = "WORKER-> "
	}else if role == CLUSTER{
		additionalInfo = "CLUSTER-> "
	}

	additionalInfo += "INFO: " + strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}

	if debug == 1{
		fmt.Printf(additionalInfo + format, a...)
	}
}

func LogError(role int, format string, a ...interface{}){
	var additionalInfo string

	if role == MASTER{
		additionalInfo = "MASTER-> "
	}else if role == WORKER{
		additionalInfo = "WORKER-> "
	}else if role == CLUSTER{
		additionalInfo = "CLUSTER-> "
	}

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