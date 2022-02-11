package RPC

//
// RPC definitions
//

type GetTaskArgs struct{}

type GetTaskReply struct{
	URL string   //url to crawl
	JobNum int
}

type FinishedTaskArgs struct{
	JobNum int
	URL string
	URLs []string
}

type FinishedTaskReply struct{}





