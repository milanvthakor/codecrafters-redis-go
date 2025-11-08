# Redis Server Refactoring Documentation

## Overview

This document explains the refactoring of the Redis server codebase to make it more modular, scalable, readable, and maintainable.

## Why Refactor?

### Problems with the Original Structure

1. **Monolithic main.go**: All command handlers, connection logic, and business logic were in a single 462-line file
2. **Tight Coupling**: Command handlers were directly coupled to the main package
3. **Hard to Test**: No clear separation of concerns made unit testing difficult
4. **Hard to Extend**: Adding new commands required modifying the main switch statement
5. **Mixed Concerns**: Protocol handling, storage, and command logic were all mixed together

### Benefits of the New Structure

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Modularity**: Components can be developed, tested, and maintained independently
3. **Scalability**: Easy to add new commands, storage backends, or protocol handlers
4. **Testability**: Each component can be unit tested in isolation
5. **Readability**: Code is organized logically, making it easier to understand
6. **Maintainability**: Changes to one component don't affect others

## Project Structure

```
app/
├── main.go                    # Entry point - minimal, just starts the server
└── internal/                  # Internal packages (not exported outside module)
    ├── cmd/                   # Command handlers
    │   ├── handler.go         # Command registry and execution
    │   ├── basic.go           # Basic commands (PING, ECHO)
    │   ├── string.go          # String commands (SET, GET, INCR)
    │   ├── list.go            # List commands (RPUSH, LPUSH, LRANGE, etc.)
    │   └── stream.go          # Stream commands (XADD, XRANGE, XREAD)
    ├── server/                # Server and connection handling
    │   └── server.go          # Server instance and connection management
    ├── storage/               # Storage layer
    │   ├── mem.go             # In-memory storage implementation
    │   └── stream_utils.go    # Stream utility functions
    ├── protocol/              # RESP protocol implementation
    │   └── resp.go            # RESP encoding/decoding
    └── errors/                # Error definitions
        └── errors.go          # Centralized error definitions
```

## Package Responsibilities

### `main` Package
- **Purpose**: Application entry point
- **Responsibilities**:
  - Parse command-line flags
  - Create server instance
  - Start listening for connections
  - Delegate connection handling to server

### `internal/cmd` Package
- **Purpose**: Command handling and execution
- **Responsibilities**:
  - Define command handler interface
  - Register all command handlers
  - Execute commands via registry pattern
  - Organize commands by category (basic, string, list, stream)

**Key Design Pattern**: **Registry Pattern**
- Commands are registered in a map
- Easy to add new commands without modifying existing code
- Commands can be looked up and executed dynamically

### `internal/server` Package
- **Purpose**: Server lifecycle and connection management
- **Responsibilities**:
  - Handle client connections
  - Manage transaction state (MULTI/EXEC/DISCARD)
  - Coordinate between protocol, commands, and storage
  - Handle connection lifecycle

### `internal/storage` Package
- **Purpose**: Data storage abstraction
- **Responsibilities**:
  - In-memory data structures (strings, lists, streams)
  - Thread-safe operations
  - Blocking operations (BLPOP, XREAD with timeout)
  - Stream ID validation and management

### `internal/protocol` Package
- **Purpose**: RESP protocol implementation
- **Responsibilities**:
  - Encode/decode RESP messages
  - Handle all RESP data types
  - Network I/O operations

### `internal/errors` Package
- **Purpose**: Centralized error definitions
- **Responsibilities**:
  - Define all application errors
  - Provide consistent error messages
  - Make errors reusable across packages

## Design Patterns Used

### 1. Registry Pattern
Commands are registered in a map, allowing dynamic lookup and execution:
```go
registry := cmd.NewRegistry()
registry.Register("SET", handleSet)
resp := registry.Execute("SET", cmd, store)
```

**Benefits**:
- Easy to add new commands
- No large switch statements
- Commands can be registered at runtime

### 2. Dependency Injection
Server receives dependencies (storage, registry) rather than creating them:
```go
type Server struct {
    store    *storage.Mem
    registry *cmd.Registry
}
```

**Benefits**:
- Easy to test (can inject mocks)
- Flexible (can swap implementations)
- Clear dependencies

### 3. Separation of Concerns
Each package has a single responsibility:
- Protocol package: Only handles RESP protocol
- Storage package: Only handles data storage
- Command package: Only handles command execution
- Server package: Only handles connection management

## How to Add a New Command

1. **Create handler function** in appropriate file (e.g., `cmd/string.go`):
```go
func handleNewCmd(cmd []*protocol.RespVal, store *storage.Mem) (string, error) {
    // Implementation
    return protocol.ToSimpleStr("OK"), nil
}
```

2. **Register the command** in `cmd/handler.go`:
```go
r.Register("NEWCMD", handleNewCmd)
```

That's it! No need to modify switch statements or main.go.

## Migration Notes

### Old Files (Can Be Removed)
- `app/resp.go` → Moved to `internal/protocol/resp.go`
- `app/mem.go` → Moved to `internal/storage/mem.go`
- `app/utils.go` → Moved to `internal/storage/stream_utils.go`
- `app/connection.go` → Functionality moved to `internal/server/server.go`

### Breaking Changes
- All internal packages are in `internal/` directory, so they're not exported outside the module
- This is intentional - it prevents external code from depending on internal implementation details

## Testing Strategy

With this structure, you can now:

1. **Unit test command handlers** independently
2. **Mock storage layer** for testing commands
3. **Test protocol encoding/decoding** separately
4. **Integration test** the full server

Example test structure:
```
app/
└── internal/
    ├── cmd/
    │   └── string_test.go
    ├── storage/
    │   └── mem_test.go
    └── protocol/
        └── resp_test.go
```

## Performance Considerations

- **No performance impact**: The refactoring maintains the same execution flow
- **Minimal overhead**: Function calls are inlined by the Go compiler
- **Same concurrency model**: Still uses goroutines for each connection

## Future Enhancements

With this structure, you can easily:

1. **Add new storage backends** (Redis-compatible, disk-based, etc.)
2. **Implement command middleware** (logging, metrics, auth)
3. **Add command aliases** (e.g., `DEL` → `DELETE`)
4. **Implement command pipelining** optimizations
5. **Add plugin system** for custom commands
6. **Implement replication** by abstracting storage layer

## Conclusion

This refactoring transforms a monolithic codebase into a well-organized, modular system that follows Go best practices and design patterns. The code is now:

- ✅ **Modular**: Clear separation of concerns
- ✅ **Scalable**: Easy to add new features
- ✅ **Readable**: Logical organization
- ✅ **Maintainable**: Changes are isolated
- ✅ **Testable**: Components can be tested independently

The structure follows the same patterns used in production Go applications and makes the codebase ready for future growth.

