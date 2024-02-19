package main

import "fmt"

type HTMLTokenizer struct {
	input string
	curr  int
}

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
			case 0x0026: //'&'
				state = CharacterReference
				returnState = append(returnState, int(Data))
			case 0x003C: //'<'
				state = TagOpen
			case 0x0000: //null
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
			case 0x0026: //'&'
				state = CharacterReference
				returnState = append(returnState, int(RCDATA))
			case 0x003C: //'<'
				state = RCDATALessThanSign
			case 0x0000: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD", // REPLACEMENT CHARACTER
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
			case 0x003C: //'<'
				state = RAWTEXTLessThanSign
			case 0x0000: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD", // REPLACEMENT CHARACTER
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
			case 0x003C: //'<'
				state = ScriptDataLessThanSign
			case 0x0000: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD", // REPLACEMENT CHARACTER
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
			case 0x0000: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD", // REPLACEMENT CHARACTER
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
			case 0x0021: //'!'
				state = MarkupDeclarationOpen
				token.curr++
			case 0x002F: //'/'
				state = EndTagOpen
				token.curr++
			case 0x003F: //'?'
				// Parse Error
				currToken.Type = CommentType
				currToken.Content = ""
				//Move to bogus comment
				state = BogusComment
			case endOfFile:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C", // LESS-THAN SIGN
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
						Content: "\u003C", // LESS-THAN SIGN
					}, &tokens)
					//Reconsume in data state
					state = Data
				}
			}
		case EndTagOpen:
			switch token.input[token.curr] {
			case 0x003E: //'>'
				// Parse Error
				state = Data
			case endOfFile:
				//Parse Error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // SOLIDUS
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
			case 0x0009: //'\t'
			case 0x000A: //'\n'
			case 0x000C: //'\f'
			case 0x0020: //' '
				state = BeforeAttributeName
			case 0x003E: //'>'
				state = Data
				emitToken(currToken, &tokens)
				if currToken.Type == StartTag {
					lastStartTagName = currToken.Name
				}
				currToken = HTMLToken{}
			case 0x002F: //'/'
				state = SelfClosingStartTag
			case 0x0000: //null
				// Parse error
				currToken.Name = currToken.Name + "\uFFFD" // REPLACEMENT CHARACTER
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
			case 0x002F: //SOLIDUS
				state = RCDATAEndTagOpen
				tmpBuffer = ""
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C", // LESS-THAN SIGN
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
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // SOLIDUS
				}, &tokens)
			}

		case RCDATAEndTagName:
			currChar := token.input[token.curr]
			if currChar == 0x0009 || currChar == 0x000A || currChar == 0x000C {
				//ignore whitespace
			} else if currChar == 0x0020 && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currChar == 0x003E && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currChar == 0x002F && lastStartTagName == currToken.Name {
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
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // Solidus
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
			case 0x002F: //SOLIDUS
				state = RAWTEXTEndTagOpen
				tmpBuffer = ""
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C", // LESS-THAN SIGN
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
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // SOLIDUS
				}, &tokens)
				//Reconsume in the RAWTEXT state
				state = RAWTEXT
			}

		case RAWTEXTEndTagName:
			currChar := token.input[token.curr]
			if currChar == 0x0009 || currChar == 0x000A || currChar == 0x000C { //tab, LF, FF
				//ignore whitespace
			} else if currChar == 0x0020 && lastStartTagName == currToken.Name {
				state = BeforeAttributeName
			} else if currChar == 0x003E && lastStartTagName == currToken.Name {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if currChar == 0x002F && lastStartTagName == currToken.Name {
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
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // SOLIDUS
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
					Content: "\u003C", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // SOLIDUS
				}, &tokens)
				//Reconsume in the Script data state
				state = Script
			}

		case ScriptDataEndTagName:
			currChar := token.input[token.curr]
			if currChar == 0x0009 || currChar == 0x000A || currChar == 0x000C {
				//ingore whitespace
			} else if currChar == 0x0020 && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currChar == 0x003E && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currChar) {
				currToken.Name = currToken.Name + string(currChar+0x0020)
				tmpBuffer = tmpBuffer + string(currChar)
			} else if isLowercase(currChar) {
				currToken.Name = currToken.Name + string(currChar)
				tmpBuffer = tmpBuffer + string(currChar)
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C",
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F",
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
			case 0x002D: //'-'
				state = ScriptDataEscapeStartDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			default:
				//Reconsume in the script data state
				state = Script
				token.curr--
			}
			token.curr++

		case ScriptDataEscapeStartDash:
			switch token.input[token.curr] {
			case 0x002D: //'-'
				state = ScriptDataEscapeStartDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			default:
				//Reconsume in the script data state
				state = Script
				token.curr--
			}
			token.curr++

		case ScriptDataEscapedDashDash:
			switch token.input[token.curr] {
			case 0x002D: //'-'
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			case 0x003C: //"<"
				state = ScriptDataEscapedLessThanSign
			case 0x003E: //">"
				state = Script
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003E",
				}, &tokens)
			case 0x0000:
				//Parse error
				state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD",
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
			if currChar == 0x002F { //Solidus
				tmpBuffer = ""
				state = ScriptDataEscapedEndTagOpen
			} else if isASCII(currChar) {
				tmpBuffer = ""
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C", // '<'
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
					Content: "\u003C", // '<'
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F", // '/'
				}, &tokens)
				state = ScriptDataEscaped
			}

		case ScriptDataEscapedEndTagName:
			currChar := token.input[token.curr]
			if currChar == 0x0009 || currChar == 0x000A || currChar == 0x000C {
				//ignore whitespace
			} else if currChar == 0x0020 && currToken.Name == lastStartTagName {
				state = BeforeAttributeName
			} else if currChar == 0x002F && currToken.Name == lastStartTagName {
				state = SelfClosingStartTag
			} else if currChar == 0x003E && currToken.Name == lastStartTagName {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currChar) {
				currToken.Name += string(currChar + 0x0020)
				tmpBuffer += string(currChar)
			} else if isLowercase(currChar) {
				currToken.Name += string(currChar)
				tmpBuffer += string(currChar)
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C",
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F",
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
			if currChar == 0x0009 || currChar == 0x000A ||
				currChar == 0x000C || currChar == 0x0020 || currChar == 0x002F {
				//ignore whitespace
			} else if currChar == 0x003E {
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
				tmpBuffer += string(currChar + 0x0020)
				emitToken(HTMLToken{
					Type:    Character,
					Content: string(currChar),
				}, &tokens)
			} else if isLowercase(currChar) {
				tmpBuffer += string(currChar + 0x0020)
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
			case 0x002D: // '-'
				state = ScriptDataDoubleEscapedDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			case 0x003C: //<
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C",
				}, &tokens)
			case 0x0000: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD",
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
			case 0x002D: // '-'
				state = ScriptDataDoubleEscapedDashDash
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			case 0x003C: // <
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C",
				}, &tokens)
			case 0x0000: // null
				//Parse error
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD",
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
			case 0x002D: // '-'
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002D",
				}, &tokens)
			case 0x003C: //'<'
				state = ScriptDataDoubleEscapedLessThanSign
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003C",
				}, &tokens)
			case 0x003E: //'>'
				state = Script
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u003E",
				}, &tokens)
			case 0x0000: //null
				//Parse error
				state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\uFFFD",
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
			if token.input[token.curr] == 0x002F {
				tmpBuffer = ""
				state = ScriptDataDoubleEscapeEnd
				emitToken(HTMLToken{
					Type:    Character,
					Content: "\u002F",
				}, &tokens)
			} else {
				state = ScriptDataDoubleEscaped
				token.curr--
			}
			token.curr++

		case ScriptDataDoubleEscapeEnd:
			currChar := token.input[token.curr]
			if currChar == 0x0009 || currChar == 0x000A || currChar == 0x000C ||
				currChar == 0x0020 || currChar == 0x002F {
				//ignore whitespace
			} else if currChar == 0x003E {
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
				tmpBuffer += string(currChar + 0x0020)
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
			case 0x0009, 0x000A, 0x000C, 0x0020: //whitespace
			case 0x003E, 0x002F: // '>' or '/'
			case endOfFile: //EOF
				state = AfterAttributeName
				token.curr--
			case 0x003D: // '='
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
			case 0x0009, 0x000A, 0x000C, 0x0020, 0x002F, 0x003E: //whitespace, '/' or '>'
			case endOfFile:
				state = AfterAttributeName
				token.curr--
			case 0x003D:
				state = BeforeAttributeValue
			case 0x0000:
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + "\uFFFD"
			default:
				if isUppercase(token.input[token.curr]) {
					idx := len(currToken.Attributes) - 1
					currToken.Attributes[idx].Name = currToken.Attributes[idx].Name + string(token.input[token.curr]+0x0020)
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
			case 0x0009, 0x000A, 0x000C, 0x0020: //whitespace
			case 0x002F: // '/'
				state = SelfClosingStartTag
			case 0x003D: // '='
				state = BeforeAttributeValue
			case 0x003E: // '>'
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
			case 0x0009, 0x000A, 0x000C, 0x0020: //whitespace
			case 0x0022: // '"'
				state = AttributeValueDoubleQuoted
			case 0x0027: // '\''
				state = AttributeValueSingleQuoted
			case 0x003E: //">"
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
			case 0x0022: //'"'
				state = AfterAttributeValueQuoted
			case 0x0026: //'&'
				returnState = append(returnState, int(AttributeValueDoubleQuoted))
				state = CharacterReference
			case 0x0000:
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + "\uFFFD"
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
			case 0x0027: //'
				state = AfterAttributeValueQuoted
			case 0x0026: //	"&"
				returnState = append(returnState, int(AttributeValueSingleQuoted))
				state = CharacterReference
			case 0x0000: //null
				//Parse error
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + "\uFFFD"
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(currToken.Attributes) - 1
				currToken.Attributes[idx].Value = currToken.Attributes[idx].Value + string(token.input[token.curr])
			}
			token.curr++

		default:
			fmt.Println("Not here yet")
			token.curr++
		}

	}
	return tokens
}
