package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/wolkdb/cen-server/backend"
	"github.com/wolkdb/cen-server/server"
)

// DefaultTransport contains all HTTP client operation parameters
var DefaultTransport http.RoundTripper = &http.Transport{
	Dial: (&net.Dialer{
		// limits the time spent establishing a TCP connection (if a new one is needed)
		Timeout:   120 * time.Second,
		KeepAlive: 120 * time.Second, // 60 * time.Second,
	}).Dial,
	//MaxIdleConns: 5,
	MaxIdleConnsPerHost: 25, // changed from 100 -> 25

	// limits the time spent reading the headers of the response.
	ResponseHeaderTimeout: 120 * time.Second,
	IdleConnTimeout:       120 * time.Second, // 90 * time.Second,

	// limits the time the client will wait between sending the request headers when including an Expect: 100-continue and receiving the go-ahead to send the body.
	ExpectContinueTimeout: 1 * time.Second,

	// limits the time spent performing the TLS handshake.
	TLSHandshakeTimeout: 5 * time.Second,
}

func httppost(url string, body []byte) (result []byte, err error) {

	httpclient := &http.Client{Timeout: time.Second * 120, Transport: DefaultTransport}
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httppost] %s", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httppost] %s", err)
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httppost] %s", err)
	}
	resp.Body.Close()

	return result, nil
}

func httpget(url string) (result []byte, err error) {

	httpclient := &http.Client{Timeout: time.Second * 120, Transport: DefaultTransport}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httpget] %s", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httpget] %s", err)
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("[cen_test:httpget] %s", err)
	}
	resp.Body.Close()

	return result, nil
}

func TestCENSimple(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
	}
	endpoint := fmt.Sprintf("%s.wolk.com:%s", hostname, server.DefaultPort)

	var reports []backend.CENReport
	var hashKeys [][]byte
	for i := 0; i < 10; i++ {
		key := make([]byte, 16)
		rand.Read(key)
		hashKey := backend.Computehash(key)
		hashKeys = append(hashKeys, hashKey)
		symptom := "sample symptom"

		report := backend.CENReport{HashedPK: hashKey, EncodedMsg: []byte(symptom)}
		reports = append(reports, report)
	}
	cenReportJSON, err := json.Marshal(reports)
	_, err = httppost(fmt.Sprintf("https://%s/%s", endpoint, server.EndpointCENReport), cenReportJSON)
	if err != nil {
		t.Fatalf("EndpointCENReport: %s", err)
	}

	var prefixHashedKey []byte
	sampleKey := hashKeys[3]
	sampleKey2 := hashKeys[6]
	prefixHashedKey = append(prefixHashedKey, sampleKey[0])
	prefixHashedKey = append(prefixHashedKey, sampleKey[1])
	prefixHashedKey = append(prefixHashedKey, (sampleKey[2]&0xC0)|(sampleKey2[0]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[0]&03<<6)|(sampleKey2[1]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[1]&03<<6)|(sampleKey2[2]&0xFC>>2))

	result, err := httppost(fmt.Sprintf("https://%s/%s/%d", endpoint, server.EndpointCENQuery, time.Now().Unix()), prefixHashedKey)
	if err != nil {
		t.Fatalf("EndpointCENReport: %s", err)
	}

	var resultreport []*backend.CENReport
	err = json.Unmarshal(result, &resultreport)
	if err != nil {
		t.Fatalf("EndpointCENReport(check1): %s", err)
	}
	for _, r := range resultreport {
		if bytes.Compare(r.HashedPK, sampleKey) == 0 || bytes.Compare(r.HashedPK, sampleKey2) == 0 {
			fmt.Println("ok")
		} else {
			fmt.Println("err")
		}
	}
}
