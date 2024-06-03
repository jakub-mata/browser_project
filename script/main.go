package main

import (
	"fmt"
)

func main() {
	body, err := httpClient("http://0.0.0.0:8080/")
	if err != nil {
		panic(err)
	}

	//htmlTokens := html.NewTokenizer(resp)

	testToken := NewHTMLTokenizer(body)
	tokens := TokenizeHTML(testToken)
	for i := 0; i < len(tokens); i++ {
		fmt.Println(tokens[i])
	}
}
