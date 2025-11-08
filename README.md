# Gokv

A Redis-compatible in-memory data store server built with Go. This implementation supports core Redis commands including string operations (SET, GET, INCR), list operations (RPUSH, LPUSH, LRANGE, LPOP, BLPOP), stream operations (XADD, XRANGE, XREAD), and transaction support (MULTI, EXEC, DISCARD).

The server uses the RESP (REdis Serialization Protocol) for client-server communication and provides thread-safe in-memory storage with support for blocking operations and stream data structures.

## Features

- **Thread-safe operations**: All storage operations are protected with read-write mutexes for concurrent access
- **Blocking operations**: Support for blocking list operations (BLPOP) with timeout handling
- **Stream data structures**: Full support for Redis streams with XADD, XRANGE, and XREAD commands
- **Transaction support**: MULTI/EXEC/DISCARD commands for atomic command execution
- **Key expiration**: Automatic key expiration with configurable time-to-live (TTL)
- **RESP protocol**: Full RESP (REdis Serialization Protocol) implementation for Redis compatibility
- **Concurrent connections**: Handles multiple client connections simultaneously using goroutines

## Supported Commands

### Basic Commands
- `PING` - Returns PONG, used for connection testing
- `ECHO <message>` - Echoes the message back to the client

### String Commands
- `SET <key> <value> [EX seconds|PX milliseconds]` - Set a key-value pair with optional expiration
- `GET <key>` - Get the value associated with a key
- `INCR <key>` - Increment the integer value of a key by 1

### List Commands
- `RPUSH <key> <value> [value ...]` - Append one or more values to the end of a list
- `LPUSH <key> <value> [value ...]` - Prepend one or more values to the beginning of a list
- `LRANGE <key> <start> <stop>` - Get a range of elements from a list
- `LLEN <key>` - Get the length of a list
- `LPOP <key> [count]` - Remove and return the first element(s) of a list
- `BLPOP <key> <timeout>` - Blocking pop from the left side of a list

### Stream Commands
- `XADD <key> <id> <field> <value> [field value ...]` - Add an entry to a stream
- `XRANGE <key> <start> <end>` - Get a range of entries from a stream
- `XREAD [BLOCK <milliseconds>] [STREAMS] <key> [key ...] <id> [id ...]` - Read entries from one or more streams

### Transaction Commands
- `MULTI` - Start a transaction block
- `EXEC` - Execute all commands queued after MULTI
- `DISCARD` - Discard all commands queued after MULTI

### Utility Commands
- `TYPE <key>` - Determine the type of value stored at a key

## Developer Setup

### Prerequisites

- Go 1.24 or later

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd gokv
```

2. Install dependencies (if any):
```bash
go mod download
```

### Running the Server

1. Build the application:
```bash
go build -o redis-server ./app
```

2. Run the server (defaults to port 6379):
```bash
./redis-server
```

Or specify a custom port:
```bash
./redis-server -port 6380
```

3. Alternatively, use the provided script:
```bash
./your_program.sh
```

### Testing the Server

You can test the server using the `redis-cli` command-line tool:

```bash
# Connect to the server
redis-cli -p 6379

# Test basic commands
> PING
PONG
> SET key value
OK
> GET key
"value"
```

Or use `telnet` or `nc` (netcat) for raw RESP protocol testing:

```bash
echo -e "*1\r\n$4\r\nPING\r\n" | nc localhost 6379
```
