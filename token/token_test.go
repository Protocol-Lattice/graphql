package token

import "testing"

func TestTokenLiterals(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		literal   string
	}{
		{ILLEGAL, "ILLEGAL"},
		{EOF, "EOF"},
		{IDENT, "IDENT"},
		{INT, "INT"},
		{STRING, "STRING"},
		{ASSIGN, "="},
		{COLON, ":"},
		{COMMA, ","},
		{SEMICOLON, ";"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{DOLLAR, "$"},
		{BANG, "!"},
	}

	for _, tt := range tests {
		if string(tt.tokenType) != tt.literal {
			t.Errorf("expected token literal for %v to be %q, got %q", tt.tokenType, tt.literal, tt.tokenType)
		}
	}
}
