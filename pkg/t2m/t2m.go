package t2m

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Topologies for node(s)
const (
	// binary tree
	topBinary = iota
	// linear tree
	topLinear = iota
	// fan tree (max level 1)
	topFan = iota
)

var topologyMap = map[string]int{
	"binary": topBinary,
	"linear": topLinear,
	"fan":    topFan,
}

// A GET request like http://domain/?count=10
// will create result in count-1 subsequent requests
// which will form a tree (binary, linear, fan)
// of node(s).
type node struct {
	// Topology of nodes
	Topology int
	// Unique ID of an request tree
	Request uuid.UUID
	// Unique index of request node
	Index int
	// Parents index of this node
	Parent int
	// Size of tree i.e. count of requests
	Count int
	// Depth of node in tree (root node at level 0)
	Level int
	// Logger used for this specific node/request
	logger *log.Logger
}

// Create child nodes
func (n *node) children() []*node {
	cn := []*node{}
	switch n.Topology {

	case topBinary:
		cn = make([]*node, 0, 2)
		for i := 0; i < 2; i++ {
			index := n.Index + 1<<(uint(n.Level+i))
			if index > n.Count {
				break
			}
			cn = append(cn,
				&node{
					Topology: n.Topology,
					Request:  n.Request,
					Count:    n.Count,
					Parent:   n.Index,
					Index:    index,
					Level:    n.Level + 1,
				})
		}

	case topLinear:
		if n.Index == n.Count {
			break
		}
		cn = []*node{
			&node{
				Topology: n.Topology,
				Request:  n.Request,
				Count:    n.Count,
				Parent:   n.Index,
				Index:    n.Index + 1,
				Level:    n.Level + 1,
			},
		}

	case topFan:
		if n.Index > 1 {
			break
		}
		cn = make([]*node, n.Count-1)
		for index := 2; index <= n.Count; index++ {
			cn[index-2] =
				&node{
					Topology: n.Topology,
					Request:  n.Request,
					Count:    n.Count,
					Parent:   n.Index,
					Index:    index,
					Level:    1,
				}
		}
	}

	return cn
}

func (n *node) spawn(c *node, url string) (*http.Response, error) {
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

type parameters struct {
	count    int
	topology int
}

func getParameters(q url.Values) parameters {
	// TODO: better error checking
	p := parameters{
		count:    1,
		topology: topBinary,
	}
	if vs := q["count"]; vs != nil {
		if i, err := strconv.Atoi(vs[0]); err == nil {
			if i < 1 {
				i = 1
			}
			if i > 4069 {
				i = 4069
			}
			p.count = i
		}
	}
	if vs := q["topology"]; vs != nil {
		p.topology = topologyMap[vs[0]]
	}
	return p
}

func (s *Server) handleRootNode(w http.ResponseWriter, r *http.Request) {
	// get query parameter and set defaults
	p := getParameters(r.URL.Query())
	uuid := uuid.New()
	prefix := fmt.Sprintf("[S: %s, R: %s, T: %d, L: %04d, P: %04d, N: %04d]\n  ",
		s.uuid, uuid, p.topology, 0, 0, 1)
	n := &node{
		Topology: p.topology,
		Request:  uuid,
		Index:    1,
		Parent:   0,
		Count:    p.count,
		Level:    0,
		logger:   log.New(os.Stdout, prefix, log.Lmicroseconds),
	}

	s.handleNode(n, w, r)
}

func (s *Server) handleInternalNode(w http.ResponseWriter, r *http.Request) {
	// decode node
	n := &node{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, n); err != nil {
		panic(err)
	}
	prefix := fmt.Sprintf("[S: %s, R: %s, T: %d, L: %04d, P: %04d, N: %04d]\n  ",
		s.uuid, n.Request, n.Topology, n.Level, n.Parent, n.Index)
	n.logger = log.New(os.Stdout, prefix, log.Lmicroseconds)

	s.handleNode(n, w, r)
}

func (s *Server) handleNode(n *node, w http.ResponseWriter, r *http.Request) {
	// here we start
	n.logger.Printf("request started")

	cn := n.children()

	// node result(s)
	nr := make(map[string]string)
	// node without children
	nr[s.uuid.String()] = fmt.Sprintf("%04d", n.Index)

	type childResult struct {
		resp *http.Response
		err  error
	}
	statusCode := http.StatusOK
	rc := make(chan childResult, len(cn))
	for _, c := range cn {
		go func(c *node) {
			resp, err := n.spawn(c, s.targetURL+"/internal")
			rc <- childResult{resp, err}
		}(c)
	}
	for range cn {
		cr := <-rc
		if cr.err != nil { // failure upon request to spawn a child
			panic(cr.err)
		}
		if cr.resp.StatusCode != http.StatusOK {
			// tread successful http response as intermediate problem
			statusCode = http.StatusServiceUnavailable // 503
		}
		body, err := ioutil.ReadAll(cr.resp.Body)
		if err != nil {
			panic(err)
		}
		cnr := make(map[string]string)
		if err := json.Unmarshal(body, &cnr); err != nil {
			panic(err)
		}

		// merge client result with node result
		for k, v := range cnr {
			if nr[k] == "" {
				nr[k] = v
			} else {
				// merge values
				nr[k] = fmt.Sprintf("%s %s", nr[k], v)
			}
		}
	}

	// create response

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	// write aggregated noderesult to response body encoded in json
	e := json.NewEncoder(w)
	if err := e.Encode(nr); err != nil {
		panic(err)
	}

	// here we end
	n.logger.Printf("request ended")
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
			fmt.Fprintf(os.Stderr, "%s\n", s)
			next.ServeHTTP(w, r)
		})
}

func (s *Server) handleFail(w http.ResponseWriter, r *http.Request) {
	log.Panicf("Terminating TCP connection to client %s", s.uuid)
}

func (s *Server) handleCrash(w http.ResponseWriter, r *http.Request) {
	log.Fatalf("Crashing server process %s", s.uuid)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s OK\n", s.uuid)
}
