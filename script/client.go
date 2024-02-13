package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func httpClient(url string) io.ReadCloser {

	client := http.Client{Timeout: time.Duration(1) * time.Second} //add cookies
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error during http request")
		panic(err)
	}
	defer resp.Body.Close()

	return resp.Body

	//fmt.Printf("Body: %s", body)

}
