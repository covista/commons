package main

import (
	"context"
	"log"

	"github.com/covista/commons/internal/server"
)

func main() {
	srv, err := server.NewWithInsecureDefaults(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Shutdown()
	if err := srv.ServeGRPC(); err != nil {
		log.Fatal(err)
	}
}
