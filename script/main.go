package main

import "os"

var baseAddress string

func main() {
	logTokens, logParser, website := startCLI()
	baseAddress = website
	body, err := httpClient(website) //run python3 http.server 8080
	if err != nil {
		panic(err)
	}
	//tokenizer
	tokenizer := NewHTMLTokenizer(body)
	tokenOutput := tokenizer.TokenizeHTML(logTokens)
	//parser
	root, err := buildParseTree(tokenOutput, logParser)
	if err != nil {
		panic(err)
	}
	//viewer
	os.Setenv("FYNE_THEME", "light")
	CreateViewer(root)
}
