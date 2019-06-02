package t2m

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// Server blabla
type Server struct {
	server *http.Server
}

// ListenAndServe start server
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// NewServer create a new server
func NewServer(addr string, targetURL string) *Server {
	r := mux.NewRouter()

	// routes
	r.HandleFunc("/", handleRootNode).Methods("GET").
		Queries("count", "{count:[1-9]\\d{0,3}}")
	r.HandleFunc("/inner", handleInnerNode).Methods("POST")

	// for now just use defaults
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

type node struct {
	Index int
	Count int
	Level int
}

func (n *node) logger() *log.Logger {
	prefix := fmt.Sprintf("ID: %d ", n.Index)
	return log.New(os.Stdout, prefix, log.Lmicroseconds)
}

func (n *node) child(i int) *node {
	return &node{
		Count: n.Count,
		Index: n.Index + 1<<(uint(n.Level+i)),
		Level: n.Level + 1,
	}
}

func (n *node) spawn(url string) error {
	// create request body
	body, err := json.Marshal(n)
	if err != nil {
		return err
	}
	// create request object
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// handle response
	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed creating child")
	}
	return nil
}

func handleRootNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	count, err := strconv.Atoi(vars["count"])
	if err != nil {
		panic(err)
	}

	n := &node{
		Index: 1,
		Count: count,
		Level: 0,
	}

	handleNode(n, w, r)
}

func handleInnerNode(w http.ResponseWriter, r *http.Request) {
	// decode node
	n := &node{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, n); err != nil {
		panic(err)
	}

	handleNode(n, w, r)
}

func handleNode(n *node, w http.ResponseWriter, r *http.Request) {
	cv := make([]*node, 0, 2) // child nodes
	nc := 1                   // node count

	// create a child if its index <= count
	for i := 0; i < 2; i++ {
		c := n.child(i)
		if c.Index > n.Count {
			break
		}
		cv = append(cv, c)
	}

	for _, c := range cv {
		c.spawn()
	}
}
