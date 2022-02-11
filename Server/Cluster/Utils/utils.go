package utils

import (
	"net/http"
	"os"
)

func LinkIsValid(link string) bool {
	_, err := http.Get(link)
	return err == nil
}

func ConvertMapToList(linksMap map[string]bool) []string{
	var links []string

	for k := range linksMap {
		links = append(links, k)
	}

	return links
}

func ConvertMapArrayToList(linksMap []map[string]int) []string{
	//create a set to remove any duplicates
	var set map[string]bool = make(map[string]bool)

	//for each depth
	for _, mp := range linksMap {
		//for each element in said depth
		for k := range mp{
			set[k] = true 
		}
	}

	return ConvertMapToList(set)
}

//
// resize by reference
//
func ResizeSlice(lst []string, requiredLen int){
	if len(lst) < requiredLen{
		return
	}
	lst = lst[:requiredLen]
	return
}

func failOnError(){
	os.Exit(1)
}