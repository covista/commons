package main

import (
	"log"

	"github.com/covista/commons/internal/config"
	"github.com/covista/commons/internal/logging"
	"github.com/covista/commons/internal/server"
)

func main() {
	srv, err := server.NewFromConfig(logging.NewContextWithLogger(), config.NewFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Shutdown()
	if err := srv.ServeGRPC(); err != nil {
		log.Fatal(err)
	}
}
