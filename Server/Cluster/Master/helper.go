package main

import (
	logger "server/cluster/logger"
	RPC "server/cluster/rpc"
)
const 
(
	TaskAvailable = iota
	TaskAssigned
	TaskDone
)

type Job struct{
	urlToCrawl string		`json:"urlToCrawl"`
	depthToCrawl int  		`json:"depthToCrawl"`
}

type FinishedJob struct{
	urlToCrawl string		`json:"urlToCrawl"`
	depthToCrawl int  		`json:"depthToCrawl"`
	results [][]string		`json:"results"`
}






//
// RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, "A worker requested to be given a task %v", args)
	reply.JobNum = -1
	reply.URL = ""

	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.currentJob{
		logger.LogInfo(logger.MASTER, 
			"A worker requested to be given a task but we have no jobs at the moment")
		return nil
	} 

	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogDelay(logger.MASTER, 
			"A worker requested to be given a task but we have already finished the job")
		return nil
	} 

	master.checkJobAvailable(reply)

	return nil
}

func (master *Master) HandleFinishedTasks(args *RPC.FinishedTaskArgs, reply *RPC.FinishedTaskReply) error {
	master.mu.Lock()
	defer master.mu.Unlock()

	if master.currentDepth >= master.jobRequiredDepth{
		logger.LogDelay(logger.MASTER, 
			"A worker has finished a task %v but we have already finished the job", args.URL)
		return nil
	} 

	if !master.currentJob {
		logger.LogDelay(logger.MASTER, 
			"A worker finished a late task %v for job num %v but there is no current job",
			args.URL, args.JobNum)
		return nil
	}

	if (master.URLsTasks[master.currentDepth][args.URL] == TaskDone){
		//already finished, do nothing
		logger.LogDelay(logger.MASTER, "Worker has finished this task with jobNum %v and URL %v " +
							"which was already finished", args.JobNum, args.URL)
		return nil
	}	

	logger.LogTaskDone(logger.MASTER, "A worker just finished this task: \n" +
		"JobNum: %v \nURL: %v \nURLsLen :%v", 
		args.JobNum, args.URL, len(args.URLs))

	master.URLsTasks[master.currentDepth][args.URL] = TaskDone //set the task as done


	if master.currentDepth + 1 < master.jobRequiredDepth{
		//add all urls to the URLsTasks next depth and set their tasks as available
		for _, v := range args.URLs{
			if (!master.urlInTasks(v, args.URL)){
				master.URLsTasks[master.currentDepth + 1][v] = TaskAvailable
			}	
		}
	}
	
	return nil
}

