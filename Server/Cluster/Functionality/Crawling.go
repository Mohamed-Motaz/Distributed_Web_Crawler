package Crawling

import (
	logger "Distributed_Web_Crawler/Logger"
	utils "Distributed_Web_Crawler/Utils"
	"net/http"

	"golang.org/x/net/html"
)

const mxTokensToParse int = 1000


func GetURLsSlice(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil{
		logger.LogError(logger.CRAWLING, logger.ESSENTIAL, "Error while getting the url %v", err)
		return nil, err
	}

	linksMap := make(map[string]int)

	tokens := html.NewTokenizer(resp.Body)
	ctr := 0

	for ctr < mxTokensToParse{
		tokenType := tokens.Next()
		
		switch tokenType{
		case html.ErrorToken:
			return utils.ConvertMapToList(linksMap), nil

		case html.StartTagToken, html.EndTagToken:
			token := tokens.Token()
			if token.Data == "a"{
				for _, attr := range token.Attr{
					if attr.Key == "href"{
						if utils.LinkIsValid(attr.Val) {
							linksMap[attr.Val] = 1;
						}
					}
				}
			}
		}
		ctr++
	}
	return []string{}, nil
}


func GetURLsMap(url string) (map[string]bool, error) {

	resp, err := http.Get(url)
	if err != nil{
		return nil, err
	}

	linksMap := make(map[string]bool)

	tokens := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokens.Next()
		
		switch tokenType{
		case html.ErrorToken:
			return linksMap, nil

		case html.StartTagToken, html.EndTagToken:
			token := tokens.Token()
			if token.Data == "a"{
				for _, attr := range token.Attr{
					if attr.Key == "href"{
						if utils.LinkIsValid(attr.Val) {
							linksMap[attr.Val] = true;
						}
					}
				}
			}
		}
	}

}

