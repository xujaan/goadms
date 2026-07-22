package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

// SSEBroker manages Server-Sent Events connections and broadcasts.
type SSEBroker struct {
	clients   map[chan string]struct{}
	mu        sync.RWMutex
	logger    *slog.Logger
}

func NewSSEBroker(logger *slog.Logger) *SSEBroker {
	return &SSEBroker{
		clients: make(map[chan string]struct{}),
		logger:  logger,
	}
}

// Subscribe adds a client channel and returns it.
func (b *SSEBroker) Subscribe() chan string {
	ch := make(chan string, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	b.logger.Debug("sse: client connected", "total", len(b.clients))
	return ch
}

// Unsubscribe removes a client channel.
func (b *SSEBroker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
	b.logger.Debug("sse: client disconnected", "total", len(b.clients))
}

// Broadcast sends an event to all connected clients.
func (b *SSEBroker) Broadcast(eventType string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(payload))

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Client too slow, skip.
		}
	}
}
