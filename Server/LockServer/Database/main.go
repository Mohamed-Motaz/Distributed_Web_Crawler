package main

import (
	dbConfig "Server/LockServer/Database/Configurations"

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
