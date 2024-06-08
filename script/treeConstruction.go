package main

import (
	"errors"
	"strings"
)

const htmlTag string = "html"

type TreeVertex struct {
	Token    HTMLToken
	Children []HTMLToken
	Text     strings.Builder
	Parent   *TreeVertex
}

type TreeRoot struct {
	Root TreeVertex
}

func findRoot(tokens []HTMLToken) (TreeRoot, int, error) {
	var root TreeRoot
	for idx, token := range tokens {
		if token.Name == htmlTag {
			root.Root.Token = token
			return root, idx, nil
		}
	}
	return root, 0, errors.New("no html tag in document")
}

func findComplementaryOpenTag(node *TreeVertex) *TreeVertex {
	prevNode := node.Parent
	for prevNode.Token.Name != node.Token.Name {
		prevNode = prevNode.Parent
	}
	return prevNode
}

func buildParseTree(tokens []HTMLToken) (TreeRoot, error) {
	root, rootIdx, err := findRoot(tokens)
	if err != nil {
		return root, err
	}

	currentNode := root.Root
	for i := rootIdx + 1; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case StartTag:
			child := TreeVertex{token, nil, strings.Builder{}, &currentNode}
			currentNode.Children = append(currentNode.Children, child.Token)
			currentNode = child
		case EndTag:
			currentNode = *findComplementaryOpenTag(&currentNode)
		case CommentType:
			//ignore
		case Character:
			currentNode.Text.WriteString(token.Content)
		case EOF:
			return root, nil
		}
	}

	return root, nil
}
