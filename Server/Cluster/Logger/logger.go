package Logger

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const debug = 1;
const log = 0;

const (
	MASTER = iota
	WORKER 
	CLUSTER 
	CRAWLING 
	DATABASE 
	MESSAGE_Q
)

const (
	LOG_INFO = iota
	LOG_ERROR 
	LOG_DELAY 
	LOG_DEBUG 
	LOG_TASK_DONE
	LOG_JOB_DONE
	LOG_MILESTONE 
)

func FailOnError(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_ERROR)
	fmt.Printf(format, a...)
	os.Exit(1)
}

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

func LogDelay(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_DELAY)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func LogDebug(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_DEBUG)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func LogTaskDone(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_TASK_DONE)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func LogJobDone(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_JOB_DONE)

	if debug == 1{
		fmt.Printf(format, a...)
	}
}

func LogMilestone(role int, format string, a ...interface{}){
	format = beautifyLogs(role, format, LOG_MILESTONE)

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
	case LOG_DELAY:
		additionalInfo = Yellow + additionalInfo + "DELAY: "
	case LOG_DEBUG:
		additionalInfo = Purple + additionalInfo + "DEBUG: "
	case LOG_TASK_DONE:
		additionalInfo = Cyan + additionalInfo + "TASK_DONE: "
	case LOG_JOB_DONE:
		additionalInfo = White + additionalInfo + "JOB_DONE: "
	case LOG_MILESTONE:
		additionalInfo = Blue + additionalInfo + "MILESTONE: "
	default:
		additionalInfo = Blue + additionalInfo + "DEFAULT: "
	}
	
	additionalInfo += strconv.Itoa(int(makeTimestamp())) + " -> "

	if format[len(format) - 1] != '\n'{
		format += "\n"
	}
	format += Reset + "\n" //reset the terminal color

	
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
	case MESSAGE_Q:
		return "MESSAGE_Q-> "
	default:
		return "UNKNOWN -> "
	}
}