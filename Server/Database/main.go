package main

import (
	logger "Server/Cluster/Logger"
	dbConfig "Server/Database/Configurations"
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	host = dbConfig.Host
	port = dbConfig.Port
	user = dbConfig.User
	password = dbConfig.Password
	dbname = dbConfig.Dbname
)


func main(){
	fmt.Println("Starting setup of postgres db")

	connectionString := fmt.Sprintf(
		"user=%v password=%v host=%v port=%v pool_max_conns=10",
		user, password, host, port)

	dbPool, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}else{
		fmt.Println("Successfully established connection with postgres")
	}
	defer dbPool.Close()

	err = dropDb(dbPool, dbname)
	err = setUpDb(dbPool, dbname)

	if err != nil{
		fmt.Printf("Unable to set up the db with err %v \n", err)
	}else{
		fmt.Printf("Successfully set up the db \n")
	}
	
}


func setUpDb(dbPool *pgxpool.Pool, dbName string) error{
	_, err := dbPool.Exec(context.Background(),
				fmt.Sprintf(
				`CREATE DATABASE "%v"
				WITH 
				OWNER = %v
				ENCODING = 'UTF8'
				TABLESPACE = pg_default
				LC_COLLATE = 'en_US.utf8'
				LC_CTYPE = 'en_US.utf8'
				CONNECTION LIMIT = -1;`, dbName, user),
			)
					
	if err != nil{
		logger.LogError(logger.DATABASE, "Error while creating the db %v %v", dbName, err)
	}else{
		logger.LogInfo(logger.DATABASE, "Successfully created the db %v", dbName)
	}

	return err;
}

func dropDb(dbPool *pgxpool.Pool, dbName string) error{
	_, err := dbPool.Exec(context.Background(),
	fmt.Sprintf(`DROP DATABASE IF EXISTS "%v";`, dbName))

	if err != nil{
		logger.LogError(logger.DATABASE, "Error while dropping the db %v %v", dbName, err)
	}else{
		logger.LogInfo(logger.DATABASE, "Successfully dropped the db %v", dbName)
	}

	return err
}
