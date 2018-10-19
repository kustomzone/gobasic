// tokenizer contains our simple BASIC tokenizer.

package tokenizer

import (
	"github.com/skx/gobasic/token"
)

// Tokenizer holds our state.
type Tokenizer struct {
	// current character position
	position int

	// next character position
	readPosition int

	// current character
	ch rune

	// rune slice of input string
	characters []rune

	// The previous token.
	prevToken token.Token
}

// New returns a Tokenizer instance from the specified string input.
func New(input string) *Tokenizer {

	//
	// NOTE: We parse line-numbers by looking for:
	//
	//  1. NEWLINE
	//  2. INT
	//
	// To ensure that we can find the line-number of the first line
	// of the input-program we prefix that input with a newline.
	//
	// Hacks are us!
	//
	l := &Tokenizer{characters: []rune("\n" + input)}
	l.readChar()
	return l
}

// read one forward character
func (l *Tokenizer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken to read next token, skipping the white space.
func (l *Tokenizer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	switch l.ch {
	case rune('='):
		tok = newToken(token.ASSIGN, l.ch)
	case rune(','):
		tok = newToken(token.COMMA, l.ch)
	case rune('+'):
		tok = newToken(token.PLUS, l.ch)
	case rune('-'):
		tok = newToken(token.MINUS, l.ch)
	case rune('/'):
		tok = newToken(token.SLASH, l.ch)
	case rune('%'):
		tok = newToken(token.MOD, l.ch)
	case rune('*'):
		tok = newToken(token.ASTERISK, l.ch)
	case rune('('):
		tok = newToken(token.LBRACKET, l.ch)
	case rune(')'):
		tok = newToken(token.RBRACKET, l.ch)
	case rune('<'):
		if l.peekChar() == rune('>') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case rune('>'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case rune('"'):
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case rune('\n'):
		tok = newToken(token.NEWLINE, rune('N'))
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
		} else {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)
		}
	}
	l.readChar()

	//
	// Hack: A number that follows a newline is a line-number,
	// not an integer.
	//
	if l.prevToken.Type == token.NEWLINE && tok.Type == token.INT {
		tok.Type = token.LINENO
	}

	//
	// Store the previous token - which is used solely for our
	// line-number hack.
	//
	l.prevToken = tok

	return tok
}

// newToken is a simple helper for returning a new token.
func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// readIdentifier is designed to read an identifier (name of variable,
// function, etc).
func (l *Tokenizer) readIdentifier() string {

	id := ""

	for isIdentifier(l.peekChar()) {
		id += string(l.ch)
		l.readChar()
	}
	id += string(l.ch)
	return id
}

// skip white space
func (l *Tokenizer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// read a number, note that this only handles integers.
func (l *Tokenizer) readNumber() string {
	str := ""

	for isDigit(l.peekChar()) {
		str += string(l.ch)
		l.readChar()
	}
	str += string(l.ch)
	return str
}

// read a string, handling "\t", "\n", etc.
func (l *Tokenizer) readString() string {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}

		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {
			l.readChar()

			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)
	}

	return out
}

// peek character looks at the next character which is available for consumption
func (l *Tokenizer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate ch is identifier or not
func isIdentifier(ch rune) bool {
	return !isDigit(ch) && !isWhitespace(ch) && !isBrace(ch) && !isOperator(ch) && !isComparison(ch) && !isCompound(ch) && !isBrace(ch) && !isParen(ch) && !isBracket(ch) && !isEmpty(ch) && (ch != rune('\n'))
}

// is white space: note that a newline is NOT considered whitespace
// as we need that in our evaluator.
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\r')
}

// is operators
func isOperator(ch rune) bool {
	return ch == rune('+') || ch == rune('-') || ch == rune('/') || ch == rune('*')
}

// is comparison
func isComparison(ch rune) bool {
	return ch == rune('=') || ch == rune('!') || ch == rune('>') || ch == rune('<')
}

// is compound
func isCompound(ch rune) bool {
	return ch == rune(',') || ch == rune(':') || ch == rune('"') || ch == rune(';')
}

// is brace
func isBrace(ch rune) bool {
	return ch == rune('{') || ch == rune('}')
}

// is bracket
func isBracket(ch rune) bool {
	return ch == rune('[') || ch == rune(']')
}

// is parenthesis
func isParen(ch rune) bool {
	return ch == rune('(') || ch == rune(')')
}

// is empty
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}