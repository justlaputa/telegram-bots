package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type InterestsResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    DataType `json:"data"`
}

type DataType struct {
	Images      ImageType       `json:"images"`
	Attribution AttributionType `json:"attribution"`
}

type ImageType struct {
	Orig OrigType `json:"orig"`
}

type OrigType struct {
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

type AttributionType struct {
	Title string `json:"title"`
}

func PintrestInterests(subject string) (InterestsResponse, error) {
	if subject == "" {
		return InterestsResponse{}, errors.New("no subject")
	}

	resp, err := http.Get(fmt.Sprintf("https://api.pinterest.com/uno/interests/%s", subject))

	if err != nil {
		log.Printf("failed to get pinterest pictures: %v", err)
		return InterestsResponse{}, err
	}

	interestsResponse := InterestsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&interestsResponse)
	if err != nil {
		log.Printf("failed to parse response json")
		return InterestsResponse{}, err
	}

	log.Printf("got response from pinterest: %+v", interestsResponse)

	return interestsResponse, nil
}
