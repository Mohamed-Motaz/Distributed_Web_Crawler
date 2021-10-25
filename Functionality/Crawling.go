package crawling

import (
	"net/http"

	"golang.org/x/net/html"
)

// func main(){
// 	testUrl := "https://www.google.com/";

// 	links, err := GetURLs(testUrl)
// 	fmt.Printf("links %+v %+v \n", links, err)
// }


func GetURLs(url string) ([]string, error) {

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

func linkIsValid(link string) bool {
	_, err := http.Get(link)
	if err != nil{
		return false
	}
	return true
}

func convertMapToList(linksMap map[string]bool) []string{
	var links []string

	for k := range linksMap {
		links = append(links, k)
	}

	return links
}