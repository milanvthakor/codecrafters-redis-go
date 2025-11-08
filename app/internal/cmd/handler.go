package cmd

import (
	"github.com/codecrafters-io/redis-starter-go/app/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

// Handler represents a command handler function.
type Handler func(cmd []*protocol.RespVal, store *storage.Mem) (string, error)

// Registry holds all registered command handlers.
type Registry struct {
	handlers map[string]Handler
}

// NewRegistry creates a new command registry with all handlers registered.
func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]Handler),
	}

	// Register all commands
	r.Register("PING", handlePing)
	r.Register("ECHO", handleEcho)
	r.Register("SET", handleSet)
	r.Register("GET", handleGet)
	r.Register("RPUSH", handleRpush)
	r.Register("LRANGE", handleLrange)
	r.Register("LPUSH", handleLpush)
	r.Register("LLEN", handleLlen)
	r.Register("LPOP", handleLpop)
	r.Register("BLPOP", handleBlpop)
	r.Register("TYPE", handleType)
	r.Register("XADD", handleXadd)
	r.Register("XRANGE", handleXrange)
	r.Register("XREAD", handleXread)
	r.Register("INCR", handleIncr)

	return r
}

// Register registers a command handler.
func (r *Registry) Register(name string, handler Handler) {
	r.handlers[name] = handler
}

// Get retrieves a command handler by name.
func (r *Registry) Get(name string) (Handler, bool) {
	handler, ok := r.handlers[name]
	return handler, ok
}

// Execute executes a command using the registry.
func (r *Registry) Execute(cmdName string, cmd []*protocol.RespVal, store *storage.Mem) string {
	handler, ok := r.Get(cmdName)
	if !ok {
		return protocol.ToSimpErr("ERR unknown command")
	}

	respStr, err := handler(cmd, store)
	if err != nil {
		return protocol.ToSimpErr(err.Error())
	}

	return respStr
}
