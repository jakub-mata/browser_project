package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func httpClient(url string) ([]byte, error) {

	client := http.Client{Timeout: time.Duration(1) * time.Second} //add cookies
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error during http request")
		var empty []byte
		return empty, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return bodyBytes, err
	}
	return bodyBytes, nil
}
