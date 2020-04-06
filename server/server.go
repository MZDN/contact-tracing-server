package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/wolkdb/cen-server/backend"
)

const (
	// adjust these below to your SSL Cert location
	sslBaseDir     = "/etc/pki/tls/certs/wildcard/wolk.com-new"
	sslKeyFileName = "www.wolk.com.key"
	caFileName     = "www.wolk.com.bundle"

	// DefaultPort is the port which the CEN HTTP server is listening in on
	DefaultPort = "8080"

	// EndpointCENReport is the name of the HTTP endpoint for GET/POST of CENReport
	EndpointCENReport = "cenreport"

	// EndpointCENKeys is the name of the HTTP endpoint for GET CenKeys
	EndpointCENKeys = "cenkeys"
)

// Server manages HTTP connections
type Server struct {
	backend  *backend.Backend
	Handler  http.Handler
	HTTPPort string
}

// NewServer returns an HTTP Server to handle simple-api-process-flow https://github.com/Co-Epi/data-models/blob/master/simple-api-process-flow.md
func NewServer(httpPort string, backend *backend.Backend) (s *Server, err error) {
	s = &Server{
		HTTPPort: httpPort,
	}
	/*
		backend, err := backend.NewBackend(connString)
		if err != nil {
			log.Printf("backend error %v", err)
			return s, err
		}
	*/
	s.backend = backend

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.getConnection)
	s.Handler = mux
	return s, nil
}

func (s *Server) getConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("getConnection")
	if strings.Contains(r.URL.Path, EndpointCENReport) {
		if r.Method == http.MethodPost {
			s.postCENReportHandler(w, r)
		} else {
			s.getCENReportHandler(w, r)
		}
	} else if strings.Contains(r.URL.Path, EndpointCENKeys) {
		s.getCENKeysHandler(w, r)
	} else {
		s.homeHandler(w, r)
	}
}

// Start kicks off the HTTP Server
func (s *Server) Start() (err error) {
	srv := &http.Server{
		Addr:         ":" + s.HTTPPort,
		Handler:      s.Handler,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ssldir := os.Getenv("SSLDIR")
	if ssldir == "" {
		ssldir = sslBaseDir
	}
	SSLKeyFile := path.Join(ssldir, sslKeyFileName)
	CAFile := path.Join(ssldir, caFileName)
	log.Printf("SSLKeyFile = %s CAFile = %s", SSLKeyFile, CAFile)

	// Note: bringing the intermediate certs with CAFile into a cert pool and the tls.Config is *necessary*
	certpool := x509.NewCertPool() // https://stackoverflow.com/questions/26719970/issues-with-tls-connection-in-golang -- instead of x509.NewCertPool()
	log.Printf("certpool %v\n", certpool)
	pem, err := ioutil.ReadFile(CAFile)
	if err != nil {
		log.Printf("Failed to read client certificate authority: %v", err)
		return fmt.Errorf("Failed to read client certificate authority: %v", err)
	}
	if !certpool.AppendCertsFromPEM(pem) {
		log.Printf("Can't parse client certificate authority")
		return fmt.Errorf("Can't parse client certificate authority")
	}

	config := tls.Config{
		ClientCAs:  certpool,
		ClientAuth: tls.NoClientCert, // tls.RequireAndVerifyClientCert,
	}
	config.BuildNameToCertificate()
	log.Printf("tls config ok %v\n", config)

	srv.TLSConfig = &config

	err = srv.ListenAndServeTLS(CAFile, SSLKeyFile)
	log.Printf("Server listening on port %s %v", s.HTTPPort, err)
	if err != nil {
		log.Printf("ListenAndServeTLS err %v", err)
		return err
	}
	log.Printf("Server listening on port %s", s.HTTPPort)
	return nil
}

// POST /cenreport
func (s *Server) postCENReportHandler(w http.ResponseWriter, r *http.Request) {
	// Read Post Body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Parse body as CENReport
	var payload backend.CENReport
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Process CENReport payload
	err = s.backend.ProcessCENReport(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write([]byte("OK"))
}

// GET /cenreport/<cenkey>
func (s *Server) getCENReportHandler(w http.ResponseWriter, r *http.Request) {
	cenKey := ""
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) >= 1 {
		cenKey = pathpieces[1]
	} else {
		http.Error(w, "Usage: Usage: /cenreport/<cenkey>", http.StatusBadRequest)
		return
	}

	// Handle CenKey
	reports, err := s.backend.ProcessGetCENReport(cenKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	responsesJSON, err := json.Marshal(reports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("getCENReportHandler: %s\n", responsesJSON)
	w.Write(responsesJSON)
}

// GET /cenkeys/<timestamp>
func (s *Server) getCENKeysHandler(w http.ResponseWriter, r *http.Request) {
	ts := uint64(0)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathpieces) > 1 {
		tsa, err := strconv.Atoi(pathpieces[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ts = uint64(tsa)
	} else {
		ts = uint64(time.Now().Unix()) - 3600
	}

	cenKeys, err := s.backend.ProcessGetCENKeys(ts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responsesJSON, err := json.Marshal(cenKeys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("genCENKeysHandler: %s\n", responsesJSON)
	w.Write(responsesJSON)
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CEN API Server v0.2.1"))
}

//POST /censymptom
func (s *Server) postSymptomHander(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Parse body as CENReport
	var payload []backend.CENSymptomReport
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := backend.ProcessSymptomReport(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write([]byte("OK"))
}

//POST /cenrequestreport
func (s *Server) postRequestSymptomReportHander(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Parse body as CENReport
	var payload backend.CENSymptomRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestid, err := s.backend.RequestSymptomReport(payload.PrefixBitVector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	//w.Write([]byte("OK"))
	w.Write(requestid)
}

//GET /censymptomresult/reportid
func (s *Server) getSymptomResultHandler(w http.ResponseWriter, r *http.Request) {
	ts := uint64(0)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	results, err := s.backend.ProcessSymptomResult(pathpieces[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responsesJSON, err := json.Marshal(resultis)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("genCENKeysHandler: %s\n", responsesJSON)
	w.Write(responsesJSON)
}

//GET /censymptomstatus/reportid
func (s *Server) getSymptomReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	ts := uint64(0)
	pathpieces := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	results, err := s.backend.GetSymptomStatus(pathpieces[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responsesJSON, err := json.Marshal(resultis)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("genCENKeysHandler: %s\n", responsesJSON)
	w.Write(responsesJSON)
}
