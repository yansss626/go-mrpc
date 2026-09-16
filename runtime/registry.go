package mrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type UnaryFunc func(ctx context.Context, req json.RawMessage) (any, error)
type StreamFunc func(ctx context.Context, req json.RawMessage, stream *StreamServer) error

type methodKey struct {
	service string
	method  string
}
type Registry struct {
	mu       sync.RWMutex
	handlers map[methodKey]MethodHandler
}

type MethodHandler struct {
	UnaryHandler  UnaryFunc
	StreamHandler StreamFunc
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[methodKey]MethodHandler),
	}
}

func (r *Registry) Register(service string, method string, handler MethodHandler) error {

	if service == "" {
		return fmt.Errorf("service is empty")
	}

	if method == "" {
		return fmt.Errorf("method is empty")
	}

	if handler.StreamHandler == nil && handler.UnaryHandler == nil {
		return fmt.Errorf("handler is nil")
	}

	key := methodKey{
		service: service,
		method:  method,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[key]; exists {
		return fmt.Errorf(
			"handler already registered: %s.%s",
			service,
			method,
		)
	}

	r.handlers[key] = handler

	return nil
}
func (r *Registry) LookupUnary(service, method string) (UnaryFunc, bool) {

	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[methodKey{
		service: service,
		method:  method,
	}]
	return handler.UnaryHandler, ok
}
func (r *Registry) LookupStream(service, method string) (StreamFunc, bool) {

	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[methodKey{
		service: service,
		method:  method,
	}]

	return handler.StreamHandler, ok
}
