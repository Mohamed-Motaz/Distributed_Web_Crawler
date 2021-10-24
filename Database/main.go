package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	dbConfig "github.com/mohamed247/Distributed_Web_Crawler/Configurations"
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

	err = setUpDbDebug(dbPool);

	if err != nil{
		fmt.Printf("Unable to set up the db with err %v \n", err)
	}else{
		fmt.Printf("Successfully set up the db \n")
	}
	
}


func setUpDb(dbPool *pgxpool.Pool) error{
	_, err := dbPool.Exec(context.Background(),
	`DROP DATABASE IF EXISTS "DistributedWebCrawler";`)

	if err != nil{
		return err;
	}

	fmt.Println("Successfully dropped DistributedWebCrawler")

	_, err = dbPool.Exec(context.Background(),
				fmt.Sprintf(
				`CREATE DATABASE "DistributedWebCrawler"
				WITH 
				OWNER = %v
				ENCODING = 'UTF8'
				LC_COLLATE = 'C'
				LC_CTYPE = 'C'
				TABLESPACE = pg_default
				CONNECTION LIMIT = -1;`, user),
			)
				

	return err;

}

func setUpDbDebug(dbPool *pgxpool.Pool) error{
	
	_, err := dbPool.Exec(context.Background(),
				fmt.Sprintf(
				`CREATE DATABASE DistributedWebCrawler
				WITH 
				OWNER = %v
				ENCODING = 'UTF8'
				LC_COLLATE = 'C'
				LC_CTYPE = 'C'
				TABLESPACE = pg_default
				CONNECTION LIMIT = -1;`, user),
			)
				

	return err;

}