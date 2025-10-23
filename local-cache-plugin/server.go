package cache

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/coredns/coredns/plugin/pkg/reuseport"
)

// HTTPServer provides HTTP endpoints for cache inspection
type HTTPServer struct {
	cache *Cache
	addr  string
	ln    net.Listener
	srv   *http.Server
	mux   *http.ServeMux
}

// NewHTTPServer creates a new HTTP server for cache inspection
func NewHTTPServer(cache *Cache, addr string) *HTTPServer {
	return &HTTPServer{
		cache: cache,
		addr:  addr,
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start() error {
	ln, err := reuseport.Listen("tcp", h.addr)
	if err != nil {
		return err
	}

	h.ln = ln
	h.mux = http.NewServeMux()
	h.mux.HandleFunc("/cache/stats", h.cache.HandleStats())
	h.mux.HandleFunc("/cache/entries", h.cache.HandleEntries())

	h.srv = &http.Server{
		Handler:      h.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		h.srv.Serve(ln)
	}()

	return nil
}

// Stop stops the HTTP server
func (h *HTTPServer) Stop() error {
	if h.srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return h.srv.Shutdown(ctx)
}