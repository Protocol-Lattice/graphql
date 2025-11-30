package graphql_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	graphql "github.com/Protocol-Lattice/graphql"
)

func TestGraphqlHandlerInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	graphql.GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestGraphqlHandlerNoDefinitions(t *testing.T) {
	payload := map[string]interface{}{
		"query": "",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	graphql.GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for empty document, got %d", resp.StatusCode)
	}
}

func TestLexerIllegalCharacter(t *testing.T) {
	input := "@"
	lexer := graphql.NewLexer(input)
	tok := lexer.NextToken()
	if tok.Type != graphql.ILLEGAL {
		t.Errorf("expected ILLEGAL token, got %s", tok.Type)
	}
}

func TestOperationDefinitionImplicitQuery(t *testing.T) {
	input := `{ hello }`
	lexer := graphql.NewLexer(input)
	parser := graphql.NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatal("expected one definition for implicit query")
	}
	op, ok := doc.Definitions[0].(*graphql.OperationDefinition)
	if !ok {
		t.Fatal("expected operation definition")
	}
	if op.Operation != "query" {
		t.Errorf("expected operation to be 'query', got %q", op.Operation)
	}
}

func TestExecutorWithRegisteredResolver(t *testing.T) {
	// Register a simple query resolver
	graphql.RegisterQueryResolver("greet", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "Hello, World!", nil
	})

	// Parse and execute a query
	input := `{ greet }`
	lexer := graphql.NewLexer(input)
	parser := graphql.NewParser(lexer)
	doc := parser.ParseDocument()

	exec := graphql.NewExecutor()
	exec.RegisterQueryResolver("greet", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "Hello, World!", nil
	})

	result, err := exec.Execute(doc, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}

	greet, ok := data["greet"].(string)
	if !ok || greet != "Hello, World!" {
		t.Errorf("expected greet to be 'Hello, World!', got %v", data["greet"])
	}
}

func TestParseVariableDefinitions(t *testing.T) {
	input := `query ($var: Int!) { hello }`
	lexer := graphql.NewLexer(input)
	parser := graphql.NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*graphql.OperationDefinition)
	if !ok {
		t.Fatal("expected an operation definition")
	}
	if len(op.VariableDefinitions) != 1 {
		t.Fatalf("expected one variable definition, got %d", len(op.VariableDefinitions))
	}
	varDef := op.VariableDefinitions[0]
	if varDef.Variable != "var" {
		t.Errorf("expected variable name 'var', got %q", varDef.Variable)
	}
	if varDef.Type.Name != "Int" {
		t.Errorf("expected type 'Int', got %q", varDef.Type.Name)
	}
	if !varDef.Type.NonNull {
		t.Errorf("expected NonNull to be true")
	}
}

func TestSubscriptionExecutor(t *testing.T) {
	exec := graphql.NewExecutor()

	// Create a simple subscription
	ch := make(chan interface{}, 1)
	ch <- "event1"
	close(ch)

	exec.RegisterSubscriptionResolver("testSub", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return ch, nil
	})

	field := &graphql.Field{Name: "testSub"}
	subCh, err := exec.ExecuteSubscription(field, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case event := <-subCh:
		if event != "event1" {
			t.Errorf("expected 'event1', got %v", event)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for subscription event")
	}
}

func TestGraphqlHandlerNilVariables(t *testing.T) {
	graphql.RegisterQueryResolver("greet", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "hi", nil
	})
	payload := map[string]interface{}{
		"query":     "{ greet }",
		"variables": nil,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	graphql.GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestLexerStringToken(t *testing.T) {
	input := `"hello world"`
	lexer := graphql.NewLexer(input)
	tok := lexer.NextToken()
	if tok.Type != graphql.STRING || tok.Literal != "hello world" {
		t.Errorf("expected string token with literal 'hello world', got Type: %s, Literal: %q", tok.Type, tok.Literal)
	}
}

func TestOperationDefinitionWithNameAndVariables(t *testing.T) {
	input := `query MyQuery($id: Int) { hello }`
	lexer := graphql.NewLexer(input)
	parser := graphql.NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*graphql.OperationDefinition)
	if !ok {
		t.Fatal("expected an operation definition")
	}
	if op.Name != "MyQuery" {
		t.Errorf("expected operation name 'MyQuery', got %q", op.Name)
	}
	if len(op.VariableDefinitions) != 1 {
		t.Errorf("expected one variable definition, got %d", len(op.VariableDefinitions))
	}
}
