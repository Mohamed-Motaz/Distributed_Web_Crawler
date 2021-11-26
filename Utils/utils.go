package utils

import "net/http"

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