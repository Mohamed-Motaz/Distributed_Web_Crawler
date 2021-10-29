package crawling

import (
	"net/http"

	logger "github.com/mohamed247/Distributed_Web_Crawler/Logger"
	"golang.org/x/net/html"
)


func GetURLsSlice(url string) ([]string, error) {

	resp, err := http.Get(url)
	if err != nil{
		logger.LogError(logger.CRAWLING, "Error while getting the url %v", err)
		return nil, err
	}

	linksMap := make(map[string]bool)

	tokens := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokens.Next()
		
		switch tokenType{
		case html.ErrorToken:
			return convertMapToList(linksMap), nil

		case html.StartTagToken, html.EndTagToken:
			token := tokens.Token()
			if token.Data == "a"{
				for _, attr := range token.Attr{
					if attr.Key == "href"{
						if linkIsValid(attr.Val) {
							linksMap[attr.Val] = true;
						}
					}
				}
			}
		}
	}

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
						if linkIsValid(attr.Val) {
							linksMap[attr.Val] = true;
						}
					}
				}
			}
		}
	}

}

func linkIsValid(link string) bool {
	_, err := http.Get(link)
	return err == nil
}

func convertMapToList(linksMap map[string]bool) []string{
	var links []string

	for k := range linksMap {
		links = append(links, k)
	}

	return links
}