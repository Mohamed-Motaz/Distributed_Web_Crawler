package logger

import (
	"fmt"
	"strconv"
	"time"
)

const debug = 1;
const log = 0;


func LogInfo(format string, a ...interface{}){
	timestamp := "INFO: " + strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}

	if debug == 1{
		fmt.Printf(timestamp + format, a...)
	}
}

func LogError(format string, a ...interface{}){
	timestamp := "ERROR: " + strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}

	if debug == 1{
		fmt.Printf(timestamp + format, a...)
	}
}


func makeTimestamp() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}