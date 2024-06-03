package main

import (
	"slices"
	"strings"
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

func isNewline(check byte) bool {
	return (check == 0x0A) || (check == 0x0D)
}

func isASCIIAlpha(s byte) bool {
	return isUppercase(s) || isLowercase(s)
}

func isASCIINumeric(s byte) bool {
	return (s >= 0x030) && (s <= 0x39)
}

func isASCIIAlphanumeric(s byte) bool {
	return isASCIIAlpha(s) || isASCIINumeric(s)
}

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
	return &HTMLTokenizer{input: input, curr: 0}
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
		emitToken(HTMLToken{
			Type:    Character,
			Content: tmpBuffer,
		}, tokens)
	}
}
