package main

import (
	"encoding/json"
	//"flag"
	//"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wolkdb/contact-tracing-server/backend"
	"github.com/wolkdb/contact-tracing-server/server"
)

const (
	version        = "0.1"
	configFileName = "ct.conf"
	defaultCTDir   = "/tmp"
)

func main() {
	ctdir := os.Getenv("CTDIR")
	if ctdir == "" {
		ctdir = defaultCTDir
	}

	confFile := filepath.Join(ctdir, configFileName)
	conf, err := loadConfig(confFile)
	if err != nil {
		log.Printf("Err - loadConfig: %v\n", err)
	}
	log.Printf("conf %v", conf)

	port := os.Getenv("PORT")
	if port == "" {
		port = server.DefaultPort
	}

	backend, err := backend.NewBackend(conf)
	if err != nil {
	}
	//backend.Start()
	s, err := server.NewServer(port, backend)
	if err != nil {
		panic(err)
	}
	s.Start()
	log.Printf("Contact Tracing Server v%s - Listening on port %s...\n", version, port)
	for {
	}
}

func loadConfig(configFile string) (*backend.Config, error) {
	conf := new(backend.Config)
	jsonString, err := ioutil.ReadFile(configFile)
	if err != nil {
		return conf, err
	}

	err = json.Unmarshal(jsonString, conf)
	if err != nil {
		return conf, err
	}
	return conf, err
}
