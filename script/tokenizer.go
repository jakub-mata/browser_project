package main

import (
	"fmt"
	"strings"
)

const endOfFile byte = byte(1)

type HTMLTokenizer struct {
	input            []byte
	curr             int
	state            State
	lastStartTagName string
	returnState      []State //a stack of return states, values are ints from enums State
	tmpBuffer        strings.Builder
	currToken        HTMLToken
}

//MAIN TOKENIZER FUNCTION

func (tokenizer HTMLTokenizer) TokenizeHTML(printTokens bool) []HTMLToken {

	var tokens []HTMLToken

	for tokenizer.curr < len(tokenizer.input) {
		currVal := tokenizer.input[tokenizer.curr]
		switch tokenizer.state {
		case Data:
			switch currVal {
			case ampersand:
				tokenizer.state = CharacterReference
				tokenizer.returnState = append(tokenizer.returnState, Data)
			case lesserThan:
				if isNotWhitespace(tokenizer.currToken.Content.String()) {
					tokenizer.currToken.Type = Character
					emitCurrToken(&tokenizer.currToken, &tokens)
				}
				tokenizer.state = TagOpen
			case null:
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(string(currVal)),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(currVal)
			}

		case RCDATA:
			switch currVal {
			case ampersand:
				tokenizer.state = CharacterReference
				tokenizer.returnState = append(tokenizer.returnState, RCDATA)
			case lesserThan:
				tokenizer.state = RCDATALessThanSign
			case null:
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(currVal)
			}

		case RAWTEXT:
			switch currVal {
			case lesserThan:
				tokenizer.state = RAWTEXTLessThanSign
			case null:
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(currVal)
			}

		case Script:
			switch currVal {
			case lesserThan:
				tokenizer.state = ScriptDataLessThanSign
			case null:
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.tmpBuffer.WriteByte(currVal)
			}

		case PLAINTEXT:
			switch currVal {
			case null:
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.tmpBuffer.WriteByte(currVal)
			}

		case TagOpen:
			switch currVal {
			case exclamationMark:
				tokenizer.state = MarkupDeclarationOpen
			case solidus:
				tokenizer.state = EndTagOpen
			case questionMark:
				// Parse Error
				tokenizer.currToken.Type = CommentType
				reconsume(&tokenizer.state, BogusComment, &tokenizer.curr)
			case endOfFile:
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(string(rune(lesserThan))),
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)

			default:
				if isASCIIAlpha(currVal) {
					// create a new start tag token
					tokenizer.currToken.Type = StartTag
					tokenizer.currToken.Name = ""
					tokenizer.currToken.Attributes = []Attribute{}
					tokenizer.currToken.SelfClosingFlag = false
					reconsume(&tokenizer.state, TagName, &tokenizer.curr)
				} else {
					// Parse Error
					emitToken(HTMLToken{
						Type:    Character,
						Content: toBuilder(string(rune(lesserThan))),
					}, &tokens)
					reconsume(&tokenizer.state, Data, &tokenizer.curr)
				}
			}

		case EndTagOpen:
			switch currVal {
			case greaterThan:
				// Parse Error
				tokenizer.state = Data
			case endOfFile:
				//Parse Error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(string(rune(lesserThan))),
				}, &tokens)
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(string(rune(solidus))),
				}, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				if isASCIIAlpha(currVal) {
					// create a new end tag token
					tokenizer.currToken.Type = EndTag
					tokenizer.currToken.Name = ""
					tokenizer.currToken.Attributes = []Attribute{}
					tokenizer.currToken.SelfClosingFlag = false
					reconsume(&tokenizer.state, TagName, &tokenizer.curr)
				} else {
					// Parse Error
					tokenizer.currToken.Type = CommentType
					tokenizer.currToken.Content = strings.Builder{}
					reconsume(&tokenizer.state, BogusComment, &tokenizer.curr)
				}
			}

		case TagName:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BeforeAttributeName
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
				if tokenizer.currToken.Type == StartTag {
					tokenizer.lastStartTagName = tokenizer.currToken.Name
				}
			case solidus:
				tokenizer.state = SelfClosingStartTag
			case null: //null
				// Parse error
				tokenizer.currToken.Name += replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)

			default:
				if isUppercase(currVal) {
					tokenizer.currToken.Name += string(currVal + 0x20)
				} else {
					tokenizer.currToken.Name += string(currVal)
				}
			}

		case RCDATALessThanSign:
			switch currVal {
			case solidus:
				tokenizer.state = RCDATAEndTagOpen
				tokenizer.tmpBuffer = strings.Builder{}
			default:
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				reconsume(&tokenizer.state, RCDATA, &tokenizer.curr)
			}

		case RCDATAEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
				tokenizer.currToken.Type = EndTag
				tokenizer.currToken.Name = ""
				tokenizer.state = RCDATAEndTagName
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)
			}
			tokenizer.curr--

		case RCDATAEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ignore whitespace
			} else if currVal == space && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = BeforeAttributeName
			} else if currVal == greaterThan && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			} else if currVal == solidus && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = SelfClosingStartTag
			} else if isUppercase(currVal) {
				tokenizer.currToken.Name += tokenizer.currToken.Name + string(currVal+0x20)
				tokenizer.tmpBuffer.WriteByte(currVal + 0x20)
			} else if isLowercase(currVal) {
				//reconsume in RCDATA state
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.currToken.Name = tokenizer.currToken.Name + string(currVal)
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)

				tokenizer.currToken.Content = tokenizer.tmpBuffer
				emitCurrToken(&tokenizer.currToken, &tokens)
				reconsume(&tokenizer.state, RCDATA, &tokenizer.curr)
			}

		case RAWTEXTLessThanSign:
			switch currVal {
			case solidus: //SOLIDUS
				tokenizer.state = RAWTEXTEndTagOpen
				tokenizer.tmpBuffer = strings.Builder{}
			default:
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				//Reconsume in RAWTEXT state
				reconsume(&tokenizer.state, RAWTEXT, &tokenizer.curr)
			}

		case RAWTEXTEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
				tokenizer.currToken.Type = EndTag
				tokenizer.currToken.Name = ""
				tokenizer.state = RAWTEXTEndTagName
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)
				//Reconsume in the RAWTEXT state
				tokenizer.state = RAWTEXT
			}
			tokenizer.curr--

		case RAWTEXTEndTagName:

			if currVal == tab || currVal == LF || currVal == FF { //tab, LF, FF
				//ignore whitespace
			} else if currVal == space && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = BeforeAttributeName
			} else if currVal == greaterThan && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			} else if currVal == solidus && tokenizer.lastStartTagName == tokenizer.currToken.Name {
				tokenizer.state = SelfClosingStartTag
			} else if isUppercase(currVal) {
				tokenizer.currToken.Name += tokenizer.currToken.Name + string(currVal+0x20)
				tokenizer.tmpBuffer.WriteByte(currVal + 0x20)
			} else if isLowercase(currVal) {
				tokenizer.currToken.Name += tokenizer.currToken.Name + string(currVal)
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)

				tokenizer.currToken.Content = tokenizer.tmpBuffer
				emitCurrToken(&tokenizer.currToken, &tokens)
				reconsume(&tokenizer.state, RAWTEXT, &tokenizer.curr)
			}

		case ScriptDataLessThanSign:
			// All reconsumed

			if isUppercase(currVal) {
				tokenizer.currToken.Type = EndTag
				tokenizer.currToken.Name = ""
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)
				//Reconsume in the Script data state
				tokenizer.state = Script
			}
			tokenizer.curr--

		case ScriptDataEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ingore whitespace
			} else if currVal == space && tokenizer.currToken.Name == tokenizer.lastStartTagName {
				tokenizer.state = SelfClosingStartTag
			} else if currVal == greaterThan && tokenizer.currToken.Name == tokenizer.lastStartTagName {
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			} else if isUppercase(currVal) {
				tokenizer.currToken.Name = tokenizer.currToken.Name + string(currVal+space)
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else if isLowercase(currVal) {
				tokenizer.currToken.Name = tokenizer.currToken.Name + string(currVal)
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)

				tokenizer.currToken.Content = tokenizer.tmpBuffer
				emitCurrToken(&tokenizer.currToken, &tokens)
				reconsume(&tokenizer.state, Script, &tokenizer.curr)
			}

		case ScriptDataEscaped:
			switch currVal {
			case lesserThan:
				tokenizer.state = ScriptDataDoubleEscapedLessThanSign
			case dash:
				tokenizer.state = ScriptDataEscapedDash
			case null:
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.tmpBuffer.WriteByte(currVal)
			}

		case ScriptDataEscapedDash:
			switch currVal {
			case dash: //'-'
				tokenizer.state = ScriptDataEscapeStartDash
				tokenizer.tmpBuffer.WriteByte(dash)
			default:
				//Reconsume in the script data state
				reconsume(&tokenizer.state, Script, &tokenizer.curr)
			}

		case ScriptDataEscapeStartDash:
			switch currVal {
			case dash: //'-'
				tokenizer.state = ScriptDataEscapeStartDash
				tokenizer.tmpBuffer.WriteByte(dash)
			default:
				//Reconsume in the script data state
				reconsume(&tokenizer.state, Script, &tokenizer.curr)
			}

		case ScriptDataEscapedDashDash:
			switch currVal {
			case dash: //'-'
				tokenizer.tmpBuffer.WriteByte(dash)
			case lesserThan: //"<"
				tokenizer.state = ScriptDataEscapedLessThanSign
			case greaterThan: //">"
				tokenizer.state = Script
				tokenizer.tmpBuffer.WriteByte(greaterThan)
			case null:
				//Parse error
				tokenizer.state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				// Parse error
				tokenizer.state = ScriptDataEscaped
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.state = ScriptDataEscaped
				tokenizer.currToken.Content.WriteByte(currVal)
			}

		case ScriptDataEscapedLessThanSign:

			if currVal == solidus {
				tokenizer.tmpBuffer.Reset()
				tokenizer.state = ScriptDataEscapedEndTagOpen
			} else if isASCIIAlpha(currVal) {
				tokenizer.tmpBuffer.Reset()
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				reconsume(&tokenizer.state, ScriptDataEscaped, &tokenizer.curr)
			}

		case ScriptDataEscapedEndTagOpen:
			// All reconsumed
			if isASCIIAlpha(currVal) {
				tokenizer.currToken.Type = EndTag
				tokenizer.currToken.Name = ""
				tokenizer.state = ScriptDataEscapedEndTagName
			} else {
				tokenizer.tmpBuffer.WriteByte(lesserThan)
				tokenizer.tmpBuffer.WriteByte(solidus)
				tokenizer.state = ScriptDataEscaped
			}
			tokenizer.curr--

		case ScriptDataEscapedEndTagName:

			if currVal == tab || currVal == LF || currVal == FF {
				//ignore whitespace
			} else if currVal == space && tokenizer.currToken.Name == tokenizer.lastStartTagName {
				tokenizer.state = BeforeAttributeName
			} else if currVal == solidus && tokenizer.currToken.Name == tokenizer.lastStartTagName {
				tokenizer.state = SelfClosingStartTag
			} else if currVal == greaterThan && tokenizer.currToken.Name == tokenizer.lastStartTagName {
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			} else if isUppercase(currVal) {
				tokenizer.currToken.Name += string(currVal + space)
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else if isLowercase(currVal) {
				tokenizer.currToken.Name += string(currVal)
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else {
				tokenizer.tmpBuffer.WriteString("</")

				tokenizer.currToken.Content = tokenizer.tmpBuffer
				emitCurrToken(&tokenizer.currToken, &tokens)
				//Reconsume
				reconsume(&tokenizer.state, ScriptDataEscaped, &tokenizer.curr)
			}

		case ScriptDataDoubleEscapeStart:

			if currVal == tab || currVal == LF ||
				currVal == FF || currVal == space || currVal == solidus {
				//ignore whitespace
			} else if currVal == greaterThan {
				if tokenizer.tmpBuffer.String() == "script" {
					tokenizer.state = ScriptDataDoubleEscaped
				} else {
					tokenizer.state = ScriptDataEscaped
				}
				tokenizer.tmpBuffer.WriteByte(currVal)

			} else if isUppercase(currVal) {
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.tmpBuffer.WriteByte(space)

			} else if isLowercase(currVal) {
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.tmpBuffer.WriteByte(space)
			} else {
				//Reconsume
				reconsume(&tokenizer.state, ScriptDataEscaped, &tokenizer.curr)
			}

		case ScriptDataDoubleEscaped:
			switch currVal {
			case dash: // '-'
				tokenizer.state = ScriptDataDoubleEscapedDash
				tokenizer.tmpBuffer.WriteByte(dash)
			case lesserThan: //<
				tokenizer.state = ScriptDataDoubleEscapedLessThanSign
				tokenizer.tmpBuffer.WriteByte(lesserThan)
			case null: //null
				// Parse error
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
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
				tokenizer.state = ScriptDataDoubleEscapedDashDash
				tokenizer.tmpBuffer.WriteByte(dash)
			case lesserThan: // <
				tokenizer.state = ScriptDataDoubleEscapedLessThanSign
				tokenizer.tmpBuffer.WriteByte(lesserThan)
			case null: // null
				//Parse error
				tokenizer.state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.state = ScriptDataDoubleEscaped
				tokenizer.tmpBuffer.WriteByte(currVal)
			}

		case ScriptDataDoubleEscapedDashDash:
			switch currVal {
			case dash: // '-'
				tokenizer.tmpBuffer.WriteByte(dash)
			case lesserThan: //'<'
				tokenizer.state = ScriptDataDoubleEscapedLessThanSign
				tokenizer.tmpBuffer.WriteByte(lesserThan)
			case greaterThan: //'>'
				tokenizer.state = Script
				tokenizer.tmpBuffer.WriteByte(greaterThan)
			case null: //null
				//Parse error
				tokenizer.state = ScriptDataDoubleEscaped
				emitToken(HTMLToken{
					Type:    Character,
					Content: toBuilder(replacementChar),
				}, &tokens)
			case endOfFile:
				//Parse error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.state = ScriptDataDoubleEscaped
				tokenizer.tmpBuffer.WriteByte(currVal)
			}

		case ScriptDataDoubleEscapedLessThanSign:
			if currVal == solidus {
				tokenizer.tmpBuffer.Reset()
				tokenizer.state = ScriptDataDoubleEscapeEnd
				tokenizer.tmpBuffer.WriteByte(solidus)
			} else {
				reconsume(&tokenizer.state, ScriptDataDoubleEscaped, &tokenizer.curr)
			}

		case ScriptDataDoubleEscapeEnd:

			if currVal == tab || currVal == LF || currVal == FF ||
				currVal == space || currVal == solidus {
				//ignore whitespace
			} else if currVal == greaterThan {
				if tokenizer.tmpBuffer.String() == "script" {
					tokenizer.state = ScriptDataEscaped
				} else {
					tokenizer.state = ScriptDataDoubleEscaped
				}
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else if isUppercase(currVal) {
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.tmpBuffer.WriteByte(space)
			} else if isLowercase(currVal) {
				tokenizer.tmpBuffer.WriteByte(currVal)
			} else {
				//Reconsume
				reconsume(&tokenizer.state, ScriptDataDoubleEscaped, &tokenizer.curr)
			}

		case BeforeAttributeName:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan, solidus, endOfFile:
				reconsume(&tokenizer.state, AfterAttributeName, &tokenizer.curr)
			case equal: // '='
				//Parse error
				tokenizer.currToken.Attributes = append(tokenizer.currToken.Attributes, Attribute{
					Name:  string(currVal),
					Value: "",
				})
				tokenizer.state = AttributeName
			default:
				tokenizer.currToken.Attributes = append(tokenizer.currToken.Attributes, Attribute{
					Name:  "",
					Value: "",
				})
				reconsume(&tokenizer.state, AttributeName, &tokenizer.curr)
			}

		case AttributeName:
			switch currVal {
			case tab, LF, FF, space, solidus, greaterThan: //whitespace, '/' or '>'
			case endOfFile:
				reconsume(&tokenizer.state, AfterAttributeName, &tokenizer.curr)
			case equal:
				tokenizer.state = BeforeAttributeValue
			case null:
				//Parse error
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Name = tokenizer.currToken.Attributes[idx].Name + replacementChar
			default:
				if isUppercase(currVal) {
					idx := len(tokenizer.currToken.Attributes) - 1
					tokenizer.currToken.Attributes[idx].Name = tokenizer.currToken.Attributes[idx].Name + string(currVal+space)
				} else {
					idx := len(tokenizer.currToken.Attributes) - 1
					tokenizer.currToken.Attributes[idx].Name = tokenizer.currToken.Attributes[idx].Name + string(currVal)
				}
			}

			//Checking duplicates
			if tokenizer.state != AttributeName {
				// Can be faster by counting occurences in a map, relying on not so many attributes
				for i := 0; i < len(tokenizer.currToken.Attributes); i++ {
					for j := i + 1; j < len(tokenizer.currToken.Attributes); j++ {
						if tokenizer.currToken.Attributes[i].Name == tokenizer.currToken.Attributes[j].Name {
							//Parse error, remove duplicate
							tokenizer.currToken.Attributes = append(tokenizer.currToken.Attributes[:j], tokenizer.currToken.Attributes[j+1:]...)
						}
					}
				}
			}

		case AfterAttributeName:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case solidus: // '/'
				tokenizer.state = SelfClosingStartTag
			case equal: // '='
				tokenizer.state = BeforeAttributeValue
			case greaterThan: // '>'
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//Parse Error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Attributes = append(tokenizer.currToken.Attributes, Attribute{
					Name:  "",
					Value: "",
				})
				reconsume(&tokenizer.state, AttributeName, &tokenizer.curr)
			}

		case BeforeAttributeValue:
			switch currVal {
			case tab, LF, FF, space: //whitespace
			case quoteMark: // '"'
				tokenizer.state = AttributeValueDoubleQuoted
			case apostrophe: // '\''
				tokenizer.state = AttributeValueSingleQuoted
			case greaterThan: //">"
				//Parse error, missing attribute
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			default:
				reconsume(&tokenizer.state, AttributeValueUnquoted, &tokenizer.curr)
			}

		case AttributeValueDoubleQuoted:
			switch currVal {
			case quoteMark: //'"'
				tokenizer.state = AfterAttributeValueQuoted
			case ampersand: //'&'
				tokenizer.returnState = append(tokenizer.returnState, AttributeValueDoubleQuoted)
				tokenizer.state = CharacterReference
			case null:
				//Parse error
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value = tokenizer.currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value += string(currVal)
			}

		case AttributeValueSingleQuoted:
			switch currVal {
			case apostrophe: //'
				tokenizer.state = AfterAttributeValueQuoted
			case ampersand: //	"&"
				tokenizer.returnState = append(tokenizer.returnState, AttributeValueSingleQuoted)
				tokenizer.state = CharacterReference
			case null: //null
				//Parse error
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value = tokenizer.currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value = tokenizer.currToken.Attributes[idx].Value + string(currVal)
			}

		case AttributeValueUnquoted:
			switch currVal {
			case tab, LF, FF, space: //whitespace
				tokenizer.state = BeforeAttributeName
			case ampersand:
				tokenizer.returnState = append(tokenizer.returnState, AttributeValueUnquoted)
				tokenizer.state = CharacterReference
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case null:
				//Parse error
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value = tokenizer.currToken.Attributes[idx].Value + replacementChar
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				idx := len(tokenizer.currToken.Attributes) - 1
				tokenizer.currToken.Attributes[idx].Value = tokenizer.currToken.Attributes[idx].Value + replacementChar
			}

		case AfterAttributeValueQuoted:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BeforeAttributeName
			case solidus:
				tokenizer.state = SelfClosingStartTag
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//Reconsume
				reconsume(&tokenizer.state, BeforeAttributeName, &tokenizer.curr)
			}

		case SelfClosingStartTag:
			switch currVal {
			case greaterThan:
				tokenizer.currToken.SelfClosingFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//Reconsume
				reconsume(&tokenizer.state, BeforeAttributeName, &tokenizer.curr)
			}

		case BogusComment:
			switch currVal {
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			case null:
				tokenizer.currToken.Content.WriteString(replacementChar)
			default:
				tokenizer.currToken.Content.WriteByte(currVal)
			}

		case MarkupDeclarationOpen:
			if currVal == dash && tokenizer.input[tokenizer.curr+1] == dash {
				tokenizer.currToken = HTMLToken{
					Type:    CommentType,
					Content: strings.Builder{},
				}
				tokenizer.state = CommentStart
				tokenizer.curr++
			} else if Lower(tokenizer.input[tokenizer.curr:tokenizer.curr+7]) == "doctype" {
				tokenizer.state = Doctype
				tokenizer.curr += 6
			} else if Lower(tokenizer.input[tokenizer.curr:tokenizer.curr+8]) == "[cdata[" {
				tokenizer.curr += 7
				//TO DO: implement CDATA section
				tokenizer.state = CDATASection
			} else {
				tokenizer.currToken = HTMLToken{
					Type:    CommentType,
					Content: strings.Builder{},
				}
				tokenizer.state = BogusComment
				tokenizer.curr--
				reconsume(&tokenizer.state, BogusComment, &tokenizer.curr)
			}

		case CommentStart:
			switch currVal {
			case dash:
				tokenizer.state = CommentStartDash
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			default:
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentStartDash:
			switch currVal {
			case dash:
				tokenizer.state = CommentEnd
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(dash)
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentLessThanSign:
			switch currVal {
			case exclamationMark:
				tokenizer.state = CommentLessThanSignBang
				tokenizer.currToken.Content.WriteByte(exclamationMark)
			case lesserThan:
				tokenizer.currToken.Content.WriteByte(lesserThan)
			default:
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentLessThanSignBang:
			if currVal == dash {
				tokenizer.state = CommentLessThanSignBangDash
			} else {
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentLessThanSignBangDash:
			if currVal == dash {
				tokenizer.state = CommentLessThanSignBangDashDash
			} else {
				reconsume(&tokenizer.state, CommentEndDash, &tokenizer.curr)
			}

		case CommentLessThanSignBangDashDash:
			if currVal == greaterThan || currVal == endOfFile {
				reconsume(&tokenizer.state, CommentEnd, &tokenizer.curr)
			} else {
				//Nested comment parse error
				reconsume(&tokenizer.state, CommentEnd, &tokenizer.curr)
			}

		case CommentEndDash:
			switch currVal {
			case dash:
				tokenizer.state = CommentEnd
			case endOfFile:
				//EOF in comment error
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(dash)
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentEnd:
			switch currVal {
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case exclamationMark:
				tokenizer.state = CommentEndBang
			case dash:
				tokenizer.currToken.Content.WriteByte(dash)
			case endOfFile:
				//EOF in commment parse error
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(dash)
				tokenizer.currToken.Content.WriteByte(dash)
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case CommentEndBang:
			switch currVal {
			case dash:
				tokenizer.currToken.Content.WriteByte(dash)
				tokenizer.currToken.Content.WriteByte(dash)
				tokenizer.currToken.Content.WriteByte(exclamationMark)
				tokenizer.state = CommentEndDash
			case greaterThan:
				//Incorrectly closed comment parse error
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//EOF in comment parse error
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.Content.WriteByte(dash)
				tokenizer.currToken.Content.WriteByte(dash)
				tokenizer.currToken.Content.WriteByte(exclamationMark)
				reconsume(&tokenizer.state, Comment, &tokenizer.curr)
			}

		case Doctype:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BeforeDOCTYPEName
			case greaterThan:
				reconsume(&tokenizer.state, BeforeAttributeName, &tokenizer.curr)
			case endOfFile:
				//end of file in doctype error
				emitCurrToken(&tokenizer.currToken, &tokens)
				tokenizer.currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//missing whitespace before name parse error
				reconsume(&tokenizer.state, BeforeDOCTYPEName, &tokenizer.curr)
			}

		case BeforeDOCTYPEName:
			if currVal == tab || currVal == LF || currVal == FF || currVal == space {
				//ignore the char
			} else if isUppercase(currVal) {
				tokenizer.currToken = HTMLToken{
					Type: DOCTYPE,
					Name: string(currVal),
				}
				tokenizer.state = DOCTYPEName
			} else if currVal == null {
				//unexpected null char parse error
				tokenizer.currToken = HTMLToken{
					Type: DOCTYPE,
					Name: replacementChar,
				}
				tokenizer.state = DOCTYPEName
			} else if currVal == greaterThan {
				//missing doctype name parse error
				emitCurrToken(&tokenizer.currToken, &tokens)
				tokenizer.currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				tokenizer.state = Data
			} else if currVal == endOfFile {
				//EOF in doctype error
				emitCurrToken(&tokenizer.currToken, &tokens)
				tokenizer.currToken = HTMLToken{
					Type:            DOCTYPE,
					ForceQuirksFlag: true,
				}
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			} else {
				tokenizer.currToken = HTMLToken{
					Type: DOCTYPE,
					Name: string(currVal),
				}
				tokenizer.state = DOCTYPEName
			}

		case DOCTYPEName:
			if currVal == tab || currVal == LF || currVal == FF || currVal == space {
				tokenizer.state = AfterDOCTYPEName
			} else if currVal == greaterThan {
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			} else if isUppercase(currVal) {
				tokenizer.currToken.Name += string(currVal - 0x20)
			} else if currVal == null {
				//unexpected null char error
				tokenizer.currToken.Name += replacementChar
			} else if currVal == endOfFile {
				//EOF in doctype error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			} else {
				tokenizer.currToken.Name += string(currVal)
			}

		case AfterDOCTYPEName:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//EOF in Doctype Parse Error
				tokenizer.currToken.ForceQuirksFlag = true
			default:
				if nameInDoctype(&tokenizer, tokenizer.curr, true) {
					tokenizer.curr += 5 //plus the default increment
					tokenizer.state = AfterDOCTYPEPublicKeyword
				} else if nameInDoctype(&tokenizer, tokenizer.curr, false) {
					tokenizer.curr += 5
					tokenizer.state = AfterDOCTYPESystemKeyword
				} else {
					//Invalid char seq
					tokenizer.currToken.ForceQuirksFlag = true
					reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
				}
			}

		case AfterDOCTYPEPublicKeyword:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BeforeDOCTYPEPublicIdentifier
			case quoteMark:
				//missing whitespace after doctype public keyword error
				tokenizer.currToken.PublicID = ""
				tokenizer.state = DOCTYPEPublicIdentifierDoubleQuoted
			case apostrophe:
				//error
				tokenizer.currToken.PublicID = ""
				tokenizer.state = DOCTYPEPublicIdentifierSingleQuoted
			case greaterThan:
				//error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			default:
				//error
				tokenizer.currToken.ForceQuirksFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case BeforeDOCTYPEPublicIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case quoteMark:
				tokenizer.currToken.PublicID = ""
				tokenizer.state = DOCTYPEPublicIdentifierDoubleQuoted
			case apostrophe:
				tokenizer.currToken.PublicID = ""
				tokenizer.state = DOCTYPEPublicIdentifierSingleQuoted
			case greaterThan:
				//error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case DOCTYPEPublicIdentifierDoubleQuoted:
			switch currVal {
			case quoteMark:
				tokenizer.state = AfterDOCTYPEPublicIdentifier
			case null:
				//error
				tokenizer.currToken.PublicID += replacementChar
			case greaterThan:
				//error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.PublicID += string(currVal)
			}

		case DOCTYPEPublicIdentifierSingleQuoted:
			switch currVal {
			case apostrophe:
				tokenizer.state = AfterDOCTYPEPublicIdentifier
			case null:
				//parse error
				tokenizer.currToken.PublicID += replacementChar
			case greaterThan:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.PublicID += string(currVal)
			}

		case AfterDOCTYPEPublicIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BetweenDOCTYPEPublicAndSystemIdentifiers
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case quoteMark:
				//parse error
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				//parse error
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierSingleQuoted
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				tokenizer.currToken.SelfClosingFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case BetweenDOCTYPEPublicAndSystemIdentifiers:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case quoteMark:
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierSingleQuoted
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case AfterDOCTYPESystemKeyword:
			switch currVal {
			case tab, LF, FF, space:
				tokenizer.state = BeforeDOCTYPESystemIdentifier
			case quoteMark:
				//parse error
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				//parse error
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierSingleQuoted
			case greaterThan:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case BeforeDOCTYPESystemIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case quoteMark:
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierDoubleQuoted
			case apostrophe:
				tokenizer.currToken.SystemID = ""
				tokenizer.state = DOCTYPESystemIdentifierSingleQuoted
			case greaterThan:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case DOCTYPESystemIdentifierDoubleQuoted:
			switch currVal {
			case quoteMark:
				tokenizer.state = AfterDOCTYPESystemIdentifier
			case null:
				//parse error
				tokenizer.currToken.SystemID += replacementChar
			case greaterThan:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.SystemID += string(currVal)
			}

		case DOCTYPESystemIdentifierSingleQuoted:
			switch currVal {
			case apostrophe:
				tokenizer.state = AfterDOCTYPESystemIdentifier
			case null:
				//parse
				tokenizer.currToken.SystemID += replacementChar
			case greaterThan:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				tokenizer.currToken.SystemID += string(currVal)
			}

		case AfterDOCTYPESystemIdentifier:
			switch currVal {
			case tab, LF, FF, space:
				//ignore
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case endOfFile:
				//parse error
				tokenizer.currToken.ForceQuirksFlag = true
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//parse error, no flag
				reconsume(&tokenizer.state, BogusDOCTYPE, &tokenizer.curr)
			}

		case BogusDOCTYPE:
			switch currVal {
			case greaterThan:
				tokenizer.state = Data
				emitCurrToken(&tokenizer.currToken, &tokens)
			case null:
				//parse error
				//ignore
			case endOfFile:
				emitCurrToken(&tokenizer.currToken, &tokens)
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				//ignore
			}

		case CDATASection:
			switch currVal {
			case rightSquareBracket:
				tokenizer.state = CDATASectionBracket
			case endOfFile:
				//parse error
				emitToken(HTMLToken{
					Type: EOF,
				}, &tokens)
			default:
				emitCurrToken(&tokenizer.currToken, &tokens)
			}

		case CDATASectionBracket:
			switch currVal {
			case rightSquareBracket:
				tokenizer.state = CDATASectionEnd
			default:
				tokenizer.tmpBuffer.WriteByte(leftSquareBracket)
				reconsume(&tokenizer.state, CDATASection, &tokenizer.curr)
			}

		case CDATASectionEnd:
			switch currVal {
			case rightSquareBracket:
				tokenizer.tmpBuffer.WriteByte(leftSquareBracket)
			case greaterThan:
				tokenizer.state = Data
			default:
				rightBracket := HTMLToken{
					Type:    Character,
					Content: toBuilder(string(rune(rightSquareBracket))),
				}
				emitToken(rightBracket, &tokens)
				emitToken(rightBracket, &tokens)
				reconsume(&tokenizer.state, CDATASection, &tokenizer.curr)
			}

		case CharacterReference:
			tokenizer.tmpBuffer.Reset()
			tokenizer.tmpBuffer.WriteByte(ampersand)
			if isASCIIAlphanumeric(currVal) {
				reconsume(&tokenizer.state, NamedCharacterReference, &tokenizer.curr)
			} else if currVal == numberSign {
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.state = NumericCharacterReference
			} else {
				prev := popState(&tokenizer.returnState)
				flushCodePoints(&tokenizer.currToken, prev, tokenizer.tmpBuffer.String(), &tokens)
				tokenizer.tmpBuffer.Reset()
				reconsume(&tokenizer.state, prev, &tokenizer.curr)
			}

		case NamedCharacterReference:
			for isASCIIAlpha(currVal) {
				tokenizer.tmpBuffer.WriteByte(currVal)
				tokenizer.curr++
				currVal = tokenizer.input[tokenizer.curr]
			}
			tokenizer.curr--

		case AmbiguousAmpersand:
			if isASCIIAlphanumeric(currVal) {
				fmt.Println("Not here yet")
			}

		default:
			fmt.Println("Not here yet")
		}
		//consume the current current character (tokenizer.curr-- in individual cases otherwise or reconsume func)
		tokenizer.curr++

	}

	if printTokens {
		logTokens(&tokens)
	}
	return tokens
}

func logTokens(tokenOutput *[]HTMLToken) {

	for _, token := range *tokenOutput {
		fmt.Println(token)
	}
}
