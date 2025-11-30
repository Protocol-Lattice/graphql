package lexer

import (
	"testing"

	"github.com/Protocol-Lattice/graphql/token"
)

func TestLexer_Numbers(t *testing.T) {
	input := "12345 67890"
	lexer := New(input)

	// First number.
	tok := lexer.NextToken()
	if tok.Type != token.INT {
		t.Fatalf("expected token type INT, got %s", tok.Type)
	}
	if tok.Literal != "12345" {
		t.Errorf("expected literal '12345', got %q", tok.Literal)
	}

	// Second number.
	tok = lexer.NextToken()
	if tok.Type != token.INT {
		t.Fatalf("expected token type INT, got %s", tok.Type)
	}
	if tok.Literal != "67890" {
		t.Errorf("expected literal '67890', got %q", tok.Literal)
	}

	// End of input.
	tok = lexer.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestLexer_Strings(t *testing.T) {
	input := `"hello world" "another string"`
	lexer := New(input)

	// First string.
	tok := lexer.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected token type STRING, got %s", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected literal 'hello world', got %q", tok.Literal)
	}

	// Second string.
	tok = lexer.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected token type STRING, got %s", tok.Type)
	}
	if tok.Literal != "another string" {
		t.Errorf("expected literal 'another string', got %q", tok.Literal)
	}

	// End of input.
	tok = lexer.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestLexer_IllegalCharacter(t *testing.T) {
	input := "@"
	lexer := New(input)

	tok := lexer.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected token type ILLEGAL, got %s", tok.Type)
	}
	if tok.Literal != "@" {
		t.Errorf("expected literal '@', got %q", tok.Literal)
	}

	tok = lexer.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}
