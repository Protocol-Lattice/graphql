package parser

import (
	"github.com/Protocol-Lattice/graphql/ast"
	"github.com/Protocol-Lattice/graphql/lexer"
	"github.com/Protocol-Lattice/graphql/token"
)

// Parser parses GraphQL source code into an AST.
type Parser struct {
	l         *lexer.Lexer // The lexer to read tokens from
	curToken  token.Token  // Current token
	peekToken token.Token  // Next token
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Initialize two tokens
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances the parser to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseDocument parses a GraphQL document.
func (p *Parser) ParseDocument() *ast.Document {
	doc := &ast.Document{}
	for p.curToken.Type != token.EOF {
		def := p.parseDefinition()
		if def != nil {
			doc.Definitions = append(doc.Definitions, def)
		}
	}
	return doc
}

// parseDefinition parses a single definition (operation or type).
func (p *Parser) parseDefinition() ast.Definition {
	// Handle operation definitions
	if p.curToken.Literal == "query" ||
		p.curToken.Literal == "mutation" ||
		p.curToken.Literal == "subscription" {
		return p.parseOperationDefinition()
	}
	// Handle implicit queries (starting with '{')
	if p.curToken.Type == token.LBRACE {
		return p.parseOperationDefinition()
	}
	// Handle type definitions
	if p.curToken.Literal == "type" {
		return p.skipTypeDefinition()
	}
	// Unknown definition, skip it
	p.nextToken()
	return nil
}

// parseOperationDefinition parses a query, mutation, or subscription operation.
func (p *Parser) parseOperationDefinition() *ast.OperationDefinition {
	op := &ast.OperationDefinition{}
	if p.curToken.Literal == "query" ||
		p.curToken.Literal == "mutation" ||
		p.curToken.Literal == "subscription" {
		op.Operation = p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == token.IDENT {
			op.Name = p.curToken.Literal
			p.nextToken()
		}
		if p.curToken.Type == token.LPAREN {
			op.VariableDefinitions = p.parseVariableDefinitions()
		}
	} else {
		op.Operation = "query"
	}
	if p.curToken.Type == token.LBRACE {
		op.SelectionSet = p.parseSelectionSet()
	}
	return op
}

// parseVariableDefinitions parses variable definitions for an operation.
func (p *Parser) parseVariableDefinitions() []ast.VariableDefinition {
	var vars []ast.VariableDefinition
	p.nextToken() // Skip '('
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.DOLLAR {
			p.nextToken() // Skip '$'
			if p.curToken.Type != token.IDENT {
				return vars
			}
			varDef := ast.VariableDefinition{}
			varDef.Variable = p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == token.COLON {
				p.nextToken()
				typeParsed := p.parseType()
				if typeParsed != nil {
					varDef.Type = *typeParsed
				}
			}
			vars = append(vars, varDef)
		}
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // Skip ')'
	return vars
}

// parseSelectionSet parses a selection set (fields within braces).
func (p *Parser) parseSelectionSet() *ast.SelectionSet {
	ss := &ast.SelectionSet{}
	p.nextToken() // skip '{'
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		sel := p.parseSelection()
		if sel != nil {
			ss.Selections = append(ss.Selections, sel)
		}
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip '}'
	return ss
}

// parseSelection parses a single selection (currently only fields).
func (p *Parser) parseSelection() ast.Selection {
	return p.parseField()
}

// parseField parses a field selection.
func (p *Parser) parseField() *ast.Field {
	field := &ast.Field{}
	if p.curToken.Type != token.IDENT {
		return nil
	}
	field.Name = p.curToken.Literal
	p.nextToken()
	if p.curToken.Type == token.LPAREN {
		field.Arguments = p.parseArguments()
	}
	if p.curToken.Type == token.LBRACE {
		field.SelectionSet = p.parseSelectionSet()
	}
	return field
}

// parseArguments parses field arguments.
func (p *Parser) parseArguments() []ast.Argument {
	var args []ast.Argument
	p.nextToken() // skip '('
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		arg := ast.Argument{}
		if p.curToken.Type == token.IDENT {
			arg.Name = p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == token.COLON {
				p.nextToken()
				arg.Value = p.parseValue()
			}
			args = append(args, arg)
		}
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip ')'
	return args
}

// parseValue parses a value (string, int, boolean, variable, object, array).
func (p *Parser) parseValue() *ast.Value {
	// Handle object literals
	if p.curToken.Type == token.LBRACE {
		return p.parseObject()
	}
	// Handle array literals
	if p.curToken.Type == token.LBRACKET {
		return p.parseArray()
	}

	val := &ast.Value{}
	switch p.curToken.Type {
	case token.INT:
		val.Kind = "Int"
		val.Literal = p.curToken.Literal
		p.nextToken()
	case token.STRING:
		val.Kind = "String"
		val.Literal = p.curToken.Literal
		p.nextToken()
	case token.IDENT:
		// Handle booleans and enums
		if p.curToken.Literal == "true" || p.curToken.Literal == "false" {
			val.Kind = "Boolean"
		} else {
			val.Kind = "Enum"
		}
		val.Literal = p.curToken.Literal
		p.nextToken()
	case token.DOLLAR:
		p.nextToken() // skip '$'
		if p.curToken.Type == token.IDENT {
			val.Kind = "Variable"
			val.Literal = p.curToken.Literal
			p.nextToken()
		} else {
			val.Kind = "Variable"
			val.Literal = ""
		}
	default:
		val.Kind = "Illegal"
		val.Literal = p.curToken.Literal
		p.nextToken()
	}
	return val
}

// parseObject parses a GraphQL object literal.
func (p *Parser) parseObject() *ast.Value {
	objFields := make(map[string]*ast.Value)
	p.nextToken() // Skip '{'
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type != token.IDENT {
			return &ast.Value{Kind: "Illegal", Literal: "expected object key"}
		}
		key := p.curToken.Literal
		p.nextToken()
		if p.curToken.Type != token.COLON {
			return &ast.Value{Kind: "Illegal", Literal: "expected colon in object"}
		}
		p.nextToken() // skip colon
		value := p.parseValue()
		objFields[key] = value
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // Skip '}'
	return &ast.Value{
		Kind:         "Object",
		ObjectFields: objFields,
	}
}

// parseArray parses an array of values.
func (p *Parser) parseArray() *ast.Value {
	arr := []*ast.Value{}
	p.nextToken() // skip '['
	for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
		val := p.parseValue()
		arr = append(arr, val)
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip ']'
	return &ast.Value{Kind: "Array", List: arr}
}

// parseType parses a GraphQL type (e.g., String, [Int!], User!).
func (p *Parser) parseType() *ast.Type {
	var t ast.Type
	if p.curToken.Type == token.LBRACKET {
		// List type
		p.nextToken()              // Skip '['
		innerType := p.parseType() // Recursively parse the inner type
		t = ast.Type{IsList: true, Elem: innerType}
		if p.curToken.Type != token.RBRACKET {
			// Handle error: expected closing bracket
		}
		p.nextToken() // Skip ']'
		// Check for non-null on the list type
		if p.curToken.Type == token.BANG {
			t.NonNull = true
			p.nextToken()
		}
		return &t
	} else if p.curToken.Type == token.IDENT {
		// Basic type
		t = ast.Type{Name: p.curToken.Literal}
		p.nextToken()
		// Check for non-null on the basic type
		if p.curToken.Type == token.BANG {
			t.NonNull = true
			p.nextToken()
		}
		return &t
	}
	return nil
}

// skipTypeDefinition parses and returns a type definition (e.g., "type Query { ... }").
func (p *Parser) skipTypeDefinition() ast.Definition {
	p.nextToken() // Skip "type"
	if p.curToken.Type != token.IDENT {
		return nil
	}
	typeName := p.curToken.Literal
	p.nextToken() // Move past type name

	// Expect an opening brace
	if p.curToken.Type != token.LBRACE {
		return nil
	}
	p.nextToken() // Skip '{'

	var fields []*ast.Field
	iterations := 0
	maxIterations := 10000 // safeguard
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		iterations++
		if iterations > maxIterations {
			break
		}
		field := p.parseTypeField()
		if field != nil {
			fields = append(fields, field)
		} else {
			p.nextToken()
		}
		if p.curToken.Type == token.COMMA {
			p.nextToken()
		}
	}
	if p.curToken.Type == token.RBRACE {
		p.nextToken() // Skip '}'
	}
	return &ast.TypeDefinition{
		Name:   typeName,
		Fields: fields,
	}
}

// parseTypeField parses a field in a type definition.
func (p *Parser) parseTypeField() *ast.Field {
	if p.curToken.Type != token.IDENT {
		return nil
	}
	field := &ast.Field{
		Name: p.curToken.Literal,
	}
	p.nextToken() // Consume the field name

	// If there's an argument list, skip it
	if p.curToken.Type == token.LPAREN {
		p.skipParenBlock()
	}

	// If a colon is present, skip the type annotation
	if p.curToken.Type == token.COLON {
		p.skipTypeAnnotation()
	}
	return field
}

// skipParenBlock skips over a parenthesized block.
func (p *Parser) skipParenBlock() {
	if p.curToken.Type != token.LPAREN {
		return
	}
	depth := 1
	p.nextToken() // Skip the opening '('
	for depth > 0 && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.LPAREN {
			depth++
		} else if p.curToken.Type == token.RPAREN {
			depth--
		}
		p.nextToken()
	}
}

// skipTypeAnnotation skips a type annotation (: Type).
func (p *Parser) skipTypeAnnotation() {
	if p.curToken.Type != token.COLON {
		return
	}
	p.nextToken() // Skip the colon

	// Check for list type
	if p.curToken.Type == token.LBRACKET {
		p.nextToken() // Consume '['
		if p.curToken.Type == token.IDENT {
			p.nextToken()
			if p.curToken.Type == token.BANG {
				p.nextToken()
			}
		}
		if p.curToken.Type == token.RBRACKET {
			p.nextToken()
		}
		if p.curToken.Type == token.BANG {
			p.nextToken()
		}
		return
	}

	// Simple type
	if p.curToken.Type == token.IDENT {
		p.nextToken()
	}
	if p.curToken.Type == token.BANG {
		p.nextToken()
	}
}
