package main

func isASCII(s byte) bool {
	return s <= 127
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
	Content         string //comments and characters
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
