package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Protocol-Lattice/graphql/ast"
	"github.com/Protocol-Lattice/graphql/lexer"
	"github.com/Protocol-Lattice/graphql/parser"
	"github.com/Protocol-Lattice/graphql/registry"
	"github.com/gorilla/websocket"
)

// GraphQLRequest represents a standard GraphQL request.
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GraphQL handles standard GraphQL HTTP requests.
func GraphQL(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req GraphQLRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Variables == nil {
		req.Variables = make(map[string]interface{})
	}

	// Lex and parse the query
	l := lexer.New(req.Query)
	p := parser.New(l)
	doc := p.ParseDocument()

	// Execute the query using the global executor
	exec := registry.GetGlobalExecutor()
	result, err := exec.Execute(doc, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the JSON result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Subscription handles GraphQL subscriptions over WebSocket.
func Subscription(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "unable to upgrade to websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Read the subscription request from the WebSocket
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("failed to read subscription message"))
		return
	}

	var req GraphQLRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("invalid subscription JSON"))
		return
	}

	// Lex, parse, and extract the subscription operation
	l := lexer.New(req.Query)
	p := parser.New(l)
	doc := p.ParseDocument()

	if len(doc.Definitions) == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("no subscription definition found"))
		return
	}

	op, ok := doc.Definitions[0].(*ast.OperationDefinition)
	if !ok || op.Operation != "subscription" {
		conn.WriteMessage(websocket.TextMessage, []byte("provided operation is not a subscription"))
		return
	}

	if len(op.SelectionSet.Selections) == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("subscription selection set is empty"))
		return
	}

	field, ok := op.SelectionSet.Selections[0].(*ast.Field)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("invalid subscription field"))
		return
	}

	// Execute the subscription
	exec := registry.GetGlobalExecutor()
	subCh, err := exec.ExecuteSubscription(field, req.Variables)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("subscription error: %v", err)))
		return
	}

	// Stream events from the subscription channel to the WebSocket
	for event := range subCh {
		if err := conn.WriteJSON(event); err != nil {
			fmt.Printf("failed to write event: %v\n", err)
			break
		}
	}
}

// Upload handles GraphQL requests with file uploads (multipart/form-data).
func Upload(w http.ResponseWriter, r *http.Request) {
	// If not multipart, delegate to regular GraphQL handler
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		GraphQL(w, r)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	operations := r.FormValue("operations")
	if operations == "" {
		http.Error(w, "missing operations field", http.StatusBadRequest)
		return
	}

	var req GraphQLRequest
	if err := json.Unmarshal([]byte(operations), &req); err != nil {
		http.Error(w, "invalid operations JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Variables == nil {
		req.Variables = make(map[string]interface{})
	}

	fileMapStr := r.FormValue("map")
	if fileMapStr == "" {
		http.Error(w, "missing map field", http.StatusBadRequest)
		return
	}

	var fileMap map[string][]string
	if err := json.Unmarshal([]byte(fileMapStr), &fileMap); err != nil {
		http.Error(w, "invalid map JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	var varMu sync.Mutex

	for fileKey, paths := range fileMap {
		wg.Add(1)
		go func(fileKey string, paths []string) {
			defer wg.Done()
			file, header, err := r.FormFile(fileKey)
			if err != nil {
				log.Printf("failed to retrieve file %s: %v", fileKey, err)
				return
			}
			defer file.Close()
			fileData, err := ioutil.ReadAll(file)
			if err != nil {
				log.Printf("failed to read file %s: %v", header.Filename, err)
				return
			}
			log.Printf("Uploaded file %q with %d bytes", header.Filename, len(fileData))
			for _, path := range paths {
				// Remove the "variables." prefix if present
				adjustedPath := strings.TrimPrefix(path, "variables.")
				varMu.Lock()
				// If the path contains a dot and the second part is numeric, update as an array
				parts := strings.Split(adjustedPath, ".")
				if len(parts) == 2 {
					if _, err := strconv.Atoi(parts[1]); err == nil {
						setNestedArrayValue(req.Variables, adjustedPath, map[string]interface{}{
							"filename": header.Filename,
							"data":     fileData,
						})
						varMu.Unlock()
						continue
					}
				}
				setNestedValue(req.Variables, adjustedPath, map[string]interface{}{
					"filename": header.Filename,
					"data":     fileData,
				})
				varMu.Unlock()
			}
		}(fileKey, paths)
	}
	wg.Wait()

	// Continue processing the GraphQL query
	l := lexer.New(req.Query)
	p := parser.New(l)
	doc := p.ParseDocument()

	exec := registry.GetGlobalExecutor()
	result, err := exec.Execute(doc, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// setNestedValue updates nested maps (non-array paths).
func setNestedValue(vars map[string]interface{}, path string, value interface{}) {
	keys := strings.Split(path, ".")
	current := vars
	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
			return
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
}

// setNestedArrayValue updates an array element given a path like "files.0".
func setNestedArrayValue(vars map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	if len(parts) != 2 {
		setNestedValue(vars, path, value)
		return
	}
	arrayKey := parts[0]
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		setNestedValue(vars, path, value)
		return
	}
	if arr, ok := vars[arrayKey].([]interface{}); ok {
		// Extend the slice if needed
		if idx >= len(arr) {
			newArr := make([]interface{}, idx+1)
			copy(newArr, arr)
			arr = newArr
			vars[arrayKey] = arr
		}
		arr[idx] = value
	} else {
		// If not an array, create one
		newArr := make([]interface{}, idx+1)
		newArr[idx] = value
		vars[arrayKey] = newArr
	}
}
