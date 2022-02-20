package RPC

//
// RPC definitions
//

//for master-worker communication -------------------------
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




//for master-lockserver communication -----------------------------
type GetJobArgs struct{
	MasterId string
	ClientId string
	JobId string
	URL string   //url to crawl
	Depth int
	NoCurrentJob bool  //for when a master wants an outstanding job from the lockserver if any are available
}

type GetJobReply struct{
	Accepted bool      //whether the lock server accepted or rejected the master's job request	MasterId string
	AlternateJob bool  //whether there is an alternate job with higher priority
	ClientId string
	JobId string       //details of alternate job
	URL string   
	Depth int
}

type FinishedJobArgs struct{
	MasterId string
	ClientId string
	JobId string
	URL string
}

type FinishedJobReply struct{}



