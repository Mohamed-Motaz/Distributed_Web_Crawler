package main

import (
	logger "Server/Cluster/Logger"
	dbConfig "Server/LockServer/Database/Configurations"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host = dbConfig.Host
	port = dbConfig.Port
	user = dbConfig.User
	password = dbConfig.Password
	dbname = dbConfig.Dbname
)

const TABLE_NAME = "infos"
const ID = "id"
const JOB_ID = "job_id"
const MASTER_ID = "master_id"
const URL_TO_CRAWL = "url_to_crawl"
const DEPTH_TO_CRAWL = "depth_to_crawl"
const TIME_ASSIGNED = "time_assigned"

type Info struct{
	Id int									`gorm:"primaryKey"` 	
	JobId	string    						//id of job				  
	MasterId string							//id of master
	UrlToCrawl string						//urlToCrawl
	DepthToCrawl int  						//depth required
	TimeAssigned int64						//epoch time in seconds	
}

type DBWrapper struct{
	db *gorm.DB
}

func main(){
	logger.LogInfo(logger.DATABASE, "Starting setup of db")
	dBWrapper := New()

	//manualTesting(dBWrapper);
	
	dBWrapper.Close()
}

//return a thread-safe *gorm.DB that can safely be used 
//by multiple goroutines
func New() *DBWrapper{
	db := connect()
	setUp(db)
	logger.LogInfo(logger.DATABASE, "Db setup complete")
	return &DBWrapper{
		db: db,
	}
	
}

func (db *DBWrapper) Close(){
}

func (db *DBWrapper) GetRecord(info *Info, query string, value interface{}){
	db.db.Where(query, value).First(info)
}

func (db *DBWrapper) GetRecords(infos *[]Info, query string, value interface{}){
	db.db.Where(query, value).Find(infos)
}

func (db *DBWrapper) GetRecordsThatPassedXSeconds(infos *[]Info, seconds int){
	now := time.Now().UnixMilli() / 1000
	start := now - int64(seconds)

	db.GetRecords(infos, TIME_ASSIGNED + " <= ?", start)
}

func (db *DBWrapper) GetAllRecords(infos *[]Info){
	db.db.Find(infos)
}

func (db *DBWrapper) AddRecord(info *Info){
	db.db.Create(info)
}

func (db *DBWrapper) UpdateRecord(info *Info ){
	db.db.Save(info)
}

//permenant deletion
func (db *DBWrapper) DeleteRecord(info *Info){
	db.db.Unscoped().Delete(info)
}

//permenant deletion, susceptible to sql injections
func (db *DBWrapper) DeleteAllRecords(table string){
	db.db.Unscoped().Exec("Delete from " + table)
}

// func (db *DBWrapper) UpdateRecord(info *Info, data map[string]interface{} ){
// 	db.db.Model(&info).Updates(data)
// }

func connect() *gorm.DB{

	dsn := fmt.Sprintf(
		"user=%v password=%v host=%v port=%v database=%v sslmode=disable",
		user, password, host, port, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil{
		logger.FailOnError(logger.DATABASE, "Unable to connect to db with this error %v", err)
	}
	return db
}

func setUp(db *gorm.DB) {
	err := db.AutoMigrate(&Info{})
	if err != nil{
		logger.FailOnError(logger.DATABASE, "Unable to migrate the tables with this error %v", err)
	}
}

func manualTesting(dBWrapper *DBWrapper) {
	info := &Info{
		JobId: "JobId",
		MasterId: "MasterId",
		UrlToCrawl: "UrlToCrawl",
		DepthToCrawl: 1,
		TimeAssigned: 10,
	}
	dBWrapper.AddRecord(info)

	info = &Info{}
	dBWrapper.GetRecord(info, JOB_ID + " = ?", "JobId")
	logger.LogInfo(logger.DATABASE, "The info retreived %+v", info)

	infos := []Info{}
	dBWrapper.GetRecords(&infos, URL_TO_CRAWL + " = ?", "UrlToCrawl")
	logger.LogInfo(logger.DATABASE, "The infos retreived %+v", infos)

	infos = []Info{}
	dBWrapper.GetAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived %+v", infos)

	info.JobId = "NNEWWWW"
	dBWrapper.UpdateRecord(info)

	infos = []Info{}
	dBWrapper.GetAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived after update %+v", infos)

	infos = []Info{}
	dBWrapper.GetRecordsThatPassedXSeconds(&infos, 20)
	logger.LogInfo(logger.DATABASE, "All infos that passedXseconds %+v", infos)

	dBWrapper.DeleteRecord(info)

	infos = []Info{}
	dBWrapper.GetAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived after deleting record %+v", infos)

	dBWrapper.DeleteAllRecords(TABLE_NAME)
}