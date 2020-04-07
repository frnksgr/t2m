package t2m

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

const (
	maxSize = 1000
)

var taskRe = regexp.MustCompile("/([^/?]*)")
var errUnknownTask = errors.New("Unknown task specified")
var errQueryParameter = errors.New("Query parameter wrong")

// node manages meta data of a request node
type node struct {
	// Unique ID of request
	RequestID uuid.UUID
	// Topology of nodes
	Topology string
	// Unique index of request node (starting at 1)
	Index int
	// Parents index of this request node (0 == no parent i.e. root node)
	ParentIndex int
	// Size of tree i.e. number of request nodes
	Size int
	// Depth of request node in tree (root node level == 0)
	Depth int
	// Name of Task
	TaskName string
	// Duration of task execution in ms
	TaskDuration int
	// Logger used for this specific request node
	logger *log.Logger
}

// construct a new node
// set defaults and update values from URL
// return errUnknownTask, ...
func newNodeFromURL(url *url.URL) (*node, error) {
	n := &node{
		RequestID:    uuid.New(),
		Topology:     "fan",
		Index:        1,
		ParentIndex:  0,
		Size:         1,
		Depth:        0,
		TaskName:     "",
		TaskDuration: 50,
	}

	// parse URL and update node values accordingly

	// get task name
	switch t := taskRe.FindStringSubmatch(url.RequestURI())[1]; t {
	case "", "sleep", "fail", "crash", "cpu", "ram":
		n.TaskName = t
	default:
		return nil, errUnknownTask
	}

	// get query parameter
	q := url.Query()

	// n.Size
	if s, ok := q["size"]; ok {
		i, err := strconv.Atoi(s[0])
		if err != nil || i < 1 || i > maxSize {
			return nil, errQueryParameter
		}
		n.Size = i
	}

	// n.Topology
	if t, ok := q["topology"]; ok {
		switch t[0] {
		case "fan", "chain", "tree":
			n.Topology = t[0]
		default:
			return nil, errQueryParameter
		}
	}

	// n.TaskDuration
	if t, ok := q["time"]; ok {
		t, err := strconv.Atoi(t[0])
		if err != nil || t < 1 {
			return nil, errQueryParameter
		}
		n.TaskDuration = t
	}

	return n, nil
}

// construct a new node from parent Node
// set defaults and update values from URL
// return errUnknownTask, ...
func (n *node) newChild() *node {
	c := &node{
		RequestID:    uuid.New(),
		Topology:     n.Topology,
		Index:        -1, // unspecified
		ParentIndex:  n.Index,
		Size:         n.Size,
		Depth:        -1, // unspecified
		TaskName:     n.TaskName,
		TaskDuration: n.TaskDuration,
	}

	return c
}

// Create child node structures
// to be passed to subsequent requests
func (n *node) children() []*node {
	cn := []*node{}
	switch n.Topology {

	case "tree":
		cn = make([]*node, 0, 2)
		for i := 0; i < 2; i++ {
			index := n.Index + 1<<(uint(n.Depth+i))
			if index > n.Size {
				break
			}
			nn := n.newChild()
			nn.Depth = n.Depth + 1
			nn.Index = index
			cn = append(cn, nn)
		}

	case "chain":
		if n.Index == n.Size {
			break
		}
		nn := n.newChild()
		nn.Index = n.Index + 1
		nn.Depth = n.Depth + 1
		cn = []*node{nn}

	case "fan":
		if n.Index > 1 {
			break
		}
		cn = make([]*node, n.Size-1)
		for index := 2; index <= n.Size; index++ {
			nn := n.newChild()
			nn.Depth = 1
			nn.Index = index
			cn[index-2] = nn
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

func (s *Server) handleRootNode(w http.ResponseWriter, r *http.Request) {
	n, err := newNodeFromURL(r.URL)
	if err != nil {
		// TODO better error handling
		log.Fatalln("Oops", err)
	}
	prefix := fmt.Sprintf("[S: %s, R: %s, D: %04d, P: %04d, N: %04d]\n  ",
		s.id, n.RequestID, 0, 0, 1)
	n.logger = log.New(os.Stdout, prefix, log.Lmicroseconds)
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
	prefix := fmt.Sprintf("[S: %s, R: %s, D: %04d, P: %04d, N: %04d]\n  ",
		s.id, n.RequestID, n.Depth, n.ParentIndex, n.Index)
	n.logger = log.New(os.Stdout, prefix, log.Lmicroseconds)
	s.handleNode(n, w, r)
}

func (s *Server) handleNode(n *node, w http.ResponseWriter, r *http.Request) {
	// here we start
	n.logger.Printf("request started")

	cn := n.children()

	// node result(s)
	nr := make(map[string]string)
	nr[s.id.String()] = fmt.Sprintf("%04d", n.Index)

	type childResult struct {
		resp *http.Response
		err  error
	}

	statusCode := http.StatusOK

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

	// Execute task on any node
	n.execTask()

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
