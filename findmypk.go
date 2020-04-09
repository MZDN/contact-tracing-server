package main

import (
	"encoding/json"
	//"flag"
	//"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wolkdb/findmypk-server/backend"
	"github.com/wolkdb/findmypk-server/server"
)

const (
	version        = "0.1"
	configFileName = "fmpk.conf"
	defaultTmpDir  = "/tmp"
)

/*
type Config struct {
	MysqlConn string `json:"mysqlConn,omitempty"`
}
*/

func main() {
	tmpdir := os.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = defaultTmpDir
	}

	confFile := filepath.Join(tmpdir, configFileName)
	conf, err := loadConfig(confFile)
	if err != nil {
		log.Printf("Err - loadConfig: %v\n", err)
	}
	log.Printf("conf %v", conf)

	port := os.Getenv("PORT")
	if port == "" {
		port = server.DefaultPort
	}

	findMyPkBackend, err := backend.NewBackend(conf)
	if err != nil {
	}
	/*
		mysqlconn := os.Getenv("MYSQLCONN")
		if mysqlconn == "" {
			mysqlconn = conf.MysqlConn
		}
		mysqlConnectionString := flag.String("conn", mysqlconn, "MySQL Connection String")
	*/

	//s, err := server.NewServer(port, *mysqlConnectionString)
	s, err := server.NewServer(port, findMyPkBackend)
	if err != nil {
		panic(err)
	}
	s.Start()
	log.Printf("FindMyPk Server v%s - Listening on port %s...\n", version, port)
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
