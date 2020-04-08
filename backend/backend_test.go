package backend

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestBackendReportQuery(t *testing.T) {
	config := new(Config)
	config.MysqlConn = "mayumi:c0v1d19w9lk3rmm!@tcp(34.83.154.244)/cenmm?charset=utf8"
	backend, err := NewBackend(config)
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	var reports []CENReport
	var hashKeys [][]byte
	for i := 0; i < 10; i++ {
		key := make([]byte, 16)
		rand.Read(key)
		hashKey := Computehash(key)
		hashKeys = append(hashKeys, hashKey)
		symptom := "sample symptom"

		report := CENReport{HashedPK: hashKey, EncodedMsg: []byte(symptom)}
		reports = append(reports, report)
	}
	err = backend.ProcessReport(reports)
	if err != nil {
		t.Fatal(err)
	}
	var prefixHashedKey []byte
	sampleKey := hashKeys[3]
	sampleKey2 := hashKeys[6]
	prefixHashedKey = append(prefixHashedKey, sampleKey[0])
	prefixHashedKey = append(prefixHashedKey, sampleKey[1])
	prefixHashedKey = append(prefixHashedKey, (sampleKey[2]&0xC0)|(sampleKey2[0]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[0]&03<<6)|(sampleKey2[1]&0xFC>>2))
	prefixHashedKey = append(prefixHashedKey, (sampleKey2[1]&03<<6)|(sampleKey2[2]&0xFC>>2))

	timestamp := uint64(time.Now().Unix())
	res, err := backend.ProcessQuery(prefixHashedKey, timestamp)
	wf, err := os.Create("/tmp/test.txt")
	defer wf.Close()
	wf.Write(prefixHashedKey)
	for _, r := range res {
		fmt.Printf("key = %x report = %s\n", r.HashedPK, r.EncodedMsg)
	}
}

/*
func TestBackendSimple(t *testing.T) {
	cenReport, cenReportKeys := GetSampleCENReportAndCENKeys(2)
	cenReportJSON, err := json.Marshal(cenReport)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	fmt.Printf("CENReportJSON Sample: %s\n", cenReportJSON)

	for i, cenReportKey := range cenReportKeys {
		fmt.Printf("CENKey %d: %s\n", i, cenReportKey)
	}

	backend, err := NewBackend(DefaultConnString)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// submit CEN Report
	err = backend.ProcessCENReport(cenReport)
	if err != nil {
		t.Fatalf("ProcessCENReport: %s", err)
	}

	// get recent CENKeys (last 10 seconds)
	curTS := uint64(time.Now().Unix())
	cenKeys, err := backend.ProcessGetCENKeys(curTS - 10)
	if err != nil {
		t.Fatalf("ProcessGetCENKeys(check1): %s", err)
	}
	fmt.Printf("ProcessGetCENKeys %d records\n", len(cenKeys))
	if len(cenKeys) < 1 {
		t.Fatalf("ProcessGetCENKeys: %d records found", len(cenKeys))
	}
	// check if the first 2 are are in the report
	found := make([]bool, len(cenKeys))
	for _, cenKey := range cenKeys {
		fmt.Printf("Recent Keys: [%s]\n", cenKey)
		for j, reportKey := range cenReportKeys {
			if cenKey == reportKey {
				found[j] = true
			}
		}
	}

	// get the report data from the first two keys
	for i := 0; i < 2; i++ {
		if !found[i] {
			t.Fatalf("ProcessGetCENKeys key 0 in report [%s] not found", cenReportKeys[0])
		}
		cenKey := cenKeys[i]
		reports, err := backend.ProcessGetCENReport(cenKey)
		if err != nil {
			t.Fatalf("ProcessGetCENReport: %s", err)
		}
		if len(reports) > 0 {
			report := reports[0]
			if !bytes.Equal(report.Report, cenReport.Report) {
				t.Fatalf("ProcessGetCENReport Report Mismatch: expected %s, got [%s]", report.Report, cenReport.Report)
			}
			fmt.Printf("ProcessGetCENReport SUCCESS (%s): [%s]\n", cenKey, report.Report)
		} else {
			t.Fatalf("ProcessGetCENReport: Report not found for cenKey %s\n", cenKey)
		}
	}

}
*/
