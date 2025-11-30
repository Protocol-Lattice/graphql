package token

// TokenType represents the type of a token in the GraphQL lexer.
type TokenType string

const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL" // Unknown token
	EOF     TokenType = "EOF"     // End of file

	// Identifiers and literals
	IDENT  TokenType = "IDENT"  // Identifiers (field names, type names, etc.)
	INT    TokenType = "INT"    // Integer literals
	STRING TokenType = "STRING" // String literals

	// Symbols
	ASSIGN    TokenType = "="  // Assignment operator
	COLON     TokenType = ":"  // Colon separator
	COMMA     TokenType = ","  // Comma separator
	SEMICOLON TokenType = ";"  // Semicolon separator
	LPAREN    TokenType = "("  // Left parenthesis
	RPAREN    TokenType = ")"  // Right parenthesis
	LBRACE    TokenType = "{"  // Left brace
	RBRACE    TokenType = "}"  // Right brace
	LBRACKET  TokenType = "["  // Left bracket
	RBRACKET  TokenType = "]"  // Right bracket

	// GraphQL extras
	DOLLAR TokenType = "$" // Variable prefix
	BANG   TokenType = "!" // Non-null marker
)

// Token represents a single token in the GraphQL source.
type Token struct {
	Type    TokenType // The type of the token
	Literal string    // The literal value of the token
}
