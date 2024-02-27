package main

import (
	"fmt"
	"strings"
)

type HTMLTokenizer struct {
	input string
	curr  int
}

const (
	ampersand       = 0x0026
	space           = 0x0020
	FF              = 0x000C
	LF              = 0x000A
	tab             = 0x0009
	greaterThan     = 0x003E
	lesserThan      = 0x003C
	equal           = 0x003D
	null            = 0x0000
	quoteMark       = 0x0022
	apostrophe      = 0x0027
	graveAccent     = 0x0060
	replacementChar = "\uFFFD"
	exclamationMark = 0x0021
	solidus         = 0x002F
	questionMark    = 0x003F
	dash            = 0x002D
)

func emitToken(tokenToEmit HTMLToken, tokens *[]HTMLToken) {
	*tokens = append(*tokens, tokenToEmit)
}

func NewHTMLTokenizer(input string) *HTMLTokenizer {
	return &HTMLTokenizer{input: input, curr: 0}
}

func TokenizeHTML(token *HTMLTokenizer) []HTMLToken {

	//var space
	var tokens []HTMLToken
	var state State = Data
	var currToken HTMLToken
	var endOfFile byte = byte(0)
	var tmpBuffer string = ""
	var lastStartTagName string = ""
	var returnState []int //a stack of return states, values are ints from enums State

	for token.curr < len(token.input) {
		currVal := token.input[token.curr]
		switch state {
		case Data:
			switch currVal {
			case ampersand:
				state = CharacterReference
				returnState = append(returnState, int(Data))
			case lesserThan: //'<'
				state = TagOpen
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case RCDATA:
			switch currVal {
			case ampersand: //'&'
				state = CharacterReference
				returnState = append(returnState, int(RCDATA))
			case lesserThan: //'<'
				state = RCDATALessThanSign
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar, // REPLACEMENT CHARACTER
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case RAWTEXT:
			switch currVal {
			case lesserThan: //'<'
				state = RAWTEXTLessThanSign
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar, // REPLACEMENT CHARACTER
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case Script:
			switch currVal {
			case lesserThan: //'<'
				state = ScriptDataLessThanSign
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar, // REPLACEMENT CHARACTER
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case PLAINTEXT:
			switch currVal {
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar, // REPLACEMENT CHARACTER
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case TagOpen:
			switch currVal {
			case exclamationMark: //'!'
				state = MarkupDeclarationOpen

			case solidus: //'/'
				state = EndTagOpen

			case questionMark: //'?'
				// Parse Error
				currToken.Type = CommentType
				currToken.Content = ""
				//Move to bogus comment
				state = BogusComment
				token.curr--
			case endOfFile:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)

			default:
				if isASCII(currVal) {
					// create a new start tag token
					currToken.Type = StartTag
					currToken.Name = ""
					currToken.Attributes = []Attribute{}
					currToken.SelfClosingFlag = false

					state = TagName
				} else {
					// Parse Error
					emitToken(HTMLToken{
						Type:    Character,
						Content: "<", // LESS-THAN SIGN
					}, &tokens)
					//Reconsume in data state
					state = Data
				}
				token.curr--
			}

		case EndTagOpen:
			switch currVal {
			case greaterThan: //'>'
				// Parse Error
				state = Data
			case endOfFile:
				//Parse Error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // SOLIDUS
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				if isASCII(currVal) {
					// create a new end tag token
					currToken.Type = EndTag
					currToken.Name = ""
					currToken.Attributes = []Attribute{}
					currToken.SelfClosingFlag = false
					state = TagName
					token.curr-- //reconsume in tag name state
				} else {
					// Parse Error
					currToken.Type = CommentType
					currToken.Content = ""
					//Reconsume in data state
					state = Data
					token.curr--
				}
			}

		case TagName:
			switch currVal {
			case tab: //'\t'
			case LF: //'\n'
			case FF: //'\f'
			case space: //' '
				state = BeforeAttributeName
			case greaterThan: //'>'
				state = Data
				emitToken(currToken, &tokens)
				if currToken.Type == StartTag {
					lastStartTagName = currToken.Name
				}
				currToken = HTMLToken{}
			case solidus: //'/'
				state = SelfClosingStartTag
			case null: //null
				// Parse error
				currToken.Name = currToken.Name + replacementChar // REPLACEMENT CHARACTER
			case endOfFile:
				// emit current tag token
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)

			default:
				if isUppercase(currVal) {
					currToken.Name += string(currVal + 0x20)
				} else {
					currToken.Name += string(currVal)
				}
			}

		case RCDATALessThanSign:
			switch currVal {
			case solidus: //SOLIDUS
				state = RCDATAEndTagOpen
				tmpBuffer = ""
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				//Reconsume in RCDATA state
				state = RCDATA
				token.curr--
			}

		case RCDATAEndTagOpen:
			// All reconsumed
			if isASCII(currVal) {
				currToken.Type = EndTag
				currToken.Name = ""
				state = RCDATAEndTagName
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // SOLIDUS
				}, &tokens)
			}
			token.curr--

		case RCDATAEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ignore whitespace
			} else if currVal == space && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currVal == greaterThan && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currVal == solidus && lastStartTagName == currToken.Name {
				state = SelfClosingStartTag
			} else if isUppercase(currVal) {
				currToken.Name += currToken.Name + string(currVal+0x20)
				tmpBuffer += string(currVal + 0x20)
			} else if isLowercase(currVal) {
				//reconsume in RCDATA state
				tmpBuffer += string(currVal)
				currToken.Name = currToken.Name + string(currVal)
			} else {
				//emit current tag token
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // Solidus
				}, &tokens)
				// emit a token for each char in tmpBuffer
				for _, char := range tmpBuffer {
					emitToken(HTMLToken{
						Type:    Character,
						Content: string(char),
					}, &tokens)
				}
				state = RCDATA
				token.curr--
			}

		case RAWTEXTLessThanSign:
			switch currVal {
			case solidus: //SOLIDUS
				state = RAWTEXTEndTagOpen
				tmpBuffer = ""
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				//Reconsume in RAWTEXT state
				state = RAWTEXT
				token.curr--
			}

		case RAWTEXTEndTagOpen:
			// All reconsumed
			if isASCII(currVal) {
				currToken.Type = EndTag
				currToken.Name = ""
				state = RAWTEXTEndTagName
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // SOLIDUS
				}, &tokens)
				//Reconsume in the RAWTEXT state
				state = RAWTEXT
			}
			token.curr--

		case RAWTEXTEndTagName:

			if currVal == tab || currVal == LF || currVal == FF { //tab, LF, FF
				//ignore whitespace
			} else if currVal == space && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currVal == greaterThan && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currVal == solidus && lastStartTagName == currToken.Name {
				state = SelfClosingStartTag
			} else if isUppercase(currVal) {
				currToken.Name += currToken.Name + string(currVal+0x20)
				tmpBuffer += string(currVal + 0x20)
			} else if isLowercase(currVal) {
				currToken.Name += currToken.Name + string(currVal)
				tmpBuffer += string(currVal)
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // SOLIDUS
				}, &tokens)
				// emit a token for each char in tmpBuffer
				for _, char := range tmpBuffer {
					emitToken(HTMLToken{
						Type:    Character,
						Content: string(char),
					}, &tokens)
				}
				// Reconsume in the RAWTEXT state
				state = RAWTEXT
				token.curr--
			}

		case ScriptDataLessThanSign:
			// All reconsumed

			if isUppercase(currVal) {
				currToken.Type = EndTag
				currToken.Name = ""
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // SOLIDUS
				}, &tokens)
				//Reconsume in the Script data state
				state = Script
			}
			token.curr--

		case ScriptDataEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ingore whitespace
			} else if currVal == space && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currVal == greaterThan && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currVal) {
				currToken.Name = currToken.Name + string(currVal+space)
				tmpBuffer = tmpBuffer + string(currVal)
			} else if isLowercase(currVal) {
				currToken.Name = currToken.Name + string(currVal)
				tmpBuffer = tmpBuffer + string(currVal)
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/",
				}, &tokens)
				for _, char := range tmpBuffer {
					emitToken(HTMLToken{
						Type:    Character,
						Content: string(char),
					}, &tokens)
				}
				state = Script
				token.curr--
			}

		case ScriptDataEscapeStart:
			switch currVal {
			case dash: //'-'
				state = ScriptDataEscapeStartDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			default:
				//Reconsume in the script data state
				state = Script
				token.curr--
			}

		case ScriptDataEscapeStartDash:
			switch currVal {
			case dash: //'-'
				state = ScriptDataEscapeStartDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			default:
				//Reconsume in the script data state
				state = Script
				token.curr--
			}

		case ScriptDataEscapedDashDash:
			switch currVal {
			case dash: //'-'
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			case lesserThan: //"<"
				state = ScriptDataEscapedLessThanSign
			case greaterThan: //">"
				state = Script
				emitToken(HTMLToken{
					Type:    Character,
					Content: ">",
				}, &tokens)
			case null:
				//Parse error
				state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar,
				}, &tokens)
			case endOfFile:
				// Parse error
				state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case ScriptDataEscapedLessThanSign:

			if currVal == solidus { //Solidus
				tmpBuffer = ""
				state = ScriptDataEscapedEndTagOpen
			} else if isASCII(currVal) {
				tmpBuffer = ""
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // '<'
				}, &tokens)
				state = ScriptDataEscaped
				token.curr--
			}

		case ScriptDataEscapedEndTagOpen:
			// All reconsumed
			if isASCII(currVal) {
				currToken.Type = EndTag
				currToken.Name = ""
				state = ScriptDataEscapedEndTagName
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // '<'
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/", // '/'
				}, &tokens)
				state = ScriptDataEscaped
			}
			token.curr--

		case ScriptDataEscapedEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ignore whitespace
			} else if currVal == space && currToken.Name == lastStartTagName {
				state = BeforeAttributeName
			} else if currVal == solidus && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currVal == greaterThan && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currVal) {
				currToken.Name += string(currVal + space)
				tmpBuffer += string(currVal)
			} else if isLowercase(currVal) {
				currToken.Name += string(currVal)
				tmpBuffer += string(currVal)
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/",
				}, &tokens)
				for _, val := range tmpBuffer {
					emitToken(HTMLToken{
						Type:    Character,
						Content: string(val),
					}, &tokens)
				}
				//Reconsume
				state = ScriptDataEscaped
				token.curr--
			}

		case ScriptDataDoubleEscapeStart:

			if currVal == tab || currVal == LF ||
				currVal == FF || currVal == space || currVal == solidus {
				//ignore whitespace
			} else if currVal == greaterThan {
				if tmpBuffer == "script" {
					state = ScriptDataDoubleEscaped
				} else {
					state = ScriptDataEscaped
				}
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)

			} else if isUppercase(currVal) {
				tmpBuffer += string(currVal + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			} else if isLowercase(currVal) {
				tmpBuffer += string(currVal + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			} else {
				//Reconsume
				state = ScriptDataEscaped
				token.curr--
			}

		case ScriptDataDoubleEscaped:
			switch currVal {
			case dash: // '-'
				state = ScriptDataDoubleEscapedDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			case lesserThan: //<
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar,
				}, &tokens)
			case endOfFile:
				// Parse error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			}

		case ScriptDataDoubleEscapedDash:
			switch currVal {
			case dash: // '-'
				state = ScriptDataDoubleEscapedDashDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			case lesserThan: // <
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
			case null: // null
				//Parse error
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar,
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case ScriptDataDoubleEscapedDashDash:
			switch currVal {
			case dash: // '-'
				emitToken(HTMLToken{
					Type:    Character,
					Content: "-",
				}, &tokens)
			case lesserThan: //'<'
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
			case greaterThan: //'>'
				state = Script
				emitToken(HTMLToken{
					Type:    Character,
					Content: ">",
				}, &tokens)
			case null: //null
				//Parse error
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: replacementChar,
				}, &tokens)
			case endOfFile:
				//Parse error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			}

		case ScriptDataDoubleEscapedLessThanSign:
			if currVal == solidus {
				tmpBuffer = ""
				state = ScriptDataDoubleEscapeEnd
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/",
				}, &tokens)
			} else {
				state = ScriptDataDoubleEscaped
				token.curr--
			}

		case ScriptDataDoubleEscapeEnd:

			if currVal == tab || currVal == LF || currVal == FF ||
				currVal == space || currVal == solidus {
				//ignore whitespace
			} else if currVal == greaterThan {
				if tmpBuffer == "script" {
					state = ScriptDataEscaped
				} else {
					state = ScriptDataDoubleEscaped
				}
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			} else if isUppercase(currVal) {
				tmpBuffer += string(currVal + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			} else if isLowercase(currVal) {
				tmpBuffer += string(currVal)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currVal),
				}, &tokens)
			} else {
				//Reconsume
				state = ScriptDataDoubleEscaped
				token.curr--
			}

		case BeforeAttributeName:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case greaterThan, solidus: // '>' or '/'
			case endOfFile: //EOF
				state = AfterAttributeName
				token.curr--
			case equal: // '='
				//Parse error
				currToken.Attributes = append(currToken.Attributes, Attribute{
					Name:  string(currVal),
					Value: "",
				})
				state = AttributeName
			default:
				currToken.Attributes = append(currToken.Attributes, Attribute{
					Name:  "",
					Value: "",
				})
				state = AttributeName
				token.curr--
			}

		case AttributeName:
			switch currVal {
			case tab, LF, FF, space, solidus, greaterThan: //whitespace, '/' or '>'
			case endOfFile:
				state = AfterAttributeName
				token.curr--
			case equal:
				state = BeforeAttributeValue
			case null:
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + replacementChar
			default:
				if isUppercase(currVal) {
					idx := len(currToken.Attributes) - 1
					currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + string(currVal+space)
				} else {
					idx := len(currToken.Attributes) - 1
					currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + string(currVal)
				}
			}

			//Checking duplicates
			if state != AttributeName {
				// Can be faster by counting occurences in a map, relying on not so many attributes
				for i := 0; i < len(currToken.Attributes); i++ {
					for j := i + 1; j < len(currToken.Attributes); j++ {
						if currToken.Attributes[i].Name == currToken.Attributes[j].Name {
							//Parse error, remove duplicate
							currToken.Attributes = append(currToken.Attributes[:j], currToken.Attributes[j+1:]...)
						}
					}
				}
			}

		case AfterAttributeName:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case solidus: // '/'
				state = SelfClosingStartTag
			case equal: // '='
				state = BeforeAttributeValue
			case greaterThan: // '>'
				state = Data
				emitToken(currToken, &tokens)
			case endOfFile:
				//Parse Error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.Attributes = append(currToken.Attributes, Attribute{
					Name:  "",
					Value: "",
				})
				state = AttributeName
				token.curr--
			}

		case BeforeAttributeValue:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case quoteMark: // '"'
				state = AttributeValueDoubleQuoted
			case apostrophe: // '\''
				state = AttributeValueSingleQuoted
			case greaterThan: //">"
				//Parse error, missing attribute
				state = Data
				emitToken(currToken, &tokens)
			default:
				state = AttributeValueUnquoted
				token.curr--
			}

		case AttributeValueDoubleQuoted:
			switch currVal {
			case quoteMark: //'"'
				state = AfterAttributeValueQuoted
			case ampersand: //'&'
				returnState = append(returnState, int(AttributeValueDoubleQuoted))
				state = CharacterReference
			case null:
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + string(currVal)
			}

		case AttributeValueSingleQuoted:
			switch currVal {
			case apostrophe: //'
				state = AfterAttributeValueQuoted
			case ampersand: //	"&"
				returnState = append(returnState, int(AttributeValueSingleQuoted))
				state = CharacterReference
			case null: //null
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + string(currVal)
			}

		case AttributeValueUnquoted:
			switch currVal {
			case tab, LF, FF, space: //whitespace
				state = BeforeAttributeName
			case ampersand:
				returnState = append(returnState, int(AttributeValueUnquoted))
				state = CharacterReference
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
			case null:
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + replacementChar
			}

		case AfterAttributeValueQuoted:
			switch currVal {
			case tab, LF, FF, space:
				state = BeforeAttributeName
			case solidus:
				state = SelfClosingStartTag
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//Reconsume
				state = BeforeAttributeName
				token.curr--
			}

		case SelfClosingStartTag:
			switch currVal {
			case greaterThan:
				currToken.SelfClosingFlag = true
				state = Data
				emitToken(currToken, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//Reconsume
				state = BeforeAttributeName
				token.curr--
			}

		case BogusComment:
			switch currVal {
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
			case endOfFile:
				emitToken(currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			case null:
				currToken.Content = currToken.Content + replacementChar
			default:
				currToken.Content = currToken.Content + string(currVal)
			}

		case MarkupDeclarationOpen:
			if currVal == dash && token.input[token.curr+1] == dash {
				currToken = HTMLToken{
					Type:    CommentType,
					Content: "",
				}
				state = CommentStart
				token.curr++
			} else if strings.ToLower(token.input[token.curr:token.curr+7]) == "doctype" {
				state = Doctype
				token.curr += 6
			} else if strings.ToLower(token.input[token.curr:token.curr+8]) == "[cdata[" {
				token.curr += 7
				//TO DO: implement CDATA section
				state = CDATASection
			} else {
				currToken = HTMLToken{
					Type:    CommentType,
					Content: "",
				}
				state = BogusComment
				token.curr--
			}

		case CommentStart:
			switch currVal {
			case dash:
				state = CommentStartDash
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			default:
				state = Comment
				token.curr--
			}

		case CommentStartDash:
			switch currVal {
			case dash:
				state = CommentEnd
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				emitToken(currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.Content = currToken.Content + string(rune(dash))
				state = Comment
				token.curr--
			}

		case CommentLessThanSign:
			switch currVal {
			case exclamationMark:
				state = CommentLessThanSignBang
				currToken.Content = currToken.Content + string(rune(exclamationMark))
			case lesserThan:
				currToken.Content = currToken.Content + string(rune(exclamationMark))
			default:
				state = Comment
				token.curr--
			}

		case CommentLessThanSignBang:
			if currVal == dash {
				state = CommentLessThanSignBangDash
			} else {
				state = Comment
				token.curr--
			}

		case CommentLessThanSignBangDash:
			if currVal == dash {
				state = CommentLessThanSignBangDashDash
			} else {
				state = CommentEndDash
				token.curr--
			}

		default:
			fmt.Println("Not here yet")
		}
		//consume the current current character (token.curr-- in individual cases otherwise)
		token.curr++

	}
	return tokens
}
