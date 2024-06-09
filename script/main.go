package main

import (
	"fmt"
	"strings"
)

func main() {
	body, err := httpClient("http://0.0.0.0:8080/")
	if err != nil {
		panic(err)
	}

	testToken := NewHTMLTokenizer(body)
	tokens := TokenizeHTML(testToken)
	/*
		for i := 0; i < len(tokens); i++ {
			fmt.Println(tokens[i])
		}
	*/
	root, err := buildParseTree(tokens)
	if err != nil {
		panic(err)
	}
	printNode(&root.Root, 0)
}

func printNode(node *TreeVertex, indentLevel int) {
	if node == nil {
		return
	}

	// Print the tag name and attributes
	fmt.Printf("%s%s%s\n", strings.Repeat("  ", indentLevel), node.Token.Name, node.Token.Attributes)

	// Print text content
	if node.Text.String() != "" {
		fmt.Printf("%s#text: \"%s\"\n", strings.Repeat("  ", indentLevel+1), node.Text.String())
	}

	// Recursively print children
	for _, child := range node.Children {
		printNode(child, indentLevel+1)
	}
}
