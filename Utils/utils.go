package Utils

import (
	"net/http"
	"os"
)

func LinkIsValid(link string) bool {
	_, err := http.Get(link)
	return err == nil
}

func ConvertMapToList(linksMap map[string]int) []string{
	var links []string

	for k := range linksMap {
		links = append(links, k)
	}

	return links
}

func ConvertMapArrayToList(linksMap []map[string]int) [][]string{
	var result [][]string = make([][]string, 0)

	//for each depth
	for _, mp := range linksMap {
		//for each element in said depth
		
		result = append(result, ConvertMapToList(mp))
	}

	return result
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

func GetEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}
