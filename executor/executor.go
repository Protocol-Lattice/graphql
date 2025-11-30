package executor

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Protocol-Lattice/graphql/ast"
)

// ResolverFunc defines the function signature for all resolvers.
type ResolverFunc func(source interface{}, args map[string]interface{}) (interface{}, error)

// Executor executes GraphQL queries against registered resolvers.
type Executor struct {
	queryResolvers        map[string]ResolverFunc
	mutationResolvers     map[string]ResolverFunc
	subscriptionResolvers map[string]ResolverFunc
}

// New creates a new Executor instance.
func New() *Executor {
	return &Executor{
		queryResolvers:        make(map[string]ResolverFunc),
		mutationResolvers:     make(map[string]ResolverFunc),
		subscriptionResolvers: make(map[string]ResolverFunc),
	}
}

// RegisterQueryResolver registers a resolver for a query field.
func (e *Executor) RegisterQueryResolver(field string, resolver ResolverFunc) {
	e.queryResolvers[field] = resolver
}

// RegisterMutationResolver registers a resolver for a mutation field.
func (e *Executor) RegisterMutationResolver(field string, resolver ResolverFunc) {
	e.mutationResolvers[field] = resolver
}

// RegisterSubscriptionResolver registers a resolver for a subscription field.
func (e *Executor) RegisterSubscriptionResolver(field string, resolver ResolverFunc) {
	e.subscriptionResolvers[field] = resolver
}

// Execute processes a parsed GraphQL document and returns the result.
func (e *Executor) Execute(doc *ast.Document, variables map[string]interface{}) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	if len(doc.Definitions) == 0 {
		return response, fmt.Errorf("no definitions found")
	}
	op, ok := doc.Definitions[0].(*ast.OperationDefinition)
	if !ok {
		return response, fmt.Errorf("unsupported definition type")
	}
	data, err := e.executeSelectionSet(nil, op.SelectionSet, variables)
	if err != nil {
		return response, err
	}
	response["data"] = data
	return response, nil
}

// ExecuteSubscription executes a subscription and returns a channel of events.
func (e *Executor) ExecuteSubscription(field *ast.Field, variables map[string]interface{}) (<-chan interface{}, error) {
	if resolver, ok := e.subscriptionResolvers[field.Name]; ok {
		args := buildArgs(field, variables)
		res, err := resolver(nil, args)
		if err != nil {
			return nil, err
		}
		// Try to type assert to a read-only channel
		if ch, ok := res.(<-chan interface{}); ok {
			return ch, nil
		}
		// Otherwise, try to type assert to a bidirectional channel
		if ch, ok := res.(chan interface{}); ok {
			return (<-chan interface{})(ch), nil
		}
		return nil, fmt.Errorf("subscription resolver for field %s did not return a channel", field.Name)
	}
	return nil, fmt.Errorf("no subscription resolver found for field %s", field.Name)
}

// executeSelectionSet traverses the selection set and resolves each field.
func (e *Executor) executeSelectionSet(source interface{}, ss *ast.SelectionSet, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, sel := range ss.Selections {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		res, err := e.resolveField(source, field, variables)
		if err != nil {
			return nil, err
		}
		if field.SelectionSet != nil {
			nested, err := e.resolveNestedSelection(res, field.SelectionSet, variables)
			if err != nil {
				return nil, err
			}
			result[field.Name] = nested
		} else {
			result[field.Name] = res
		}
	}
	return result, nil
}

// resolveField looks up and executes the appropriate resolver for a field.
func (e *Executor) resolveField(source interface{}, field *ast.Field, variables map[string]interface{}) (interface{}, error) {
	// At the top level, source is nil, so try both query and mutation resolvers
	if source == nil {
		// First, try the query resolver
		if resolver, ok := e.queryResolvers[field.Name]; ok {
			args := buildArgs(field, variables)
			return resolver(source, args)
		}
		// Next, try the mutation resolver
		if resolver, ok := e.mutationResolvers[field.Name]; ok {
			args := buildArgs(field, variables)
			return resolver(source, args)
		}
	}

	// If the source is not nil, use reflection to resolve nested fields
	if source != nil {
		return reflectResolve(source, field)
	}

	return nil, fmt.Errorf("no resolver found for field %s", field.Name)
}

// reflectResolve uses reflection to find a field value on a source struct.
func reflectResolve(source interface{}, field *ast.Field) (interface{}, error) {
	val := reflect.ValueOf(source)
	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("source is nil")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source is not a struct")
	}

	typ := val.Type()
	// Loop through all fields
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		// Check if the field name matches (case-insensitive)
		if strings.EqualFold(sf.Name, field.Name) {
			return val.Field(i).Interface(), nil
		}
		// Also check the "json" tag if present
		if tag, ok := sf.Tag.Lookup("json"); ok {
			tagName := strings.Split(tag, ",")[0]
			if strings.EqualFold(tagName, field.Name) {
				return val.Field(i).Interface(), nil
			}
		}
	}

	return nil, fmt.Errorf("no resolver found for field %s via reflection", field.Name)
}

// resolveNestedSelection handles nested selection sets for both objects and slices.
func (e *Executor) resolveNestedSelection(res interface{}, ss *ast.SelectionSet, variables map[string]interface{}) (interface{}, error) {
	val := reflect.ValueOf(res)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return res, nil
		}
		if val.Elem().Kind() == reflect.Struct {
			return e.executeSelectionSet(res, ss, variables)
		}
	case reflect.Struct:
		return e.executeSelectionSet(res, ss, variables)
	case reflect.Slice:
		var arr []interface{}
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			sub, err := e.executeSelectionSet(item, ss, variables)
			if err != nil {
				return nil, err
			}
			arr = append(arr, sub)
		}
		return arr, nil
	}
	return res, nil
}

// buildArgs constructs a map of argument names to values.
func buildArgs(field *ast.Field, variables map[string]interface{}) map[string]interface{} {
	args := make(map[string]interface{})
	for _, arg := range field.Arguments {
		args[arg.Name] = buildValue(arg.Value, variables)
	}
	return args
}

// buildValue converts an AST Value to a Go value.
func buildValue(val *ast.Value, variables map[string]interface{}) interface{} {
	switch val.Kind {
	case "Variable":
		if v, ok := variables[val.Literal]; ok {
			return v
		}
		return nil
	case "Int":
		i, err := strconv.Atoi(val.Literal)
		if err != nil {
			return 0
		}
		return i
	case "String":
		return val.Literal
	case "Boolean":
		return val.Literal == "true"
	case "Object":
		m := make(map[string]interface{})
		for key, fieldVal := range val.ObjectFields {
			m[key] = buildValue(fieldVal, variables)
		}
		return m
	case "Array":
		arr := []interface{}{}
		for _, elem := range val.List {
			arr = append(arr, buildValue(elem, variables))
		}
		return arr
	default:
		return val.Literal
	}
}
