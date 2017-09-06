package image

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

//Google Custom Search Engine Image search

const (
	// GCSEEndpoint google custom search engine api endpoint
	GCSEEndpoint = "https://www.googleapis.com/customsearch/v1"
)

// GoogleCustomSearch google cse
type GoogleCustomSearch struct {
	EngineID string
	APIKey   string
}

// NewGoogleCustomSearch create new search provider
func NewGoogleCustomSearch(cx, key string) SearchProvider {
	return &GoogleCustomSearch{cx, key}
}

type searchResult struct {
	Items []gcseItem `json:"items"`
}

type gcseItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

//Search search google custom engine
func (g *GoogleCustomSearch) Search(query string) ([]Image, error) {
	images := []Image{}

	client := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("GET", GCSEEndpoint, nil)
	if err != nil {
		log.Printf("failed to cretae request object, %v", err)
		return images, err
	}

	q := req.URL.Query()
	q.Add("key", g.APIKey)
	q.Add("cx", g.EngineID)
	q.Add("searchType", "image")
	q.Add("q", query)

	req.URL.RawQuery = q.Encode()

	log.Printf("sending request: %+v", req)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to call pinterest pin search api, %v", err)
		return images, err
	}

	defer resp.Body.Close()
	searchResponse := &searchResult{}

	log.Printf("got response from google serach: %+v", resp)

	err = json.NewDecoder(resp.Body).Decode(searchResponse)
	if err != nil {
		log.Printf("could not parse response body to json, %v", err)
		return images, err
	}

	if len(searchResponse.Items) == 0 {
		log.Printf("could not get any pin from the query: %s", query)
		return images, errors.New("found 0 image")
	}

	log.Printf("google cse returns %d image results", len(searchResponse.Items))

	for _, item := range searchResponse.Items {
		images = append(images, Image{item.Title, item.Link})
	}

	return images, nil
}
