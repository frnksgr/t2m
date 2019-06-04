package t2m

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Server blabla
type Server struct {
	uuid      uuid.UUID
	server    *http.Server
	targetURL string
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

	// for now just use defaults
	s := &Server{
		uuid: uuid.New(),
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		targetURL: targetURL,
	}

	// routes
	r.HandleFunc("/", s.handleRootNode).Methods("GET").
		Queries("count", "{count:[1-9]\\d{0,3}}")
	r.HandleFunc("/inner", s.handleInnerNode).Methods("POST")

	return s
}

type node struct {
	Request uuid.UUID
	Index   int
	Count   int
	Level   int
	logger  *log.Logger
}

func (n *node) child(i int) *node {
	return &node{
		Request: n.Request,
		Count:   n.Count,
		Index:   n.Index + 1<<(uint(n.Level+i)),
		Level:   n.Level + 1,
	}
}

func (n *node) spawn(c *node, url string) (*http.Response, error) {
	n.logger.Printf("spawning child, index: %04d\n", c.Index)
	// create request body
	body, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	// create request object
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) handleRootNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	count, err := strconv.Atoi(vars["count"])
	if err != nil {
		panic(err)
	}

	uuid := uuid.New()
	prefix := fmt.Sprintf("[S: %s, R: %s, N: %04d] ", s.uuid, uuid, 1)
	n := &node{
		Request: uuid,
		Index:   1,
		Count:   count,
		Level:   0,
		logger:  log.New(os.Stdout, prefix, log.Lmicroseconds),
	}

	s.handleNode(n, w, r)
}

func (s *Server) handleInnerNode(w http.ResponseWriter, r *http.Request) {
	// decode node
	n := &node{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, n); err != nil {
		panic(err)
	}
	prefix := fmt.Sprintf("[S: %s, R: %s, N: %04d] ", s.uuid, n.Request, n.Index)
	n.logger = log.New(os.Stdout, prefix, log.Lmicroseconds)

	s.handleNode(n, w, r)
}

func (s *Server) handleNode(n *node, w http.ResponseWriter, r *http.Request) {
	cn := make([]*node, 0, 2) // child nodes
	nc := 1                   // node count

	// create a child if its index <= count
	for i := 0; i < 2; i++ {
		c := n.child(i)
		if c.Index > n.Count {
			break
		}
		cn = append(cn, c)
	}

	type childResult struct {
		resp *http.Response
		err  error
	}
	statusCode := http.StatusOK
	rc := make(chan childResult, len(cn))
	for _, c := range cn {
		go func(c *node) {
			resp, err := n.spawn(c, s.targetURL+"/inner")
			rc <- childResult{resp, err}
		}(c)
	}
	for range cn {
		cr := <-rc
		if cr.err != nil {
			panic(cr.err)
		}
		if cr.resp.StatusCode != http.StatusOK {
			statusCode = http.StatusServiceUnavailable // 503
		}
		body, err := ioutil.ReadAll(cr.resp.Body)
		if err != nil {
			statusCode = http.StatusServiceUnavailable
		}
		cnc, err := strconv.Atoi(strings.TrimSpace(string(body)))
		if err != nil {
			statusCode = http.StatusServiceUnavailable
		}
		nc += cnc
	}
	// create response
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, nc)
}

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
			fmt.Fprintf(os.Stderr, "ID: %s\n", s)
			next.ServeHTTP(w, r)
		})
}
