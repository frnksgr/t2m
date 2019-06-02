package main

import (
	"fmt"
	"log"

	"github.com/frnksgr/t2m/pkg/t2m"
)

var config = struct {
	ListeningPort    string
	ListeningAddress string
	TargetURL        string
}{
	ListeningPort:    "8080",
	ListeningAddress: "0.0.0.0",
	TargetURL:        "http://localhost:8080",
}

func main() {
	addr := fmt.Sprintf("%s:%s", config.ListeningAddress,
		config.ListeningPort)
	srv := t2m.NewServer(addr, config.TargetURL)
	log.Fatal(srv.ListenAndServe())
}
