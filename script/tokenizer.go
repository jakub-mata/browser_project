package main

import "fmt"

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
		switch state {
		case Data:
			switch token.input[token.curr] {
			case ampersand:
				state = CharacterReference
				returnState = append(returnState, int(Data))
			case lesserThan: //'<'
				state = TagOpen
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(token.input[token.curr]),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case RCDATA:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case RAWTEXT:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case Script:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case PLAINTEXT:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case TagOpen:
			switch token.input[token.curr] {
			case exclamationMark: //'!'
				state = MarkupDeclarationOpen
				token.curr++
			case solidus: //'/'
				state = EndTagOpen
				token.curr++
			case questionMark: //'?'
				// Parse Error
				currToken.Type = CommentType
				currToken.Content = ""
				//Move to bogus comment
				state = BogusComment
			case endOfFile:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
				token.curr++
			default:
				if isASCII(token.input[token.curr]) {
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
			}
		case EndTagOpen:
			switch token.input[token.curr] {
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
				if isASCII(token.input[token.curr]) {
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
			token.curr++

		case TagName:
			switch token.input[token.curr] {
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
				if isUppercase(token.input[token.curr]) {
					currToken.Name += string(token.input[token.curr] + 0x20)
				} else {
					currToken.Name += string(token.input[token.curr])
				}
			}
			token.curr++

		case RCDATALessThanSign:
			switch token.input[token.curr] {
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
			token.curr++

		case RCDATAEndTagOpen:
			// All reconsumed
			if isASCII(token.input[token.curr]) {
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

		case RCDATAEndTagName:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF || currChar == FF {
				//ignore whitespace
			} else if currChar == space && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currChar == greaterThan && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currChar == solidus && lastStartTagName == currToken.Name {
				state = SelfClosingStartTag
			} else if isUppercase(currChar) {
				currToken.Name += currToken.Name + string(currChar+0x20)
				tmpBuffer += string(currChar + 0x20)
			} else if isLowercase(currChar) {
				//reconsume in RCDATA state
				tmpBuffer += string(currChar)
				currToken.Name = currToken.Name + string(currChar)
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
			token.curr++

		case RAWTEXTLessThanSign:
			switch token.input[token.curr] {
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
			token.curr++

		case RAWTEXTEndTagOpen:
			// All reconsumed
			if isASCII(token.input[token.curr]) {
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

		case RAWTEXTEndTagName:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF || currChar == FF { //tab, LF, FF
				//ignore whitespace
			} else if currChar == space && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currChar == greaterThan && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currChar == solidus && lastStartTagName == currToken.Name {
				state = SelfClosingStartTag
			} else if isUppercase(currChar) {
				currToken.Name += currToken.Name + string(currChar+0x20)
				tmpBuffer += string(currChar + 0x20)
			} else if isLowercase(currChar) {
				currToken.Name += currToken.Name + string(currChar)
				tmpBuffer += string(currChar)
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
			token.curr++

		case ScriptDataLessThanSign:
			// All reconsumed
			currChar := token.input[token.curr]
			if isUppercase(currChar) {
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

		case ScriptDataEndTagName:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF || currChar == FF {
				//ingore whitespace
			} else if currChar == space && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currChar == greaterThan && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currChar) {
				currToken.Name = currToken.Name + string(currChar+space)
				tmpBuffer = tmpBuffer + string(currChar)
			} else if isLowercase(currChar) {
				currToken.Name = currToken.Name + string(currChar)
				tmpBuffer = tmpBuffer + string(currChar)
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
			token.curr++

		case ScriptDataEscapeStart:
			switch token.input[token.curr] {
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
			token.curr++

		case ScriptDataEscapeStartDash:
			switch token.input[token.curr] {
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
			token.curr++

		case ScriptDataEscapedDashDash:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case ScriptDataEscapedLessThanSign:
			currChar := token.input[token.curr]
			if currChar == solidus { //Solidus
				tmpBuffer = ""
				state = ScriptDataEscapedEndTagOpen
			} else if isASCII(currChar) {
				tmpBuffer = ""
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // '<'
				}, &tokens)
				state = ScriptDataEscaped
				token.curr--
			}
			token.curr++

		case ScriptDataEscapedEndTagOpen:
			// All reconsumed
			if isASCII(token.input[token.curr]) {
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

		case ScriptDataEscapedEndTagName:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF || currChar == FF {
				//ignore whitespace
			} else if currChar == space && currToken.Name == lastStartTagName {
				state = BeforeAttributeName
			} else if currChar == solidus && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currChar == greaterThan && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currChar) {
				currToken.Name += string(currChar + space)
				tmpBuffer += string(currChar)
			} else if isLowercase(currChar) {
				currToken.Name += string(currChar)
				tmpBuffer += string(currChar)
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
			token.curr++

		case ScriptDataDoubleEscapeStart:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF ||
				currChar == FF || currChar == space || currChar == solidus {
				//ignore whitespace
			} else if currChar == greaterThan {
				if tmpBuffer == "script" {
					state = ScriptDataDoubleEscaped
				} else {
					state = ScriptDataEscaped
				}
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)

			} else if isUppercase(currChar) {
				tmpBuffer += string(currChar + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else if isLowercase(currChar) {
				tmpBuffer += string(currChar + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else {
				//Reconsume
				state = ScriptDataEscaped
				token.curr--
			}
			token.curr++

		case ScriptDataDoubleEscaped:
			switch token.input[token.curr] {
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
			token.curr++

		case ScriptDataDoubleEscapedDash:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case ScriptDataDoubleEscapedDashDash:
			switch token.input[token.curr] {
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
					Content: string(token.input[token.curr]),
				}, &tokens)
			}
			token.curr++

		case ScriptDataDoubleEscapedLessThanSign:
			if token.input[token.curr] == solidus {
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
			token.curr++

		case ScriptDataDoubleEscapeEnd:
			currChar := token.input[token.curr]
			if currChar == tab || currChar == LF || currChar == FF ||
				currChar == space || currChar == solidus {
				//ignore whitespace
			} else if currChar == greaterThan {
				if tmpBuffer == "script" {
					state = ScriptDataEscaped
				} else {
					state = ScriptDataDoubleEscaped
				}
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else if isUppercase(currChar) {
				tmpBuffer += string(currChar + space)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else if isLowercase(currChar) {
				tmpBuffer += string(currChar)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else {
				//Reconsume
				state = ScriptDataDoubleEscaped
				token.curr--
			}
			token.curr++

		case BeforeAttributeName:
			switch token.input[token.curr] {
			case tab, LF, FF, space: //whitespace
			case greaterThan, solidus: // '>' or '/'
			case endOfFile: //EOF
				state = AfterAttributeName
				token.curr--
			case equal: // '='
				//Parse error
				currToken.Attributes = append(currToken.Attributes, Attribute{
					Name:  string(token.input[token.curr]),
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
			token.curr++

		case AttributeName:
			switch token.input[token.curr] {
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
				if isUppercase(token.input[token.curr]) {
					idx := len(currToken.Attributes) - 1
					currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + string(token.input[token.curr]+space)
				} else {
					idx := len(currToken.Attributes) - 1
					currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + string(token.input[token.curr])
				}
			}
			token.curr++
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
			switch token.input[token.curr] {
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
			token.curr++

		case BeforeAttributeValue:
			switch token.input[token.curr] {
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
			token.curr++

		case AttributeValueDoubleQuoted:
			switch token.input[token.curr] {
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
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + string(token.input[token.curr])
			}
			token.curr++

		case AttributeValueSingleQuoted:
			switch token.input[token.curr] {
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
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + string(token.input[token.curr])
			}
			token.curr++

		case AttributeValueUnquoted:
			switch token.input[token.curr] {
			case tab, LF, FF, space: //whitespace
				state = BeforeAttributeName
			}

		default:
			fmt.Println("Not here yet")
			token.curr++
		}

	}
	return tokens
}
