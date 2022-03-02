package MessageQueue

//queue names
const JOBS_ASSIGNED_QUEUE = "jobsAssigned"
const DONE_JOBS_QUEUE = "doneJobs"


//objects passed into and out of messageQ

type Job struct{
	ClientId string			`json:"clientId"`
	JobId string			`json:"jobId"`
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
}

type DoneJob struct{
	ClientId string			`json:"clientId"`
	JobId string			`json:"jobId"`
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
	Results [][]string		`json:"results"`
}
