package database

import "time"

const DESC = "DESC";
const ASC = "ASC";

const TABLE_NAME = "infos"
const ID = "id"
const JOB_ID = "job_id"
const MASTER_ID = "master_id"
const URL_TO_CRAWL = "url_to_crawl"
const DEPTH_TO_CRAWL = "depth_to_crawl"
const TIME_ASSIGNED = "time_assigned"


//all the methods available to the user of this package

func (db *DBWrapper) GetRecordByJobId(info *Info, jobId string){
	db.getRecord(info, JOB_ID + " = ?", jobId)
}

func (db *DBWrapper) GetRecordsThatPassedXSeconds(infos *[]Info, seconds int){
	now := time.Now().UnixMilli() / 1000
	start := now - int64(seconds)

	db.getRecordsOrderBy(infos, TIME_ASSIGNED + " <= ?", start, TIME_ASSIGNED, ASC)
}

func (db *DBWrapper) AddRecord(info *Info){
	db.addRecord(info)
}

func (db *DBWrapper) UpdateRecord(info *Info){
	db.updateRecord(info)
}

func (db *DBWrapper) DeleteRecord(info *Info){
	db.deleteRecord(info)
}

func (db *DBWrapper) DeleteAllRecords(){
	db.deleteAllRecords(TABLE_NAME)
}

