package main

func isASCII(s byte) bool {
	if s > 127 {
		return false
	}
	return true
}

func isUppercase(s byte) bool {
	if s >= 65 && s <= 90 {
		return true
	}
	return false
}

type HTMLTokenizer struct {
	input string
	curr  int
}

func emitToken(tokenToEmit HTMLToken, tokens *[]HTMLToken) {
	*tokens = append(*tokens, tokenToEmit)
}

func NewHTMLTokenizer(input string) *HTMLTokenizer {
	return &HTMLTokenizer{input: input}
}

func TokenizeHTML(token *HTMLTokenizer) []HTMLToken {

	var tokens []HTMLToken
	var state State = Data
	var currToken HTMLToken
	var endOfFile byte = byte(0)
	var tmpBuffer string 
	var lastStartTagName string = ""

	for token.curr < len(token.input) {
		switch state {
		case Data:
			switch token.input[token.curr] {
			case 0x0026: //'&'
				state = CharacterReference
				//returnState(Data)
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
				//returnState(RCDATA)
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
				case '\u002F': //SOLIDUS
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
				emitToken({
					Type:    Character,
					Content: "\u002F", // SOLIDUS
				}, &tokens)
			}

		case RCDATAEndTagName:
			switch token.input[token.curr] {
				case 0x0009: //'\t'
				case 0x000A: // '\n'
				case 0x000C:  // '\f'
				case 0x0020:  // ' '
					if lastStartTagName == currToken.Name {
						state = BeforeAttributeName
					} else {
						continue
					}
					
			}

		}

	}	
	return tokens
}
