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

	var reports []FMReport
	var hashKeys [][]byte
	for i := 0; i < 10; i++ {
		key := make([]byte, 16)
		rand.Read(key)
		hashKey := Computehash(key)
		hashKeys = append(hashKeys, hashKey)
		symptom := "sample symptom"

		report := FMReport{HashedPK: hashKey, EncodedMsg: []byte(symptom)}
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

		report := FMReport{HashedPK: hashKey, EncodedMsg: []byte(symptom)}
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
	fmReport, fmReportKeys := GetSampleFMReportAndFMKeys(2)
	fmReportJSON, err := json.Marshal(fmReport)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	fmt.Printf("FMReportJSON Sample: %s\n", fmReportJSON)

	for i, fmReportKey := range fmReportKeys {
		fmt.Printf("FMKey %d: %s\n", i, fmReportKey)
	}

	backend, err := NewBackend(DefaultConnString)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// submit FM Report
	err = backend.ProcessFMReport(fmReport)
	if err != nil {
		t.Fatalf("ProcessFMReport: %s", err)
	}

	// get recent FMKeys (last 10 seconds)
	curTS := uint64(time.Now().Unix())
	fmKeys, err := backend.ProcessGetFMKeys(curTS - 10)
	if err != nil {
		t.Fatalf("ProcessGetFMKeys(check1): %s", err)
	}
	fmt.Printf("ProcessGetFMKeys %d records\n", len(fmKeys))
	if len(fmKeys) < 1 {
		t.Fatalf("ProcessGetFMKeys: %d records found", len(fmKeys))
	}
	// check if the first 2 are are in the report
	found := make([]bool, len(fmKeys))
	for _, fmKey := range fmKeys {
		fmt.Printf("Recent Keys: [%s]\n", fmKey)
		for j, reportKey := range fmReportKeys {
			if fmKey == reportKey {
				found[j] = true
			}
		}
	}

	// get the report data from the first two keys
	for i := 0; i < 2; i++ {
		if !found[i] {
			t.Fatalf("ProcessGetFMKeys key 0 in report [%s] not found", fmReportKeys[0])
		}
		fmKey := fmKeys[i]
		reports, err := backend.ProcessGetFMReport(fmKey)
		if err != nil {
			t.Fatalf("ProcessGetFMReport: %s", err)
		}
		if len(reports) > 0 {
			report := reports[0]
			if !bytes.Equal(report.Report, fmReport.Report) {
				t.Fatalf("ProcessGetFMReport Report Mismatch: expected %s, got [%s]", report.Report, fmReport.Report)
			}
			fmt.Printf("ProcessGetFMReport SUCCESS (%s): [%s]\n", fmKey, report.Report)
		} else {
			t.Fatalf("ProcessGetFMReport: Report not found for fmKey %s\n", fmKey)
		}
	}

}
*/
