package main

import (
	"fmt"
	"strings"
)

type TreeVertex struct {
	Token    HTMLToken
	Children []*TreeVertex
	Text     strings.Builder
	Parent   *TreeVertex
}

type TreeRoot struct {
	Root *TreeVertex
}

func createRoot(rootToken HTMLToken) TreeRoot {
	return TreeRoot{&TreeVertex{rootToken, nil, strings.Builder{}, nil}}
}

func findComplementaryOpenTag(node *TreeVertex, name string) *TreeVertex {
	prevNode := node
	for prevNode.Token.Name != name {
		prevNode = prevNode.Parent
	}
	return prevNode.Parent
}

func buildParseTree(tokens []HTMLToken, printParser bool) (*TreeRoot, error) {
	root := createRoot(tokens[0])

	currentNode := root.Root
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case StartTag, DOCTYPE:
			child := TreeVertex{token, nil, strings.Builder{}, currentNode}
			currentNode.Children = append(currentNode.Children, &child)
			if !child.Token.SelfClosingFlag {
				currentNode = &child
			} //otherwise stays the same
		case EndTag:
			currentNode = findComplementaryOpenTag(currentNode, token.Name)
		case CommentType:
			//ignore
		case Character:
			currentNode.Text.WriteString(token.Content.String())
		case EOF:
			return &root, nil
		}
	}

	if printParser {
		printTree(*root.Root, 0)
	}
	return &root, nil
}

func printTree(root TreeVertex, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Printf("  ")
	}
	fmt.Printf("Name: %s, Text: %s", root.Token.Name, &root.Text)
	fmt.Println()

	for _, child := range root.Children {
		printTree(*child, depth+1)
	}
}
