package main

import "fmt"

//ISSUES:
//- leveling

func main() {
	body, err := httpClient("http://0.0.0.0:8080/")
	if err != nil {
		panic(err)
	}

	tokenizer := NewHTMLTokenizer(body)
	tokenOutput := tokenizer.TokenizeHTML()
	for _, token := range tokenOutput {
		fmt.Println(token)
	}

}
