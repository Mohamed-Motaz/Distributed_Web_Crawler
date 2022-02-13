package RPC

//
// RPC definitions
//

//for master-worker communication
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




//for master-lockserver communication
type GetJobArgs struct{
	Id string
	URL string   //url to crawl
	JobNum int
}

type GetJobReply struct{
	Accepted bool      //whether the lock server accepted or rejected the master's job request
}

type FinishedJobArgs struct{
	Id int
	JobNum int
	URL string
}

type FinishedJobReply struct{}



