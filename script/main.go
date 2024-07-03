package main

var printTokens bool = false
var printParser bool = false

func main() {
	body, err := httpClient("http://0.0.0.0:8080/")
	if err != nil {
		panic(err)
	}

	tokenizer := NewHTMLTokenizer(body)
	tokenOutput := tokenizer.TokenizeHTML(printTokens)

	root, err := buildParseTree(tokenOutput, printParser)
	if err != nil {
		panic(err)
	}

	CreateViewer(root)
}
