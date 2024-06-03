package main

import (
	"fmt"
)

type HTMLTokenizer struct {
	input []byte
	curr  int
}

//MAIN TOKENIZER FUNCTION

func TokenizeHTML(token *HTMLTokenizer) []HTMLToken {

	//var space
	var tokens []HTMLToken
	var state State = Data
	var currToken HTMLToken
	var endOfFile byte = byte(0)
	var tmpBuffer string = ""
	var lastStartTagName string = ""
	var returnState []State //a stack of return states, values are ints from enums State

	for token.curr < len(token.input) {
		currVal := token.input[token.curr]
		switch state {
		case Data:
			switch currVal {
			case ampersand:
				state = CharacterReference
				returnState = append(returnState, Data)
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
				returnState = append(returnState, RCDATA)
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
				reconsume(&state, BogusComment, &token.curr)
			case endOfFile:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<", // LESS-THAN SIGN
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)

			default:
				if isASCIIAlpha(currVal) {
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
				if isASCIIAlpha(currVal) {
					// create a new end tag token
					currToken.Type = EndTag
					currToken.Name = ""
					currToken.Attributes = []Attribute{}
					currToken.SelfClosingFlag = false
					//reconsume in tag name state
					reconsume(&state, TagName, &token.curr)
				} else {
					// Parse Error
					currToken.Type = CommentType
					currToken.Content = ""
					//Reconsume in data state
					reconsume(&state, Data, &token.curr)
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
				reconsume(&state, RCDATA, &token.curr)
			}

		case RCDATAEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
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
				reconsume(&state, RCDATA, &token.curr)
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
				reconsume(&state, RAWTEXT, &token.curr)
			}

		case RAWTEXTEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
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
				reconsume(&state, RAWTEXT, &token.curr)
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
				reconsume(&state, Script, &token.curr)
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
				reconsume(&state, Script, &token.curr)
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
				reconsume(&state, Script, &token.curr)
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

			if currVal == solidus {
				tmpBuffer = ""
				state = ScriptDataEscapedEndTagOpen
			} else if isASCIIAlpha(currVal) {
				tmpBuffer = ""
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
				reconsume(&state, ScriptDataEscaped, &token.curr)
			}

		case ScriptDataEscapedEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
				currToken.Type = EndTag
				currToken.Name = ""
				state = ScriptDataEscapedEndTagName
			} else {
				emitToken(HTMLToken{
					Type:    Character,
					Content: "<",
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: "/",
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
				reconsume(&state, ScriptDataEscaped, &token.curr)
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
				reconsume(&state, ScriptDataEscaped, &token.curr)
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
				reconsume(&state, ScriptDataDoubleEscaped, &token.curr)
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
				reconsume(&state, ScriptDataDoubleEscaped, &token.curr)
			}

		case BeforeAttributeName:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case greaterThan, solidus: // '>' or '/'
			case endOfFile: //EOF
				reconsume(&state, AfterAttributeName, &token.curr)
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
				reconsume(&state, AttributeName, &token.curr)
			}

		case AttributeName:
			switch currVal {
			case tab, LF, FF, space, solidus, greaterThan: //whitespace, '/' or '>'
			case endOfFile:
				reconsume(&state, AfterAttributeName, &token.curr)
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
				reconsume(&state, AttributeName, &token.curr)
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
				reconsume(&state, AttributeValueUnquoted, &token.curr)
			}

		case AttributeValueDoubleQuoted:
			switch currVal {
			case quoteMark: //'"'
				state = AfterAttributeValueQuoted
			case ampersand: //'&'
				returnState = append(returnState, AttributeValueDoubleQuoted)
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
				returnState = append(returnState, AttributeValueSingleQuoted)
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
				returnState = append(returnState, AttributeValueUnquoted)
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
				reconsume(&state, BeforeAttributeName, &token.curr)
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
				reconsume(&state, BeforeAttributeName, &token.curr)
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
			} else if Lower(token.input[token.curr:token.curr+7]) == "doctype" {
				state = Doctype
				token.curr += 6
			} else if Lower(token.input[token.curr:token.curr+8]) == "[cdata[" {
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
				reconsume(&state, BogusComment, &token.curr)
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
				reconsume(&state, Comment, &token.curr)
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
				reconsume(&state, Comment, &token.curr)
			}

		case CommentLessThanSign:
			switch currVal {
			case exclamationMark:
				state = CommentLessThanSignBang
				currToken.Content = currToken.Content + string(rune(exclamationMark))
			case lesserThan:
				currToken.Content = currToken.Content + string(rune(exclamationMark))
			default:
				reconsume(&state, Comment, &token.curr)
			}

		case CommentLessThanSignBang:
			if currVal == dash {
				state = CommentLessThanSignBangDash
			} else {
				reconsume(&state, Comment, &token.curr)
			}

		case CommentLessThanSignBangDash:
			if currVal == dash {
				state = CommentLessThanSignBangDashDash
			} else {
				reconsume(&state, CommentEndDash, &token.curr)
			}

		case CommentLessThanSignBangDashDash:
			if currVal == greaterThan || currVal == endOfFile {
				reconsume(&state, CommentEnd, &token.curr)
			} else {
				//Nested comment parse error
				reconsume(&state, CommentEnd, &token.curr)
			}

		case CommentEndDash:
			switch currVal {
			case dash:
				state = CommentEnd
			case endOfFile:
				//EOF in comment error
				emitToken(currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.Content += "-"
				reconsume(&state, Comment, &token.curr)
			}

		case CommentEnd:
			switch currVal {
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case exclamationMark:
				state = CommentEndBang
			case dash:
				currToken.Content += "-"
			case endOfFile:
				//EOF in commment parse error
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.Content += "--"
				reconsume(&state, Comment, &token.curr)
			}

		case CommentEndBang:
			switch currVal {
			case dash:
				currToken.Content += "--!"
				state = CommentEndDash
			case greaterThan:
				//Incorrectly closed comment parse error
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//EOF in comment parse error
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.Content += "--!"
				reconsume(&state, Comment, &token.curr)
			}

		case Doctype:
			switch currVal {
			case tab, LF, FF, space:
				state = BeforeDOCTYPEName
			case greaterThan:
				reconsume(&state, BeforeAttributeName, &token.curr)
			case endOfFile:
				//end of file in doctype error
				emitToken(currToken, &tokens)
				currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//missing whitespace before name parse error
				reconsume(&state, BeforeDOCTYPEName, &token.curr)
			}

		case BeforeDOCTYPEName:
			if currVal == tab || currVal == LF || currVal == FF || currVal == space {
				//ignore the char
			} else if isUppercase(currVal) {
				currToken = HTMLToken{
					Type: DOCTYPE,
					Name: string(currVal),
				}
				state = DOCTYPEName
			} else if currVal == null {
				//unexpected null char parse error
				currToken = HTMLToken{
					Type: DOCTYPE,
					Name: replacementChar,
				}
				state = DOCTYPEName
			} else if currVal == greaterThan {
				//missing doctype name parse error
				emitToken(currToken, &tokens)
				currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				state = Data
			} else if currVal == endOfFile {
				//EOF in doctype error
				emitToken(currToken, &tokens)
				currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			} else {
				currToken = HTMLToken{
					Type: DOCTYPE,
					Name: string(currVal),
				}
				state = DOCTYPEName
			}

		case DOCTYPEName:
			if currVal == tab || currVal == LF || currVal == FF || currVal == space {
				state = AfterDOCTYPEName
			} else if currVal == greaterThan {
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			} else if isUppercase(currVal) {
				currToken.Name += string(currVal - 0x20)
			} else if currVal == null {
				//unexpected null char error
				currToken.Name += replacementChar
			} else if currVal == endOfFile {
				//EOF in doctype error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			} else {
				currToken.Name += string(currVal)
			}

		case AfterDOCTYPEName:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//EOF in Doctype Parse Error
				currToken.ForceQuirksFlag = true
			default:
				if nameInDoctype(token, token.curr, true) {
					token.curr += 5 //plus the default increment
					state = AfterDOCTYPEPublicKeyword
				} else if nameInDoctype(token, token.curr, false) {
					token.curr += 5
					state = AfterDOCTYPESystemKeyword
				} else {
					//Invalid char seq
					currToken.ForceQuirksFlag = true
					reconsume(&state, BogusDOCTYPE, &token.curr)
				}
			}

		case AfterDOCTYPEPublicKeyword:
			switch currVal {
			case tab, LF, FF, space:
				state = BeforeDOCTYPEPublicIdentifier
			case quoteMark:
				//missing whitespace after doctype public keyword error
				currToken.PublicID = ""
				state = DOCTYPEPublicIdentifierDoubleQuoted
			case apostrophe:
				//error
				currToken.PublicID = ""
				state = DOCTYPEPublicIdentifierSingleQuoted
			case greaterThan:
				//error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				reconsume(&state, BogusDOCTYPE, &token.curr)
			default:
				//error
				currToken.ForceQuirksFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case BeforeDOCTYPEPublicIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case quoteMark:
				currToken.PublicID = ""
				state = DOCTYPEPublicIdentifierDoubleQuoted
			case apostrophe:
				currToken.PublicID = ""
				state = DOCTYPEPublicIdentifierSingleQuoted
			case greaterThan:
				//error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				currToken.ForceQuirksFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case DOCTYPEPublicIdentifierDoubleQuoted:
			switch currVal {
			case quoteMark:
				state = AfterDOCTYPEPublicIdentifier
			case null:
				//error
				currToken.PublicID += replacementChar
			case greaterThan:
				//error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.PublicID += string(currVal)
			}

		case DOCTYPEPublicIdentifierSingleQuoted:
			switch currVal {
			case apostrophe:
				state = AfterDOCTYPEPublicIdentifier
			case null:
				//parse error
				currToken.PublicID += replacementChar
			case greaterThan:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.PublicID += string(currVal)
			}

		case AfterDOCTYPEPublicIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				state = BetweenDOCTYPEPublicAndSystemIdentifiers
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case quoteMark:
				//parse error
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				//parse error
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierSingleQuoted
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				currToken.SelfClosingFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case BetweenDOCTYPEPublicAndSystemIdentifiers:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case quoteMark:
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierSingleQuoted
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				currToken.ForceQuirksFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case AfterDOCTYPESystemKeyword:
			switch currVal {
			case tab, LF, FF, space:
				state = BeforeDOCTYPESystemIdentifier
			case quoteMark:
				//parse error
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				//parse error
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierSingleQuoted
			case greaterThan:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				currToken.ForceQuirksFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case BeforeDOCTYPESystemIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case quoteMark:
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				currToken.SystemID = ""
				state = DOCTYPESystemIdentifierSingleQuoted
			case greaterThan:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				currToken.ForceQuirksFlag = true
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case DOCTYPESystemIdentifierDoubleQuoted:
			switch currVal {
			case quoteMark:
				state = AfterDOCTYPESystemIdentifier
			case null:
				//parse error
				currToken.SystemID += replacementChar
			case greaterThan:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.SystemID += string(currVal)
			}

		case DOCTYPESystemIdentifierSingleQuoted:
			switch currVal {
			case apostrophe:
				state = AfterDOCTYPESystemIdentifier
			case null:
				//parse
				currToken.SystemID += replacementChar
			case greaterThan:
				//parse error
				currToken.ForceQuirksFlag = true
				state = Data
				emitToken(currToken, &tokens)
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				currToken.SystemID += string(currVal)
			}

		case AfterDOCTYPESystemIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
			case endOfFile:
				//parse error
				currToken.ForceQuirksFlag = true
				emitToken(currToken, &tokens)
				currToken = HTMLToken{}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error, no flag
				reconsume(&state, BogusDOCTYPE, &token.curr)
			}

		case BogusDOCTYPE:
			switch currVal {
			case greaterThan:
				state = Data
				emitToken(currToken, &tokens)
			case null:
				//parse error
				//ignore
			case endOfFile:
				emitToken(currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//ignore
			}

		case CDATASection:
			switch currVal {
			case rightSquareBracket:
				state = CDATASectionBracket
			case endOfFile:
				//parse error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitToken(currToken, &tokens)
			}

		case CDATASectionBracket:
			switch currVal {
			case rightSquareBracket:
				state = CDATASectionEnd
			default:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "]",
				}, &tokens)
				reconsume(&state, CDATASection, &token.curr)
			}

		case CDATASectionEnd:
			switch currVal {
			case rightSquareBracket:
				emitToken(HTMLToken{
					Type:    Character,
					Content: "]",
				}, &tokens)
			case greaterThan:
				state = Data
			default:
				rightBracket := HTMLToken{
					Type:    Character,
					Content: "]",
				}
				emitToken(rightBracket, &tokens)
				emitToken(rightBracket, &tokens)
				reconsume(&state, CDATASection, &token.curr)
			}

		case CharacterReference:
			tmpBuffer = "&"
			if isASCIIAlphanumeric(currVal) {
				reconsume(&state, NamedCharacterReference, &token.curr)
			} else if currVal == numberSign {
				tmpBuffer += string(currVal)
				state = NumericCharacterReference
			} else {
				prev := popState(&returnState)
				flushCodePoints(&currToken, prev, tmpBuffer, &tokens)
				tmpBuffer = ""
				reconsume(&state, prev, &token.curr)
			}

		case NamedCharacterReference:
			for isASCIIAlpha(currVal) {
				tmpBuffer += string(currVal)
				token.curr++
				currVal = token.input[token.curr]
			}
			token.curr--

		case AmbiguousAmpersand:
			if isASCIIAlphanumeric(currVal) {
				fmt.Println("Not here yet")
			}

		default:
			fmt.Println("Not here yet")
		}
		//consume the current current character (token.curr-- in individual cases otherwise or reconsume func)
		token.curr++

	}
	return tokens
}
