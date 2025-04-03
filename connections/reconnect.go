package connections

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	// DefaultInitialBackoff is the default initial backoff duration for reconnection attempts
	DefaultInitialBackoff = 5 * time.Second

	// DefaultMaxBackoff is the default maximum backoff duration for reconnection attempts
	DefaultMaxBackoff = 60 * time.Second

	// DefaultBackoffFactor is the default factor by which the backoff duration increases between attempts
	DefaultBackoffFactor = 1.5
)

// StreamerFactory is a function that creates a new Streamer
type StreamerFactory[T any] func() (Streamer[T], error)

// ReconnectingStreamer wraps a Streamer with reconnection logic
type ReconnectingStreamer[T any] struct {
	factory         StreamerFactory[T]
	currentStreamer Streamer[T]
	ctx             context.Context
	cancel          context.CancelFunc

	initialBackoff time.Duration
	maxBackoff     time.Duration
	backoffFactor  float64

	logger      Logger
	mu          sync.Mutex
	isConnected bool

	// For testing and monitoring
	reconnectCount int
}

// ReconnectingOptions configures the ReconnectingStreamer
type ReconnectingOptions struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	Logger         Logger
}

// DefaultReconnectingOptions returns the default options for ReconnectingStreamer
func DefaultReconnectingOptions() ReconnectingOptions {
	return ReconnectingOptions{
		InitialBackoff: DefaultInitialBackoff,
		MaxBackoff:     DefaultMaxBackoff,
		BackoffFactor:  DefaultBackoffFactor,
		Logger:         NewDefaultLogger(),
	}
}

// NewReconnectingStreamer creates a new ReconnectingStreamer with the given factory
// The factory function is called to create a new Streamer when needed (on first connect or reconnect)
// The provided context controls the lifetime of the ReconnectingStreamer.
func NewReconnectingStreamer[T any](ctx context.Context, factory StreamerFactory[T], options ReconnectingOptions) (*ReconnectingStreamer[T], error) {
	// Create a cancellable context derived from the parent context
	derivedCtx, cancel := context.WithCancel(ctx)

	r := &ReconnectingStreamer[T]{
		factory:        factory,
		ctx:            derivedCtx, // Store the derived context
		cancel:         cancel,     // Store the cancel function for Close()
		initialBackoff: options.InitialBackoff,
		maxBackoff:     options.MaxBackoff,
		backoffFactor:  options.BackoffFactor,
		logger:         options.Logger,
	}

	// Initial connection using the derived context for the factory call
	// Although the factory might use its own context internally (like in the gRPC example),
	// we check our own context first before attempting connection.
	select {
	case <-r.ctx.Done():
		cancel() // Clean up the derived context
		return nil, fmt.Errorf("initial connection cancelled by parent context: %w", r.ctx.Err())
	default:
		// Continue with initial connection attempt
	}

	streamer, err := factory()
	if err != nil {
		r.logger.Error("failed to create initial streamer: %v", err)
		cancel()      // Clean up the derived context if initial connection fails
		return r, err // Return r so Close() can still be called if needed, along with the error
	}

	r.currentStreamer = streamer
	r.isConnected = true
	r.logger.Info("successfully established initial connection")

	return r, nil
}

// Streamer function that implements reconnection logic
func (r *ReconnectingStreamer[T]) Streamer() Streamer[T] {
	var generator Streamer[T] = func() (T, error) {
		r.mu.Lock()
		streamer := r.currentStreamer
		r.mu.Unlock()

		var zeroValue T
		for {
			// Try to get data from the current streamer
			data, err := streamer()
			if err == nil {
				return data, nil
			}

			// If context is done, return error
			select {
			case <-r.ctx.Done():
				// Use r.ctx.Err() to provide the reason for cancellation
				return zeroValue, fmt.Errorf("context canceled: %w", r.ctx.Err())
			default:
				// Context not canceled, continue with reconnection
			}

			// Handle reconnection
			r.mu.Lock()
			r.isConnected = false
			r.logger.Error("connection lost: %v, attempting to reconnect", err)

			// Reconnection attempt with exponential backoff
			backoff := r.initialBackoff
			for attempts := 0; r.ctx.Err() == nil; attempts++ {
				r.mu.Unlock()
				// Add jitter (e.g., up to 1 second)
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				time.Sleep(backoff + jitter)
				r.mu.Lock()

				if r.ctx.Err() != nil {
					break // Exit loop immediately if context is cancelled during sleep
				}

				r.logger.Info("attempting reconnection (attempt #%d)...", attempts+1)

				newStreamer, err := r.factory()
				if err == nil {
					r.logger.Info("successfully reconnected")
					r.currentStreamer = newStreamer
					streamer = newStreamer
					r.isConnected = true
					r.reconnectCount++
					r.mu.Unlock()
					break
				}

				r.logger.Error("reconnection failed: %v", err)

				// Increase backoff for next attempt, bounded by maxBackoff
				backoff = time.Duration(float64(backoff) * r.backoffFactor)
				if backoff > r.maxBackoff {
					backoff = r.maxBackoff
				}
				r.logger.Info("next reconnection attempt in %s", backoff)
			}

			// If still locked, unlock
			if r.isConnected {
				continue
			}

			r.mu.Unlock()
			// Ensure the error reflects the context cancellation if that was the cause
			finalErr := errors.New("reconnection failed")
			if r.ctx.Err() != nil {
				finalErr = fmt.Errorf("reconnection failed: %w", r.ctx.Err())
			}
			return zeroValue, finalErr
		}
	}

	return generator
}

// Close stops the reconnection process
func (r *ReconnectingStreamer[T]) Close() {
	r.cancel()
}

// Into passes messages to the provided channel with reconnection support
func (r *ReconnectingStreamer[T]) Into(ch chan T) {
	r.Streamer().Into(ch)
}

// Channel creates a channel that receives messages with reconnection support
func (r *ReconnectingStreamer[T]) Channel(size int) chan T {
	return r.Streamer().Channel(size)
}

// IsConnected returns whether the streamer is currently connected
func (r *ReconnectingStreamer[T]) IsConnected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isConnected
}

// GetReconnectCount returns the number of successful reconnections
func (r *ReconnectingStreamer[T]) GetReconnectCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reconnectCount
}
