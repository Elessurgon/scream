package lexer

import (
	"fmt"
	"strings"
	"unicode"

	"scream/token"
)

type Lexer struct {
	position int

	readPosition int

	ch rune

	characters []rune

	prevToken token.Token
}

func New(input string) *Lexer {
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

func (l *Lexer) GetLine() int {
	line := 0
	chars := len(l.characters)
	i := 0

	for i < l.readPosition && i < chars {

		if l.characters[i] == rune('\n') {
			line++
		}

		i++
	}
	return line
}
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	if l.ch == rune('#') ||
		(l.ch == rune('/') && l.peekChar() == rune('/')) {
		l.skipComment()
		return (l.NextToken())
	}

	if l.ch == rune('/') && l.peekChar() == rune('*') {
		l.skipMultiLineComment()
	}

	switch l.ch {
	case rune('&'):
		if l.peekChar() == rune('&') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch)}
		}
	case rune('|'):
		if l.peekChar() == rune('|') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch)}
		}

	case rune('='):
		tok = newToken(token.ASSIGN, l.ch)
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case rune(';'):
		tok = newToken(token.SEMICOLON, l.ch)
	case rune('?'):
		tok = newToken(token.QUESTION, l.ch)
	case rune('('):
		tok = newToken(token.LPAREN, l.ch)
	case rune(')'):
		tok = newToken(token.RPAREN, l.ch)
	case rune(','):
		tok = newToken(token.COMMA, l.ch)
	case rune('.'):
		if l.peekChar() == rune('.') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.DOTDOT, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.PERIOD, l.ch)
		}
	case rune('+'):
		if l.peekChar() == rune('+') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.PLUS_PLUS, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.PLUS_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.PLUS, l.ch)
		}
	case rune('%'):
		tok = newToken(token.MOD, l.ch)
	case rune('{'):
		tok = newToken(token.LBRACE, l.ch)
	case rune('}'):
		tok = newToken(token.RBRACE, l.ch)
	case rune('-'):
		if l.peekChar() == rune('-') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MINUS_MINUS, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MINUS_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case rune('/'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.SLASH_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			if l.prevToken.Type == token.RBRACKET ||
				l.prevToken.Type == token.RPAREN ||
				l.prevToken.Type == token.IDENT ||
				l.prevToken.Type == token.INT ||
				l.prevToken.Type == token.FLOAT {

				tok = newToken(token.SLASH, l.ch)
			} else {
				str, err := l.readRegexp()
				if err == nil {
					tok.Type = token.REGEXP
					tok.Literal = str
				} else {
					fmt.Printf("%s\n", err.Error())
					tok.Type = token.REGEXP
					tok.Literal = str
				}
			}
		}
	case rune('*'):
		if l.peekChar() == rune('*') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.POW, Literal: string(ch) + string(l.ch)}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASTERISK_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case rune('<'):
		if l.peekChar() == rune('=') {
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
	case rune('~'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.CONTAINS, Literal: string(ch) + string(l.ch)}
		}

	case rune('!'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			if l.peekChar() == rune('~') {
				ch := l.ch
				l.readChar()
				tok = token.Token{Type: token.NOT_CONTAINS, Literal: string(ch) + string(l.ch)}

			} else {
				tok = newToken(token.BANG, l.ch)
			}
		}
	case rune('"'):
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case rune('`'):
		tok.Type = token.BACKTICK
		tok.Literal = l.readBacktick()
	case rune('['):
		tok = newToken(token.LBRACKET, l.ch)
	case rune(']'):
		tok = newToken(token.RBRACKET, l.ch)
	case rune(':'):
		tok = newToken(token.COLON, l.ch)
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:

		if isDigit(l.ch) {
			tok = l.readDecimal()
			l.prevToken = tok
			return tok

		}

		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdentifier(tok.Literal)
		l.prevToken = tok
		return tok
	}
	l.readChar()
	l.prevToken = tok
	return tok
}

func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readIdentifier() string {

	valid := map[string]bool{
		"directory.glob":     true,
		"math.abs":           true,
		"math.random":        true,
		"math.sqrt":          true,
		"os.environment":     true,
		"os.getenv":          true,
		"os.setenv":          true,
		"string.interpolate": true,
	}

	types := []string{"string.",
		"array.",
		"integer.",
		"float.",
		"hash.",
		"object."}

	id := ""

	position := l.position
	rposition := l.readPosition

	for isIdentifier(l.ch) {
		id += string(l.ch)
		l.readChar()
	}

	if strings.Contains(id, ".") {

		ok := valid[id]

		if !ok {
			for _, i := range types {
				if strings.HasPrefix(id, i) {
					ok = true
				}
			}
		}

		if !ok {

			offset := strings.Index(id, ".")
			id = id[:offset]

			l.position = position
			l.readPosition = rposition
			for offset > 0 {
				l.readChar()
				offset--
			}
		}
	}

	return id
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

func (l *Lexer) skipMultiLineComment() {
	found := false

	for !found {
		if l.ch == rune(0) {
			found = true
		}
		if l.ch == '*' && l.peekChar() == '/' {
			found = true

			l.readChar()
		}

		l.readChar()
	}

	l.skipWhitespace()
}

func (l *Lexer) readNumber() string {
	str := ""

	accept := "0123456789"

	if l.ch == '0' && l.peekChar() == 'x' {
		accept = "0x123456789abcdefABCDEF"
	}

	if l.ch == '0' && l.peekChar() == 'b' {
		accept = "b01"
	}

	for strings.Contains(accept, string(l.ch)) {
		str += string(l.ch)
		l.readChar()
	}
	return str
}
func (l *Lexer) readDecimal() token.Token {

	integer := l.readNumber()
	if l.ch == rune('.') && isDigit(l.peekChar()) {
		l.readChar()
		fraction := l.readNumber()
		return token.Token{Type: token.FLOAT, Literal: integer + "." + fraction}
	}
	return token.Token{Type: token.INT, Literal: integer}
}

func (l *Lexer) readString() string {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}

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

func (l *Lexer) readRegexp() (string, error) {
	out := ""

	for {
		l.readChar()

		if l.ch == rune(0) {
			return "unterminated regular expression", fmt.Errorf("unterminated regular expression")
		}
		if l.ch == '/' {

			l.readChar()

			flags := ""

			for l.ch == rune('i') || l.ch == rune('m') {

				if !strings.Contains(flags, string(l.ch)) {

					tmp := strings.Split(flags, "")
					tmp = append(tmp, string(l.ch))
					flags = strings.Join(tmp, "")

				}

				l.readChar()
			}

			if len(flags) > 0 {
				out = "(?" + flags + ")" + out
			}
			break
		}
		out = out + string(l.ch)
	}

	return out, nil
}
func (l *Lexer) readBacktick() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '`' {
			break
		}
	}
	out := string(l.characters[position:l.position])
	return out
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

func isIdentifier(ch rune) bool {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '?' || ch == '$' || ch == '_' {
		return true
	}

	return false
}

func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
