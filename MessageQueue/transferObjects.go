package MessageQueue

//queue names
const JOBS_QUEUE = "jobs"
const DONE_JOBS_QUEUE = "doneJobs"


//objects passed into and out of messageQ

type Job struct{
	JobId string			`json:"jobId"`
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
}

type DoneJob struct{
	JobId string			`json:"jobId"`
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
	Results [][]string		`json:"results"`
}
