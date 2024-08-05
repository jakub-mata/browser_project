package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func httpClient(url string) ([]byte, error) {

	client := http.Client{Timeout: time.Duration(3) * time.Second} //add cookies
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error during http request")
		var empty []byte
		return empty, err
	}
	defer resp.Body.Close()

	bodyBytes, err := checkEncoding(resp)

	if err != nil {
		return bodyBytes, err
	}
	return bodyBytes, nil
}

func checkEncoding(resp *http.Response) ([]byte, error) {
	contentType := resp.Header.Get("Content-Type")
	fmt.Println("Content-type:", contentType)

	body := (resp.Body).(io.Reader)
	if strings.Contains(contentType, "charset=ISO-8859-1") {
		body = transform.NewReader(body, charmap.ISO8859_1.NewDecoder())
	}

	bodyBytes, err := io.ReadAll(body)
	return bodyBytes, err
}
