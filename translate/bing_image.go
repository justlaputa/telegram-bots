package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ImageApiEndpoint = "https://api.cognitive.microsoft.com/bing/v7.0/images/search"
	ImageCount       = 10
	JapanMarketCode  = "ja-jp"
)

type BingImageSearch struct {
	subscriptionKey string
}

type ImageSearchResponse struct {
	Value []ImageValue `json:"value"`
}

type ImageValue struct {
	Name           string  `json:"name"`
	ThumbnailURL   string  `json:"thumbnailUrl"`
	ContentURL     string  `json:"contentUrl"`
	ContentSize    string  `json:"contentSize"`
	EncodingFormat string  `json:"encodingFormat"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
}

func (bing BingImageSearch) getOneImage(query string) (string, error) {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		log.Printf("query string is empty")
		return "", errors.New("query string can not be empty")
	}
	client := &http.Client{Timeout: time.Second * 10}

	request, err := http.NewRequest("GET", ImageApiEndpoint, nil)
	if err != nil {
		log.Printf("failed to create new http request: %v", err)
		return "", err
	}

	request.Header.Add("Ocp-Apim-Subscription-Key", bing.subscriptionKey)
	q := request.URL.Query()
	q.Add("q", query)
	q.Add("mkt", JapanMarketCode)
	q.Add("count", strconv.Itoa(ImageCount))
	request.URL.RawQuery = q.Encode()

	resp, err := client.Do(request)
	if err != nil {
		log.Printf("failed to call bing image search api, %v", err)
		return "", err
	}

	defer resp.Body.Close()
	searchResponse := &ImageSearchResponse{}

	log.Printf("got response from bing image serach: %v", resp)

	err = json.NewDecoder(resp.Body).Decode(searchResponse)
	if err != nil {
		log.Printf("could not parse response body to json, %v", err)
		return "", err
	}

	if len(searchResponse.Value) == 0 {
		log.Printf("could not get any image from the query: %s", query)
		return "", errors.New("found 0 image")
	}

	n := len(searchResponse.Value)
	i := rand.Intn(n)
	return searchResponse.Value[i].ContentURL, nil
}

func NewBingImageSearchProvider(subscriptionKey string) *BingImageSearch {
	if len(strings.TrimSpace(subscriptionKey)) == 0 {
		log.Fatal("could not create bing image search from empty subscription key")
	}

	return &BingImageSearch{subscriptionKey}
}
