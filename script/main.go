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
	/*for _, token := range tokenOutput {
		fmt.Println(token)
	}*/
	root, err := buildParseTree(tokenOutput)
	if err != nil {
		panic(err)
	}
	printTree(root.Root, 0)
}

func printTree(root TreeVertex, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("Name: %s, Text: %s", root.Token.Name, &root.Text)
	fmt.Println()

	for _, child := range root.Children {
		printTree(*child, depth+1)
	}
}
