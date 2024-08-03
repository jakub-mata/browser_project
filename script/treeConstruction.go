package main

import (
	"fmt"
)

type TreeVertex struct {
	Token    HTMLToken
	Children []*TreeVertex
	Parent   *TreeVertex
}

type TreeRoot struct {
	Root *TreeVertex
}

func createRoot(rootToken HTMLToken) TreeRoot {
	return TreeRoot{&TreeVertex{rootToken, nil, nil}}
}

func findComplementaryOpenTag(node *TreeVertex, name string) (*TreeVertex, error) {
	prevNode := node
	for prevNode.Token.Name != name {
		if prevNode == nil {
			return nil, fmt.Errorf("no complimentary tag found")
		}
		prevNode = prevNode.Parent
	}
	return prevNode.Parent, nil
}

func buildParseTree(tokens []HTMLToken, printParser bool) (*TreeRoot, error) {
	root := createRoot(tokens[0])

	currentNode := root.Root
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case StartTag, DOCTYPE:
			child := TreeVertex{token, nil, currentNode}
			currentNode.Children = append(currentNode.Children, &child)
			if !child.Token.SelfClosingFlag {
				currentNode = &child
			} //otherwise stays the same
		case EndTag:
			compNode, err := findComplementaryOpenTag(currentNode, token.Name)
			if err != nil {
				return nil, err
			}
			currentNode = compNode
		case CommentType:
			//ignore
		case Character:
			child := TreeVertex{token, nil, currentNode}
			currentNode.Children = append(currentNode.Children, &child)
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
	if root.Token.Type == Character {
		fmt.Printf("Name: %s, Text: %s", root.Token.Name, &root.Token.Content)
	} else {
		fmt.Printf("Name: %s", root.Token.Name)
	}
	fmt.Println()

	for _, child := range root.Children {
		printTree(*child, depth+1)
	}
}
