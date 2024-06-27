package main

import (
	"fmt"
	"strings"
)

type TokenType uint8

const (
	DOCTYPE TokenType = iota
	StartTag
	EndTag
	CommentType
	Character
	EOF
)

type HTMLToken struct {
	Type            TokenType
	Name            string //doctype and tags
	PublicID        string //doctype
	SystemID        string //doctype
	ForceQuirksFlag bool   //doctype
	SelfClosingFlag bool   //start and end tags
	Attributes      []Attribute
	Content         strings.Builder //comments and characters
}

func (t HTMLToken) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Type: %v, Name: %s, Content: %s, ",
		getTokenType(t.Type), t.Name, t.Content.String()))

	for i, attr := range t.Attributes {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("Attribute[%d]: Name=%s, Value=%s", i+1, attr.Name, attr.Value))
	}

	return sb.String()
}

func getTokenType(t TokenType) string {
	switch t {
	case DOCTYPE:
		return "DOCTYPE"
	case StartTag:
		return "StartTag"
	case EndTag:
		return "EndTag"
	case CommentType:
		return "Comment"
	case Character:
		return "Character"
	case EOF:
		return "EOF"
	}
	return "Error"
}

type Attribute struct {
	Name  string
	Value string
}

type State uint8

const (
	Data State = iota
	RCDATA
	RAWTEXT
	Script
	PLAINTEXT
	TagOpen
	EndTagOpen
	TagName
	RCDATALessThanSign
	RCDATAEndTagOpen
	RCDATAEndTagName
	RAWTEXTLessThanSign
	RAWTEXTEndTagOpen
	RAWTEXTEndTagName
	ScriptDataLessThanSign
	ScriptDataEndTagOpen
	ScriptDataEndTagName
	ScriptDataEscapeStart
	ScriptDataEscapeStartDash
	ScriptDataEscaped
	ScriptDataEscapedDash
	ScriptDataEscapedDashDash
	ScriptDataEscapedLessThanSign
	ScriptDataEscapedEndTagOpen
	ScriptDataEscapedEndTagName
	ScriptDataDoubleEscapeStart
	ScriptDataDoubleEscaped
	ScriptDataDoubleEscapedDash
	ScriptDataDoubleEscapedDashDash
	ScriptDataDoubleEscapedLessThanSign
	ScriptDataDoubleEscapeEnd
	BeforeAttributeName
	AttributeName
	AfterAttributeName
	BeforeAttributeValue
	AttributeValueDoubleQuoted
	AttributeValueSingleQuoted
	AttributeValueUnquoted
	AfterAttributeValueQuoted
	SelfClosingStartTag
	BogusComment
	MarkupDeclarationOpen
	CommentStart
	CommentStartDash
	Comment
	CommentLessThanSign
	CommentLessThanSignBang
	CommentLessThanSignBangDash
	CommentLessThanSignBangDashDash
	CommentEndDash
	CommentEnd
	CommentEndBang
	Doctype
	BeforeDOCTYPEName
	DOCTYPEName
	AfterDOCTYPEName
	AfterDOCTYPEPublicKeyword
	BeforeDOCTYPEPublicIdentifier
	DOCTYPEPublicIdentifierDoubleQuoted
	DOCTYPEPublicIdentifierSingleQuoted
	AfterDOCTYPEPublicIdentifier
	BetweenDOCTYPEPublicAndSystemIdentifiers
	AfterDOCTYPESystemKeyword
	BeforeDOCTYPESystemIdentifier
	DOCTYPESystemIdentifierDoubleQuoted
	DOCTYPESystemIdentifierSingleQuoted
	AfterDOCTYPESystemIdentifier
	BogusDOCTYPE
	CDATASection
	CDATASectionBracket
	CDATASectionEnd
	CharacterReference
	NamedCharacterReference
	AmbiguousAmpersand
	NumericCharacterReference
	HexadecimalCharacterReferenceStart
	DecimalCharacterReferenceStart
	HexadecimalCharacterReference
	DecimalCharacterReference
	NumericCharacterReferenceEnd
)
