package queue

import (
	"context"
	"sync"
)

// jobRegistry implements JobRegistry interface
type jobRegistry struct {
	mu       sync.RWMutex
	handlers map[string]JobHandler
}

// NewJobRegistry creates a new job registry
func NewJobRegistry() JobRegistry {
	return &jobRegistry{
		handlers: make(map[string]JobHandler),
	}
}

// Register registers a job handler
func (r *jobRegistry) Register(jobType string, handler JobHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[jobType] = handler
}

// Get gets a job handler by type
func (r *jobRegistry) Get(jobType string) (JobHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, exists := r.handlers[jobType]
	return handler, exists
}

// asynqServer wraps Asynq server for job processing
type asynqServer struct {
	registry JobRegistry
}

// NewAsynqServer creates a new Asynq server wrapper
func NewAsynqServer(registry JobRegistry) *asynqServer {
	return &asynqServer{
		registry: registry,
	}
}

// ProcessTask processes a task by delegating to registered handler
func (s *asynqServer) ProcessTask(ctx context.Context, jobType string, payload []byte) error {
	handler, exists := s.registry.Get(jobType)
	if !exists {
		return nil // Skip unknown jobs
	}

	return handler(ctx, payload)
}
