package main

import (
	"strings"
)

//const htmlTag string = "html"

type TreeVertex struct {
	Token    HTMLToken
	Children []*TreeVertex
	Text     strings.Builder
	Parent   *TreeVertex
}

type TreeRoot struct {
	Root TreeVertex
}

func createRoot(rootToken HTMLToken) TreeRoot {
	return TreeRoot{TreeVertex{rootToken, nil, strings.Builder{}, nil}}
}

func findComplementaryOpenTag(node *TreeVertex, name string) *TreeVertex {
	prevNode := node
	for prevNode.Token.Name != name {
		prevNode = prevNode.Parent
	}
	return prevNode
}

func buildParseTree(tokens []HTMLToken) (TreeRoot, error) {
	root := createRoot(tokens[0]) //findRoot(tokens)

	currentNode := &root.Root
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch token.Type {
		case StartTag, DOCTYPE:
			child := TreeVertex{token, nil, strings.Builder{}, currentNode}
			currentNode.Children = append(currentNode.Children, &child)
			currentNode = &child
		case EndTag:
			currentNode = findComplementaryOpenTag(currentNode, token.Name)
		case CommentType:
			//ignore
		case Character:
			currentNode.Text.WriteString(token.Content.String())
		case EOF:
			return root, nil
		}
	}

	return root, nil
}
