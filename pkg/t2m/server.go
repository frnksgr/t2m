package t2m

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// request logging middleware used for debugging
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				http.Error(w, fmt.Sprint(err),
					http.StatusInternalServerError)
				return
			}
			s := strings.ReplaceAll(
				strings.ReplaceAll(string(dump), "\r\n", "\n"),
				"\n", "\n  ")
			fmt.Fprintf(os.Stderr, "%s\n", s)
			next.ServeHTTP(w, r)
		})
}

// Server the HTTP server
type Server struct {
	// Unique ID of this server
	id     uuid.UUID
	server *http.Server
	// Target URL for subsequent requests
	targetURL string // balanced as binary tree. I.e. each request will at most create 2
	// sub requests. Each request is marked by a node.
}

// NewServer create a new server
func NewServer(addr string, targetURL string) *Server {
	r := mux.NewRouter()

	// Just use defaults
	s := &Server{
		id: uuid.New(),
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		targetURL: targetURL,
	}

	// --- ROUTES ---

	// Get help
	r.HandleFunc("/help", handleHelp).Methods("GET")
	// Health check for LM
	r.HandleFunc("/healthz", s.handleHealthz).Methods("GET")
	// Internal requests
	r.HandleFunc("/internal", s.handleInternalNode).Methods("POST")
	// External requests
	r.HandleFunc("/{task:sleep|fail|crash|cpu|ram}", s.handleRootNode).Methods("GET")
	r.HandleFunc("/", s.handleRootNode).Methods("GET")
	return s
}

// Health endpoint
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s OK\n", s.id)
}

// ListenAndServe start server
func (s *Server) ListenAndServe() error {
	if os.Getenv("DEBUG") != "" {
		s.server.Handler = requestLogger(s.server.Handler)
	}
	return s.server.ListenAndServe()
}
