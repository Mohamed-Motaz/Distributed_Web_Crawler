package RPC

//
// RPC definitions
//

type GetTaskArgs struct{
	WokerId string   //worker current id for master to keep track
	URL string   //url to crawl

}

type GetTaskReply struct{}

type FinishedTaskArgs struct{
	URL string
	URLs []string
}

type FinishedTaskReply struct{}





