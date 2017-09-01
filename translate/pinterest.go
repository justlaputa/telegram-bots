package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	// PinterestAPIEndpoint is the api endpoint
	PinterestAPIEndpoint = "https://api.pinterest.com/v1"
)

// Pinterest pinterest api object
type Pinterest struct {
	APIToken string
}

type searchResult struct {
	Data []pinData `json:"data"`
}

type pinData struct {
	ID    string    `json:"id"`
	Image imageData `json:"image"`
}

type imageData struct {
	Original struct {
		URL    string      `json:"url"`
		Width  json.Number `json:"width"`
		Height json.Number `json:"height"`
	} `json:"original"`
}

// Search search pinned image of the user account
func (p *Pinterest) Search(query string) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("GET", p.userPinSearchAPI(), nil)
	if err != nil {
		log.Printf("failed to cretae pinterest search request object, %v", err)
		return "", err
	}

	q := req.URL.Query()
	q.Add("access_token", p.APIToken)
	q.Add("query", query)
	q.Add("fields", "id,image")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to call pinterest pin search api, %v", err)
		return "", err
	}

	defer resp.Body.Close()
	searchResponse := &searchResult{}

	log.Printf("got response from pin serach: %v", resp)

	err = json.NewDecoder(resp.Body).Decode(searchResponse)
	if err != nil {
		log.Printf("could not parse response body to json, %v", err)
		return "", err
	}

	if len(searchResponse.Data) == 0 {
		log.Printf("could not get any pin from the query: %s", query)
		return "", errors.New("found 0 image")
	}

	n := len(searchResponse.Data)
	i := rand.Intn(n)

	return searchResponse.Data[i].Image.Original.URL, nil
}

func (p *Pinterest) userPinSearchAPI() string {
	return fmt.Sprintf("%s/me/search/pins/", PinterestAPIEndpoint)
}
