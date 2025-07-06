<p align="center">
  <img src="https://github.com/user-attachments/assets/773487c2-743d-4fc4-8d6e-7e3d9cb5ee21" alt="centered image">
</p>

**vibeGraphQL** is a minimalistic GraphQL library for Go that supports **queries**, **mutations**, and **subscriptions** with a clean and intuitive API. 
It was vibe coded using ChatGPT o3 model.

## ✨ Features

- 🔍 **Query resolvers** for fetching data  
- 🛠️ **Mutation resolvers** for updating data  
- 📡 **Subscription resolvers** for real-time updates  
- 🧵 Thread-safe in-memory data handling
- 📂 Multiple files uploader, alike apollo uploader
- 🔌 Simple HTTP handler integration (`/graphql` and `/subscriptions`)  

---

## 🚀 Getting Started

### 1. Install

```bash
go get github.com/Raezil/vibeGraphql
```

### 2. Define Your Schema and Resolvers

```go
if err := RegisterResolversFromSDL("schema.graphql"); err != nil {
	log.Fatalf("Failed to register resolvers: %v", err)
}
```

### 3. Define schema.graphql
```
type Query {
  user(id: ID!): User
  users(ids: [ID!]!): [User]
}

type Mutation {
  uploadFiles(files: [FileInput]!): [String]
  updateUser(id: ID!, name: String!, age: Int!): User
}

type Subscription {
  userUpdates: User
}

type User {
  id: String!
  name: String!
  age: Int!
}
```


### 4. Start HTTP Server

```go
http.HandleFunc("/graphql", graphql.GraphqlHandler)
http.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

log.Fatal(http.ListenAndServe(":8080", nil))
```

---

## 🧪 Full Example

Here is a full example using `vibeGraphql`:

```go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	graphql "github.com/Raezil/vibeGraphql"
)

// schemaDocument holds the parsed SDL document.
var schemaDocument *graphql.Document

// LoadSchemaSDL reads the SDL file, lexes/parses it into a Document,
// and stores it in the package-level variable.
// LoadSchemaSDL reads the SDL file and stores the parsed Document in schemaDocument.
func LoadSchemaSDL(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SDL file %q: %v", filePath, err)
	}
	lexer := graphql.NewLexer(string(data))
	parser := graphql.NewParser(lexer)
	doc := parser.ParseDocument()
	schemaDocument = doc
	return nil
}

// RegisterResolversFromSDL uses schemaDocument to register resolvers for Query, Mutation, and Subscription.
func RegisterResolversFromSDL(filePath string) error {
	if err := LoadSchemaSDL(filePath); err != nil {
		return err
	}
	if schemaDocument == nil {
		return fmt.Errorf("schemaDocument is nil")
	}

	availableResolvers := map[string]graphql.ResolverFunc{
		"user":        userResolver,
		"users":       usersResolver,
		"updateUser":  updateUserResolver,
		"uploadFiles": uploadFilesResolver,
		"userUpdates": userSubscriptionResolver,
	}

	// Helper to register resolvers for a given type.
	registerForType := func(typeName string, registerFunc func(string, graphql.ResolverFunc)) error {
		for _, def := range schemaDocument.Definitions {
			typeDef, ok := def.(*graphql.TypeDefinition)
			if !ok {
				continue
			}
			if typeDef.Name != typeName {
				continue
			}
			for _, field := range typeDef.Fields {
				if resolver, ok := availableResolvers[field.Name]; ok {
					registerFunc(field.Name, resolver)
					fmt.Printf("Registered resolver for %s.%s\n", typeName, field.Name)
				} else {
					fmt.Printf("No resolver found for %s.%s; skipping\n", typeName, field.Name)
				}
			}
		}
		return nil
	}

	if err := registerForType("Query", graphql.RegisterQueryResolver); err != nil {
		return err
	}
	if err := registerForType("Mutation", graphql.RegisterMutationResolver); err != nil {
		return err
	}
	if err := registerForType("Subscription", graphql.RegisterSubscriptionResolver); err != nil {
		return err
	}

	fmt.Println("Resolvers registered from SDL successfully.")
	return nil
}

// --- Resolver implementations and sample user data below ---

// User represents a sample user.
type User struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Friends []*User `json:"friends,omitempty"`
}

var (
	userStore = map[string]*User{
		"123": {ID: "123", Name: "John Doe", Age: 30, Friends: []*User{
			{ID: "456", Name: "Jane Smith", Age: 25, Friends: []*User{
				{ID: "789", Name: "Bob Johnson", Age: 28},
			}},
			{ID: "789", Name: "Bob Johnson", Age: 28},
		}},
		"456": {ID: "456", Name: "Jane Smith", Age: 25},
		"789": {ID: "789", Name: "Bob Johnson", Age: 28},
	}
	mu sync.Mutex
)

func userResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id argument missing or not a string")
	}
	mu.Lock()
	defer mu.Unlock()
	user, exists := userStore[id]
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	return user, nil
}

func usersResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	idsRaw, ok := args["ids"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("ids argument missing or not an array")
	}
	ids := make([]string, len(idsRaw))
	for i, v := range idsRaw {
		idStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		ids[i] = idStr
	}
	mu.Lock()
	defer mu.Unlock()
	var users []*User
	for _, id := range ids {
		if user, exists := userStore[id]; exists {
			users = append(users, user)
		}
	}
	return users, nil
}

func updateUserResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id argument missing or not a string")
	}
	newName, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name argument missing or not a string")
	}
	newAge, ok := args["age"].(int)
	if !ok {
		return nil, fmt.Errorf("age argument missing or not an int")
	}
	mu.Lock()
	defer mu.Unlock()
	user, exists := userStore[id]
	if !exists {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	user.Name = newName
	user.Age = newAge
	return user, nil
}

func userSubscriptionResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	ch := make(chan interface{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			user := userStore["123"]
			mu.Unlock()
			ch <- user
		}
	}()
	return ch, nil
}

func uploadFilesResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	rawFiles, ok := args["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("files argument not found or invalid")
	}
	targetDir := "./tmp"
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", targetDir, err)
	}
	var results []string
	for idx, raw := range rawFiles {
		fileData, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("file at index %d is invalid", idx)
		}
		filename, ok := fileData["filename"].(string)
		if !ok {
			return nil, fmt.Errorf("filename not provided for file at index %d", idx)
		}
		data, ok := fileData["data"].([]byte)
		if !ok {
			return nil, fmt.Errorf("file data not provided for file %q", filename)
		}
		filepath := fmt.Sprintf("%s/%s", targetDir, filename)
		if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to save file %q: %v", filename, err)
		}
		log.Printf("uploadFilesResolver: Received file %q with %d bytes", filename, len(data))
		results = append(results, fmt.Sprintf("Uploaded file %q (%d bytes)", filename, len(data)))
	}
	return results, nil
}

func main() {
	if err := RegisterResolversFromSDL("schema.graphql"); err != nil {
		log.Fatalf("Failed to register resolvers: %v", err)
	}

	// Use the GraphqlUploadHandler for /graphql to support file uploads.
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", graphql.GraphqlUploadHandler)
	mux.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Graceful shutdown setup.
	go func() {
		fmt.Println("GraphQL server is running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	// Listen for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting")
}
```

## 💬 Contributing

We welcome contributions! Feel free to open issues, feature requests or submit PRs.


---
