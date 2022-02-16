package database

import (
	logger "Server/Cluster/Logger"
	dbConfig "Server/LockServer/Database/Configurations"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	user = dbConfig.User
	password = dbConfig.Password
	dbname = dbConfig.Dbname
	DB_PORT = "DB_PORT"
)


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

//return a thread-safe *gorm.DB that can safely be used 
//by multiple goroutines
func New(myHost, myPort string) *DBWrapper{
	db := connect(myHost, myPort)
	setUp(db)
	logger.LogInfo(logger.DATABASE, "Db setup complete")
	return &DBWrapper{
		db: db,
	}
	
}
func (db *DBWrapper) Close(){
}

func connect(myHost, myPort string) *gorm.DB{

	dsn := fmt.Sprintf(
		"user=%v password=%v host=%v port=%v database=%v sslmode=disable",
		user, password, myHost, myPort, dbname)

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

//BASE METHODS ------------------------------------------
func (db *DBWrapper) getRecord(info *Info, query string, value interface{}){
	db.db.Where(query, value).First(info)
}

func (db *DBWrapper) getRecords(infos *[]Info, query string, value interface{}){
	db.db.Where(query, value).Find(infos)
}

func (db *DBWrapper) getRecordsOrderBy(infos *[]Info, query string, value interface{}, orderCol string, orderType string){
	db.db.Where(query, value).Order(orderCol + " " + orderType).Find(infos)
}

func (db *DBWrapper) getAllRecords(infos *[]Info){
	db.db.Find(infos)
}

func (db *DBWrapper) addRecord(info *Info){
	db.db.Create(info)
}

func (db *DBWrapper) updateRecord(info *Info ){
	db.db.Save(info)
}

//permenant deletion
func (db *DBWrapper) deleteRecord(info *Info){
	db.db.Unscoped().Delete(info)
}

//permenant deletion, susceptible to sql injections
func (db *DBWrapper) deleteAllRecords(table string){
	db.db.Unscoped().Exec("Delete from " + table)
}

// func (db *DBWrapper) UpdateRecord(info *Info, data map[string]interface{} ){
// 	db.db.Model(&info).Updates(data)
// }
















func ManualTesting(dBWrapper *DBWrapper) {
	info := &Info{
		JobId: "JobId",
		MasterId: "MasterId",
		UrlToCrawl: "UrlToCrawl",
		DepthToCrawl: 1,
		TimeAssigned: 100,
	}
	dBWrapper.addRecord(info)

	info = &Info{}
	dBWrapper.getRecord(info, JOB_ID + " = ?", "JobId")
	logger.LogInfo(logger.DATABASE, "The info retreived %+v", info)

	infos := []Info{}
	dBWrapper.getRecords(&infos, URL_TO_CRAWL + " = ?", "UrlToCrawl")
	logger.LogInfo(logger.DATABASE, "The infos retreived %+v", infos)

	infos = []Info{}
	dBWrapper.getAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived %+v", infos)

	info.JobId = "NNEWWWW"
	dBWrapper.updateRecord(info)

	infos = []Info{}
	dBWrapper.getAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived after update %+v", infos)

	infos = []Info{}
	dBWrapper.GetRecordsThatPassedXSeconds(&infos, 20)
	logger.LogInfo(logger.DATABASE, "All infos that passedXseconds %+v", infos)

	//dBWrapper.deleteRecord(info)

	infos = []Info{}
	dBWrapper.getAllRecords(&infos)
	logger.LogInfo(logger.DATABASE, "All infos retreived after deleting record %+v", infos)

	//dBWrapper.deleteAllRecords(TABLE_NAME)

		
	//database.ManualTesting(l.dbWrapper);
	info = &Info{}
	dBWrapper.GetRecordByJobId(info, "NNEWWWW")
	fmt.Printf("%+v\n\n", info)

	info.JobId = "Hello"
	dBWrapper.UpdateRecord(info)
	dBWrapper.GetRecordByJobId(info, "Hello")
	fmt.Printf("%+v\n\n", info)

	info = &Info{}
	dBWrapper.GetRecordByJobId(info, "NNEWWWW")
	info.JobId = "Bro"
	dBWrapper.UpdateRecord(info)

	info = &Info{}
	dBWrapper.GetRecordByJobId(info, "Bro")
	fmt.Printf("%+v\n\n", info)

	dBWrapper.DeleteRecord(info)

	info = &Info{}
	dBWrapper.GetRecordByJobId(info, "Bro")
	fmt.Printf("%+v\n\n", info)	


	infos = []Info{}
	dBWrapper.GetRecordsThatPassedXSeconds(&infos, 20)
	fmt.Printf("%+v\n\n", infos)

}