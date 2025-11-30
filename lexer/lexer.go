package lexer

import (
	"unicode"

	"github.com/Protocol-Lattice/graphql/token"
)

// Lexer tokenizes GraphQL source code.
type Lexer struct {
	input        string // The input string
	position     int    // Current position in input (points to current char)
	readPosition int    // Next reading position (after current char)
	ch           byte   // Current char under examination
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar advances the lexer to the next character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII 0 signifies end-of-input
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	switch l.ch {
	case '=':
		tok = token.Token{Type: token.ASSIGN, Literal: string(l.ch)}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.ch)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.ch)}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.ch)}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.ch)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.ch)}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.ch)}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.ch)}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: string(l.ch)}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: string(l.ch)}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '$':
		tok = token.Token{Type: token.DOLLAR, Literal: string(l.ch)}
	case '!':
		tok = token.Token{Type: token.BANG, Literal: string(l.ch)}
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.IDENT
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch)}
		}
	}
	l.readChar()
	return tok
}

// skipWhitespace advances the lexer past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier from the input.
func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// readNumber reads a number from the input.
func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// readString reads a string literal from the input.
func (l *Lexer) readString() string {
	// skip opening quote
	l.readChar()
	start := l.position
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	str := l.input[start:l.position]
	// skip closing quote
	l.readChar()
	return str
}

// isLetter checks if a byte is a letter or underscore.
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isDigit checks if a byte is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
