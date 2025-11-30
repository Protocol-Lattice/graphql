<p align="center">
  <img src="https://github.com/user-attachments/assets/773487c2-743d-4fc4-8d6e-7e3d9cb5ee21" alt="centered image">
</p>

# Protocol Lattice GraphQL

**graphql** is a lightweight, modular GraphQL library for Go that supports **queries**, **mutations**, and **subscriptions** with a clean and intuitive API. It is designed to be simple to use while providing a robust foundation for building GraphQL servers.

## ✨ Features

- 🔍 **Query resolvers** for fetching data
- 🛠️ **Mutation resolvers** for updating data
- 📡 **Subscription resolvers** for real-time updates (WebSocket)
- 🧵 **Thread-safe** in-memory data handling
- 📂 **File Uploads** support (multipart/form-data)
- 🔌 **Simple HTTP handler integration** (`/graphql` and `/subscriptions`)
- 🧩 **Modular Architecture** for better maintainability and extensibility

## 🚀 Getting Started

### 1. Install

```bash
go get github.com/Protocol-Lattice/graphql
```

### 2. Define Your Schema and Resolvers

You can register resolvers programmatically using the `graphql` package facade.

```go
import (
    "github.com/Protocol-Lattice/graphql"
)

// Register a simple query resolver
graphql.RegisterQueryResolver("hello", func(source interface{}, args map[string]interface{}) (interface{}, error) {
    return "Hello, World!", nil
})
```

### 3. Start HTTP Server

Use the provided handlers to serve your GraphQL API.

```go
import (
    "net/http"
    "log"
    "github.com/Protocol-Lattice/graphql"
)

func main() {
    // Standard GraphQL handler (supports queries, mutations, and file uploads)
    http.HandleFunc("/graphql", graphql.GraphqlUploadHandler)

    // Subscription handler (WebSocket)
    http.HandleFunc("/subscriptions", graphql.SubscriptionHandler)

    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 🏗️ Modular Architecture

This project follows a clean, modular architecture:

- **`token/`**: Token definitions and constants.
- **`lexer/`**: Lexical analysis of GraphQL queries.
- **`ast/`**: Abstract Syntax Tree node definitions.
- **`parser/`**: Parsing logic to convert tokens into AST.
- **`executor/`**: Core execution logic for queries, mutations, and subscriptions.
- **`handler/`**: HTTP and WebSocket handlers.
- **`registry/`**: Centralized registry for resolvers.
- **`graphql.go`**: Public API facade for easy integration.

## 🧪 Examples

Check out the [examples/server](examples/server) directory for a complete, runnable example that demonstrates:
- Loading schema from SDL
- Implementing Query, Mutation, and Subscription resolvers
- Handling file uploads
- Setting up the HTTP server

## 💬 Contributing

We welcome contributions! Feel free to open issues, feature requests, or submit PRs.

## 📄 License

This project is licensed under the MIT License.
