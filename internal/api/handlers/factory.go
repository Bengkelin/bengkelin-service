package handlers

import (
	"sync"
)

// HandlerFactory manages handler instances
type HandlerFactory struct {
	authHandler    AuthHandlerInterface
	healthHandler  HealthHandlerInterface
	metricsHandler MetricsHandlerInterface
	mu             sync.RWMutex
}

var (
	factory *HandlerFactory
	once    sync.Once
)

// GetHandlerFactory returns the singleton handler factory
func GetHandlerFactory() *HandlerFactory {
	once.Do(func() {
		factory = &HandlerFactory{}
	})
	return factory
}

// GetAuthHandler returns the auth handler instance
func (f *HandlerFactory) GetAuthHandler() AuthHandlerInterface {
	f.mu.RLock()
	if f.authHandler != nil {
		defer f.mu.RUnlock()
		return f.authHandler
	}
	f.mu.RUnlock()
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Double-check locking pattern
	if f.authHandler == nil {
		f.authHandler = NewAuthHandler()
	}
	
	return f.authHandler
}

// GetHealthHandler returns the health handler instance
func (f *HandlerFactory) GetHealthHandler() HealthHandlerInterface {
	f.mu.RLock()
	if f.healthHandler != nil {
		defer f.mu.RUnlock()
		return f.healthHandler
	}
	f.mu.RUnlock()
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Double-check locking pattern
	if f.healthHandler == nil {
		f.healthHandler = NewHealthHandler()
	}
	
	return f.healthHandler
}

// GetMetricsHandler returns the metrics handler instance
func (f *HandlerFactory) GetMetricsHandler() MetricsHandlerInterface {
	f.mu.RLock()
	if f.metricsHandler != nil {
		defer f.mu.RUnlock()
		return f.metricsHandler
	}
	f.mu.RUnlock()
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Double-check locking pattern
	if f.metricsHandler == nil {
		f.metricsHandler = NewMetricsHandler()
	}
	
	return f.metricsHandler
}

// Legacy functions for backward compatibility
func GetAuthHandler() AuthHandlerInterface {
	return GetHandlerFactory().GetAuthHandler()
}

func GetHealthHandler() HealthHandlerInterface {
	return GetHandlerFactory().GetHealthHandler()
}

func GetMetricsHandler() MetricsHandlerInterface {
	return GetHandlerFactory().GetMetricsHandler()
}

// Reset resets all handlers (useful for testing)
func Reset() {
	factory = nil
	once = sync.Once{}
}