package t2m

import (
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Server the HTTP server
type Server struct {
	// Unique ID of this server
	uuid   uuid.UUID
	server *http.Server
	// Target URL for subsequent requests
	targetURL string // balanced as binary tree. I.e. each request will at most create 2
	// sub requests. Each request is marked by a node.

}

// ListenAndServe start server
func (s *Server) ListenAndServe() error {
	if os.Getenv("DEBUG") != "" {
		s.server.Handler = requestLogger(s.server.Handler)
	}
	return s.server.ListenAndServe()
}

// NewServer create a new server
func NewServer(addr string, targetURL string) *Server {
	r := mux.NewRouter()

	// Just use defaults
	s := &Server{
		uuid: uuid.New(),
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		targetURL: targetURL,
	}

	// --- ROUTES ---

	// Ingress route for spanning request tree
	r.HandleFunc("/", s.handleRootNode).Methods("GET")
	// internal requests
	r.HandleFunc("/internal", s.handleInternalNode).Methods("POST")

	// Get help
	r.HandleFunc("/help", handleHelp).Methods("GET")

	// Some helpful handlers
	r.HandleFunc("/fail", s.handleFail)
	r.HandleFunc("/crash", s.handleCrash)
	r.HandleFunc("/healthz", s.handleHealthz)

	return s
}
