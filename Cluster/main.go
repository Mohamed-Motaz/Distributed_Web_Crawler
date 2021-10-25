package main

import (
	"fmt"
	"sync"

	master "github.com/mohamed247/Distributed_Web_Crawler/Cluster/Master"
)


func main(){
	fmt.Printf("Setting up master")

	master.MakeMaster()

	wg := sync.WaitGroup{}
	wg.Wait() 
}