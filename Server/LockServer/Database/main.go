package main

import (
	logger "Server/Cluster/Logger"
	dbConfig "Server/LockServer/Database/Configurations"
	"fmt"

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

type Info struct{
	JobId	int    `gorm:"primaryKey"` 		//primary key					  
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
	db.db.First(info, query, value)
}

func (db *DBWrapper) GetRecords(infos *[]Info, query string, value interface{}){
	db.db.Where(query, value).Find(infos)
}

func (db *DBWrapper) GetAllRecords(infos *[]Info){
	db.db.Find(infos)
}

func (db *DBWrapper) AddRecord(info *Info){
	db.db.Create(info)
}

func (db *DBWrapper) UpdateRecord(info *Info, data map[string]interface{} ){
	db.db.Model(&info).Updates(data)
}

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