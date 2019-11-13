package t2m

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/google/uuid"
)

// Topologies for node(s)
const (
	// binary tree
	topFan = iota
	topChain
	topBtree
)

// nodeTypes
const (
	ntLeave = iota // a leave node i.e.no children
	ntInner        // an inner node i.e with children
)

var topologyMap = map[string]int{
	"fan":   topFan,
	"btree": topBtree,
	"chain": topChain,
}

// Node bla bla
// A GET request like http://task?count=10
// will create result in count-1 subsequent requests
// which will form a tree (btree, chain, fan)
// of node(s).
type node struct {
	// Topology of nodes
	Topology int
	// Unique ID of a request node in tree
	Request uuid.UUID
	// Unique index of request node (starting at 1)
	Index int
	// Parents index of this request node (0 == no parent i.e. root node)
	Parent int
	// Size of tree i.e. count of request nodes
	Count int
	// Depth of request node in tree (root node level == 0)
	Level int
	// Request URI defining task to be executed
	URI string
	// true if leave else false
	isLeave bool
	// Logger used for this specific request node
	logger *log.Logger
}

// Create child nodes structures for this particular node
func (n *node) children() []*node {
	cn := []*node{}
	switch n.Topology {

	case topBtree:
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
					URI:      n.URI,
				})
		}

	case topChain:
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
				URI:      n.URI,
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
					URI:      n.URI,
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
		count: 1,
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
		URI:      r.URL.RequestURI(),
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
	nr[s.uuid.String()] = fmt.Sprintf("%04d", n.Index)

	type childResult struct {
		resp *http.Response
		err  error
	}

	statusCode := http.StatusOK
	if len(cn) == 0 {
		n.isLeave = true
	}

	rc := make(chan childResult, len(cn))
	// spawn child nodes
	for _, c := range cn {
		go func(c *node) {
			resp, err := n.spawn(c, s.targetURL+"/internal")
			rc <- childResult{resp, err}
		}(c)
	}
	// fetch results from child nodes
	for range cn {
		cr := <-rc
		if cr.err != nil { // failure upon request to spawn a child
			panic(cr.err)
		}
		if cr.resp.StatusCode != http.StatusOK {
			// tread non 200 http response as intermediate problem
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

	// Execute task
	if err := n.execTask(); err != nil {
		panic(err)
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	// write aggregated noderesult to response body encoded in json
	e := json.NewEncoder(w)
	if err := e.Encode(nr); err != nil {
		panic(err)
	}

	// we are done with this node
	n.logger.Printf("request ended")
}
