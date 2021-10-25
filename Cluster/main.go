package main

import (
	"fmt"

	master "github.com/mohamed247/Distributed_Web_Crawler/Cluster/Master"
)


func main(){
	fmt.Printf("Building cluster")

	master.MakeMaster()

	fmt.Print()
}