package RPC

//
// RPC definitions
//

type TaskArgs struct{
	MasterId string   //make sure sending to the correct master
	WokerId string   //worker current id for master to keep track
	URL string   //url to crawl

}

type TaskReply struct{
	validCommand bool //worker sending request belongs to master's cluster
	URL string
	MasterId string
	URLs []string
}



