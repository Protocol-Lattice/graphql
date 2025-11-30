package ast

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Document represents a complete GraphQL document.
// It contains a list of definitions (operations or type definitions).
type Document struct {
	Definitions []Definition
}

// TokenLiteral returns a string representation of the document.
func (d *Document) TokenLiteral() string {
	if len(d.Definitions) > 0 {
		return d.Definitions[0].TokenLiteral()
	}
	return ""
}

// Definition is an interface for all top-level definitions in a GraphQL document.
type Definition interface {
	Node
}

// OperationDefinition represents a GraphQL operation (query, mutation, or subscription).
type OperationDefinition struct {
	Operation           string               // "query", "mutation", or "subscription"
	Name                string               // Optional operation name
	VariableDefinitions []VariableDefinition // Variable definitions for this operation
	SelectionSet        *SelectionSet        // The fields to select
}

// TokenLiteral returns the operation name or type.
func (op *OperationDefinition) TokenLiteral() string {
	if op.Name != "" {
		return op.Name
	}
	return op.Operation
}

// VariableDefinition represents a variable definition in an operation.
type VariableDefinition struct {
	Variable string // Variable name (without $)
	Type     Type   // The type of the variable
}

// TokenLiteral returns the variable name.
func (v *VariableDefinition) TokenLiteral() string {
	return v.Variable
}

// Type represents a GraphQL type (e.g., String, [Int!], User).
type Type struct {
	Name    string // Base type name
	NonNull bool   // Whether the type is non-nullable (!)
	IsList  bool   // Whether the type is a list ([])
	Elem    *Type  // Element type if this is a list
}

// SelectionSet represents a set of fields to select.
type SelectionSet struct {
	Selections []Selection
}

// Selection is an interface for all selections (fields, fragments, etc.).
type Selection interface {
	Node
}

// Field represents a single field selection in a GraphQL query.
type Field struct {
	Name         string        // Field name
	Arguments    []Argument    // Field arguments
	SelectionSet *SelectionSet // Nested selections (if any)
}

// TokenLiteral returns the field name.
func (f *Field) TokenLiteral() string {
	return f.Name
}

// Argument represents an argument passed to a field.
type Argument struct {
	Name  string // Argument name
	Value *Value // Argument value
}

// TokenLiteral returns the argument name.
func (a *Argument) TokenLiteral() string {
	return a.Name
}

// Value represents a value in GraphQL (string, int, variable, object, array, etc.).
type Value struct {
	Kind         string            // "Int", "String", "Boolean", "Variable", "Enum", "Object", "Array"
	Literal      string            // The literal value
	ObjectFields map[string]*Value // For object values
	List         []*Value          // For array values
}

// TokenLiteral returns the literal value.
func (v *Value) TokenLiteral() string {
	return v.Literal
}

// TypeDefinition represents a type definition in a GraphQL schema (e.g., "type Query { ... }").
type TypeDefinition struct {
	Name   string   // Type name
	Fields []*Field // Fields in this type
}

// TokenLiteral returns the type name.
func (t *TypeDefinition) TokenLiteral() string {
	return t.Name
}
