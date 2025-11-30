// Package vibeGraphql provides a lightweight GraphQL implementation for Go.
// It includes lexing, parsing, execution, and HTTP handler support.
package graphql

import (
	"github.com/Protocol-Lattice/graphql/ast"
	"github.com/Protocol-Lattice/graphql/executor"
	"github.com/Protocol-Lattice/graphql/handler"
	"github.com/Protocol-Lattice/graphql/lexer"
	"github.com/Protocol-Lattice/graphql/parser"
	"github.com/Protocol-Lattice/graphql/registry"
	"github.com/Protocol-Lattice/graphql/token"
)

// ===========================
// Re-exported Types
// ===========================

// Token types
type (
	TokenType = token.TokenType
	Token     = token.Token
)

// Token constants
const (
	ILLEGAL   = token.ILLEGAL
	EOF       = token.EOF
	IDENT     = token.IDENT
	INT       = token.INT
	STRING    = token.STRING
	ASSIGN    = token.ASSIGN
	COLON     = token.COLON
	COMMA     = token.COMMA
	SEMICOLON = token.SEMICOLON
	LPAREN    = token.LPAREN
	RPAREN    = token.RPAREN
	LBRACE    = token.LBRACE
	RBRACE    = token.RBRACE
	LBRACKET  = token.LBRACKET
	RBRACKET  = token.RBRACKET
	DOLLAR    = token.DOLLAR
	BANG      = token.BANG
)

// AST types
type (
	Node                = ast.Node
	Document            = ast.Document
	Definition          = ast.Definition
	OperationDefinition = ast.OperationDefinition
	VariableDefinition  = ast.VariableDefinition
	Type                = ast.Type
	SelectionSet        = ast.SelectionSet
	Selection           = ast.Selection
	Field               = ast.Field
	Argument            = ast.Argument
	Value               = ast.Value
	TypeDefinition      = ast.TypeDefinition
)

// Executor types
type (
	ResolverFunc = executor.ResolverFunc
	Executor     = executor.Executor
)

// Lexer type
type Lexer = lexer.Lexer

// Parser type
type Parser = parser.Parser

// ===========================
// Convenience Functions
// ===========================

// NewLexer creates a new lexer for the given GraphQL source.
func NewLexer(input string) *Lexer {
	return lexer.New(input)
}

// NewParser creates a new parser for the given lexer.
func NewParser(l *Lexer) *Parser {
	return parser.New(l)
}

// NewExecutor creates a new executor instance.
func NewExecutor() *Executor {
	return executor.New()
}

// ===========================
// Global Registry Functions
// ===========================

// RegisterQueryResolver registers a query resolver in the global registry.
func RegisterQueryResolver(field string, resolver ResolverFunc) {
	registry.RegisterQueryResolver(field, resolver)
}

// RegisterMutationResolver registers a mutation resolver in the global registry.
func RegisterMutationResolver(field string, resolver ResolverFunc) {
	registry.RegisterMutationResolver(field, resolver)
}

// RegisterSubscriptionResolver registers a subscription resolver in the global registry.
func RegisterSubscriptionResolver(field string, resolver ResolverFunc) {
	registry.RegisterSubscriptionResolver(field, resolver)
}

// ===========================
// HTTP Handlers
// ===========================

// GraphqlHandler handles standard GraphQL HTTP requests.
// For backward compatibility with existing code.
var GraphqlHandler = handler.GraphQL

// GraphqlUploadHandler handles GraphQL requests with file upload support.
var GraphqlUploadHandler = handler.Upload

// SubscriptionHandler handles GraphQL subscriptions over WebSocket.
var SubscriptionHandler = handler.Subscription
