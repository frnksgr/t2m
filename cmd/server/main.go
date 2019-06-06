package main

import (
	"fmt"
	"log"
	"os"

	"github.com/frnksgr/t2m/pkg/config"
	"github.com/frnksgr/t2m/pkg/t2m"
)

var cfg = struct {
	ListeningPort    string
	ListeningAddress string
	TargetURL        string
}{
	ListeningPort:    "8080",
	ListeningAddress: "0.0.0.0",
	TargetURL:        "http://localhost:8080",
}

func init() {
	if err := config.FromEnv(&cfg); err != nil {
		panic(err)
	}
}

func main() {
	addr := fmt.Sprintf("%s:%s", cfg.ListeningAddress, cfg.ListeningPort)
	srv := t2m.NewServer(addr, cfg.TargetURL)

	log.Println("Version", t2m.Version)
	// print cofiguration if in debug mode
	if os.Getenv("DEBUG") != "" {
		fmt.Printf("Configuration: %+v\n", cfg)
	}
	// start server
	log.Println("Starting http server on ", addr, "...")
	log.Fatal(srv.ListenAndServe())
}
