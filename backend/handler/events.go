package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"socialai/service"
	"time"
)

// SSE compares bounded search snapshots across instances through shared Elasticsearch.
// This deliberately avoids an in-memory event bus that misses writes on other GAE instances.
func eventsHandler(w http.ResponseWriter, r *http.Request) {
	option, err := parseSearch(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unavailable", 503)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.NewTimer(4 * time.Minute)
	defer timeout.Stop()
	var previous []byte
	send := func() bool {
		if !validSession(currentUser(r)) {
			fmt.Fprint(w, "event: session-expired\ndata: {}\n\n")
			flusher.Flush()
			return false
		}
		posts, err := service.SearchPosts(r.Context(), option.user, option.keywords, option.mediaType, option.offset, option.limit)
		if err != nil {
			fmt.Fprint(w, "event: unavailable\ndata: {}\n\n")
			flusher.Flush()
			return false
		}
		payload, err := json.Marshal(posts)
		if err != nil {
			return false
		}
		if !bytes.Equal(previous, payload) {
			if _, err := fmt.Fprintf(w, "event: posts\ndata: %s\n\n", payload); err != nil {
				return false
			}
			previous = payload
		} else {
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return false
			}
		}
		flusher.Flush()
		return true
	}
	if !send() {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout.C:
			return
		case <-ticker.C:
			if !send() {
				return
			}
		}
	}
}
