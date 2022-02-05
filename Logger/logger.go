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

const LOG_INFO = 0
const LOG_ERROR = 1

func LogInfo(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_INFO)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func LogError(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_ERROR)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func beautifyLogs(role int, format string, logType int) string {
	additionalInfo := determineRole(role)

	switch logType {
	case LOG_INFO:
		additionalInfo = Green + additionalInfo + "INFO: "
	case LOG_ERROR:
		additionalInfo = Red + additionalInfo + "ERROR: "
	default:
		additionalInfo = Blue + additionalInfo + "DEFAULT: "
	}
	
	additionalInfo += strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}
	format += "\n"

	return additionalInfo + format
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