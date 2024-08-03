package main

import (
	"strings"
	"unicode"
)

const (
	ampersand          = 0x0026
	space              = 0x0020
	FF                 = 0x000C
	LF                 = 0x000A
	tab                = 0x0009
	greaterThan        = 0x003E
	lesserThan         = 0x003C
	equal              = 0x003D
	null               = 0x0000
	quoteMark          = 0x0022
	apostrophe         = 0x0027
	graveAccent        = 0x0060
	replacementChar    = "\uFFFD"
	exclamationMark    = 0x0021
	solidus            = 0x002F
	questionMark       = 0x003F
	dash               = 0x002D
	rightSquareBracket = 0x005D
	leftSquareBracket  = 0x005B
	numberSign         = 0x0023
)

func isASCIIAlpha(s byte) bool {
	return isUppercase(s) || isLowercase(s)
}

/*
func isASCIINumeric(s byte) bool {
	return (s >= 0x030) && (s <= 0x39)
}

func isASCIIAlphanumeric(s byte) bool {
	return isASCIIAlpha(s) || isASCIINumeric(s)
}
*/

func isUppercase(s byte) bool {
	if s >= 65 && s <= 90 {
		return true
	}
	return false
}

func isLowercase(s byte) bool {
	if s >= 97 && s <= 122 {
		return true
	}
	return false
}

func Lower(bytes []byte) string {
	var sb strings.Builder
	for i := 0; i < len(bytes); i++ {
		sb.WriteString(string(bytes[i]))
	}
	return strings.ToLower(sb.String())
}

func emitToken(tokenToEmit HTMLToken, tokens *[]HTMLToken) {
	*tokens = append(*tokens, tokenToEmit)
}

func emitCurrToken(currToken *HTMLToken, tokens *[]HTMLToken) {
	emitToken(*currToken, tokens)
	*currToken = HTMLToken{}
}

func NewHTMLTokenizer(input []byte) *HTMLTokenizer {
	var states []State
	return &HTMLTokenizer{
		input:            input,
		curr:             0,
		state:            Data,
		returnState:      states,
		tmpBuffer:        strings.Builder{},
		lastStartTagName: "",
		currToken:        HTMLToken{},
	}
}

func reconsume(state *State, switchTo State, pointer *int) {
	*state = switchTo
	*pointer--
}

func nameInDoctype(token *HTMLTokenizer, pointer int, id bool) bool {
	publicLowercase := [6]byte{
		0x70,
		0x75,
		0x62,
		0x6C,
		0x69,
		0x63,
	}

	systemLowercase := [6]byte{
		0x73,
		0x79,
		0x73,
		0x74,
		0x65,
		0x6D,
	}

	var word []byte
	if id {
		word = publicLowercase[:]
	} else {
		word = systemLowercase[:]
	}

	for i := 0; i < 6; i++ {
		if token.input[pointer+i] != word[i] || token.input[pointer+i] != word[i]-0x20 {
			return false
		}
	}
	return true
}

/*
	func popState(returnState *[]State) State {
		popped := (*returnState)[len(*returnState)-1]
		*returnState = (*returnState)[:len(*returnState)-1]
		return popped
	}

	func flushCodePoints(currToken *HTMLToken, returnState State, tmpBuffer string, tokens *[]HTMLToken) {
		asAnAttribute := []State{
			AttributeValueDoubleQuoted,
			AttributeValueSingleQuoted,
			AttributeValueUnquoted,
		}
		if slices.Contains(asAnAttribute, returnState) {
			currToken.Attributes[len(currToken.Attributes)-1].Value = tmpBuffer
		} else {
			var content strings.Builder
			content.WriteString(tmpBuffer)
			emitToken(HTMLToken{
				Type:    Character,
				Content: content,
			}, tokens)
		}
	}
*/
func isNotWhitespace(sb string) bool {
	for i := 0; i < len(sb); i++ {
		if !unicode.IsSpace(rune(sb[i])) {
			return true
		}
	}
	return false
}

func toBuilder(char string) strings.Builder {
	var sb strings.Builder
	sb.WriteString(char)
	return sb
}

func createSelfClosingTrie() TrieRoot {
	selfClosingTrieRoot := TrieVertex{Value: "", Children: []*TrieVertex{
		{Value: "r", Children: []*TrieVertex{
			{Value: "h", Children: []*TrieVertex{
				{Value: ""},
			}},
			{Value: "b", Children: []*TrieVertex{
				{Value: ""},
			}},
		}}}}
	return TrieRoot{Root: &selfClosingTrieRoot}
}

func selfClosingWithoutSolidus(tokenizer *HTMLTokenizer) bool {
	curr := selfClosingTrieRoot
	for _, child := range curr.Root.Children {
		match := traverseTrie(child, tokenizer, 1)
		if match {
			return match
		}
	}
	return false
}

func traverseTrie(vertex *TrieVertex, tokenizer *HTMLTokenizer, level int) bool {
	if vertex.Value == "" {
		return true
	}
	if vertex.Value != string(tokenizer.input[tokenizer.curr-level]) {
		return false
	} else {
		for _, child := range vertex.Children {
			match := traverseTrie(child, tokenizer, level+1)
			if match {
				return true
			}
		}
		return false
	}
}

type TrieRoot struct {
	Root *TrieVertex
}

type TrieVertex struct {
	Value    string
	Children []*TrieVertex
}
