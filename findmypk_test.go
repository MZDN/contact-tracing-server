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

	"github.com/wolkdb/findmypk-server/backend"
	"github.com/wolkdb/findmypk-server/server"
)

const endpoint = "covidwatch.wolk.com"

//const endpoint = "d1.wolk.com:8080"

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
		return result, fmt.Errorf("[findmypk_test:httppost] %s", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return result, fmt.Errorf("[findmypk_test:httppost] %s", err)
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("[findmypk_test:httppost] %s", err)
	}
	resp.Body.Close()

	return result, nil
}

func httpget(url string) (result []byte, err error) {

	httpclient := &http.Client{Timeout: time.Second * 120, Transport: DefaultTransport}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, fmt.Errorf("[findmypk_test:httpget] %s", err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return result, fmt.Errorf("[findmypk_test:httpget] %s", err)
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("[findmypk_test:httpget] %s", err)
	}
	resp.Body.Close()

	return result, nil
}

func TestFMPKSimple(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
	}
	localendpoint := fmt.Sprintf("%s.wolk.com:%s", hostname, server.DefaultPort)

	var reports []backend.FMReport
	var hashKeys [][]byte
	for i := 0; i < 10; i++ {
		key := make([]byte, 16)
		rand.Read(key)
		hashKey := backend.Computehash(key)
		hashKeys = append(hashKeys, hashKey)
		symptom := "sample symptom"

		report := backend.FMReport{HashedPK: hashKey, EncodedMsg: []byte(symptom)}
		reports = append(reports, report)
	}
	fmReportJSON, err := json.Marshal(reports)
	_, err = httppost(fmt.Sprintf("https://%s/%s", localendpoint, server.EndpointFMReport), fmReportJSON)
	if err != nil {
		t.Fatalf("EndpointFMReport: %s", err)
	}

	var prefixHashedKey []byte
	sampleKey := hashKeys[3]
	sampleKey2 := hashKeys[6]
	prefixHashedKey = append(prefixHashedKey, sampleKey[0])
	prefixHashedKey = append(prefixHashedKey, sampleKey[1])
	prefixHashedKey = append(prefixHashedKey, (sampleKey[2]&0xC0)|(sampleKey2[0]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[0]&03<<6)|(sampleKey2[1]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[1]&03<<6)|(sampleKey2[2]&0xFC>>2))

	result, err := httppost(fmt.Sprintf("https://%s/%s/%d", localendpoint, server.EndpointFMQuery, time.Now().Unix()), prefixHashedKey)
	if err != nil {
		t.Fatalf("EndpointFMReport: %s", err)
	}

	var resultreport []*backend.FMReport
	err = json.Unmarshal(result, &resultreport)
	if err != nil {
		t.Fatalf("EndpointFMReport(check1): %s", err)
	}
	for _, r := range resultreport {
		if bytes.Compare(r.HashedPK, sampleKey) == 0 || bytes.Compare(r.HashedPK, sampleKey2) == 0 {
			fmt.Println("ok")
		} else {
			fmt.Println("err")
		}
	}
}

func TestFMPKLong(t *testing.T) {
	var reports []backend.FMReport
	var hashKeys [][]byte
	key := make([]byte, 16)
	msg := make([]byte, 128)
	timeStart := time.Now()
	for reportNum := 0; reportNum < 10; reportNum++ {
		for i := 0; i < 100; i++ {
			rand.Read(key)
			hashKey := backend.Computehash(key)
			hashKeys = append(hashKeys, hashKey)
			rand.Read(msg)

			report := backend.FMReport{HashedPK: hashKey, EncodedMsg: msg}
			reports = append(reports, report)
		}
		fmReportJSON, err := json.Marshal(reports)
		timeReportStart := time.Now()
		_, err = httppost(fmt.Sprintf("https://%s/%s", endpoint, server.EndpointFMReport), fmReportJSON)
		fmt.Printf("request %d time %v\n", reportNum, time.Since(timeReportStart))
		if err != nil {
			t.Fatalf("EndpointFMReport: %s", err)
		}
	}
	fmt.Printf("request totaltime = %v\n", time.Since(timeStart))

	queryTimeTotalStart := time.Now()
	for queryNum := 0; queryNum < 10; queryNum++ {
		var prefixHashedKey []byte
		sampleKey := hashKeys[rand.Intn(100)]
		sampleKey2 := hashKeys[rand.Intn(100)]
		prefixHashedKey = append(prefixHashedKey, sampleKey[0])
		prefixHashedKey = append(prefixHashedKey, sampleKey[1])
		prefixHashedKey = append(prefixHashedKey, (sampleKey[2]&0xC0)|(sampleKey2[0]&0xFC>>2))
		prefixHashedKey = append(prefixHashedKey, (sampleKey2[0]&03<<6)|(sampleKey2[1]&0xFC>>2))
		prefixHashedKey = append(prefixHashedKey, (sampleKey2[1]&03<<6)|(sampleKey2[2]&0xFC>>2))

		queryTimeStart := time.Now()
		result, err := httppost(fmt.Sprintf("https://%s/%s/%d", endpoint, server.EndpointFMQuery, time.Now().Unix()), prefixHashedKey)
		fmt.Printf("query %d time %v\n", queryNum, time.Since(queryTimeStart))
		if err != nil {
			t.Fatalf("EndpointFMReport: %s", err)
		}

		var resultreport []*backend.FMReport
		err = json.Unmarshal(result, &resultreport)
		if err != nil {
			t.Fatalf("EndpointFMReport(check1): %s", err)
		}
		for _, r := range resultreport {
			if bytes.Compare(r.HashedPK, sampleKey) == 0 || bytes.Compare(r.HashedPK, sampleKey2) == 0 {
				//fmt.Println("ok")
			} else {
				fmt.Println("err")
			}
		}
	}
	fmt.Printf("query total%v\n", time.Since(queryTimeTotalStart))
}
