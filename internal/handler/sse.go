package handler

import (
	"fmt"
	"net/http"

	"github.com/jan/goadms/internal/service"
)

type SSEHandler struct {
	broker *service.SSEBroker
}

func NewSSEHandler(broker *service.SSEBroker) *SSEHandler {
	return &SSEHandler{broker: broker}
}

// Stream handles GET /api/v1/events — Server-Sent Events endpoint.
func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.broker.Subscribe()
	defer h.broker.Unsubscribe(ch)

	// Send initial connection event.
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		}
	}
}
