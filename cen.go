package main

import (
	"flag"
	//"fmt"
	"log"
	"os"

	"github.com/wolkdb/cen-server/backend"
	"github.com/wolkdb/cen-server/server"
)

const (
	version = "0.2"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = server.DefaultPort
	}
	mysqlConnectionString := flag.String("conn", backend.DefaultConnString, "MySQL Connection String")

	s, err := server.NewServer(port, *mysqlConnectionString)
	if err != nil {
		panic(err)
	}
	s.Start()
	log.Printf("CEN Server v%s - Listening on port %s...\n", version, port)
	for {
	}
}
