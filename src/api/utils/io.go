package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func GetRequestReturnStuff[T any](url string) T {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)

	response, responseErr := client.Do(request)
	if responseErr != nil {
		log.Fatal(responseErr)
	}
	defer response.Body.Close()

	var objects T
	if jsonErr := json.NewDecoder(response.Body).Decode(&objects); jsonErr != nil {
		log.Fatal(responseErr)
	}

	return objects
}
