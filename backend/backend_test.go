package backend

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestBackendReportQuery(t *testing.T) {
	config := new(Config)
	config.BigtableProject = "us-west1-wlk"
	config.BigtableInstance = "findmypk"

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
	time.Sleep(time.Second * 3)
	scantime := time.Now()
	time.Sleep(time.Second * 3)
	fmt.Println("scantime", scantime.UnixNano())
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
	sampleKey := hashKeys[3][:3]
	sampleKey2 := hashKeys[6][:3]
	sampleKey3 := hashKeys[13][:3]
	sampleKey4 := hashKeys[16][:3]
	prefixHashedKey = append(prefixHashedKey, sampleKey...)
	prefixHashedKey = append(prefixHashedKey, sampleKey2...)
	prefixHashedKey = append(prefixHashedKey, sampleKey3...)
	prefixHashedKey = append(prefixHashedKey, sampleKey4...)
	time.Sleep(time.Second * 3)

	res, err := backend.ProcessQuery(prefixHashedKey, scantime.Unix())
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
