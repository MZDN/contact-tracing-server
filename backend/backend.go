package backend

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	// TableCENKeys stores the mapping between CENKeys and CENReports.
	TableCENKeys = "CENKeys"

	// TableCENKeys stores the mapping between CENKeys and CENReports.
	TableCENReport = "CENReport"
)

// Backend holds a client to connect  to the BigTable backend
type Backend struct {
	db *sql.DB
}

// CENReport payload is sent by client to /cenreport when user reports symptoms
type CENReport struct {
	ReportID        string `json:"reportID,omitempty"`
	Report          []byte `json:"report,omitempty"`  // this is expected to be a JSON blob but the server doesn't need to parse it
	CENKeys         string `json:"cenKeys,omitempty"` // comma separated list of base64 AES Keys
	ReportMimeType  string `json:"reportMimeType,omitempty"`
	ReportTimeStamp uint64 `json:"reportTimeStamp,omitempty"`
	StoredTimeStamp uint64 `json:"storedTimeStamp,omitempty"`
}

type CENReportMeta struct {
	Report         [][]byte `json:"report,omitempty"`
	ReportMetaData string   `json:"reportMetaData,omitempty"`
	L              string   `json:"l,omitempty"`
	key            string   `json:"key,omitempty"`
	j              int      `json:"j,omitempty"`
	jMax           int      `json:"jMax,omitempty"`
}

type CENSymptom struct {
	ReportID        string `json:"reportID,omitempty"`
	Symptoms        []int  `json:"symptoms,omitempty"`
	CENKeys         string `json:"cenKeys,omitempty"` // comma separated list of base64 AES Keys
	ReportMimeType  string `json:"reportMimeType,omitempty"`
	ReportTimeStamp uint64 `json:"reportTimeStamp,omitempty"`
	StoredTimeStamp uint64 `json:"storedTimeStamp,omitempty"`
}

type CENStatus struct {
	ReportID        string `json:"reportID,omitempty"`
	StatusID        int    `json:"statusID,omitempty"`
	CENKeys         string `json:"cenKeys,omitempty"` // comma separated list of base64 AES Keys
	ReportMimeType  string `json:"reportMimeType,omitempty"`
	ReportTimeStamp uint64 `json:"reportTimeStamp,omitempty"`
	StoredTimeStamp uint64 `json:"storedTimeStamp,omitempty"`
}

type CENSymptomReport struct {
	HashedPK []byte `json:"hashedPK"`
	EncMsg   []byte `json:"encMsg"`
}

type CENSymptomRequest struct {
	PrefixBitVector []byte `json:"prefixBitVector"`
}

type Config struct {
	MysqlConn string `json:"mysqlConn,omitempty"`
}

// NewBackend sets up a client connection to BigTable to manage incoming payloads
//func NewBackend(mysqlConnectionString string) (backend *Backend, err error) {
func NewBackend(conf *Config) (backend *Backend, err error) {
	mysqlconn := os.Getenv("MYSQLCONN")
	if mysqlconn == "" {
		mysqlconn = conf.MysqlConn
	}
	mysqlConnectionString := flag.String("conn", mysqlconn, "MySQL Connection String")

	backend = new(Backend)
	backend.db, err = sql.Open("mysql", *mysqlConnectionString)
	if err != nil {
		return backend, err
	}

	return backend, nil
}

// ProcessCENReport manages the API Endpoint to POST /cenreport
//  Input: CENReport
//  Output: error
//  Behavior: write report bytes to "report" table; write row for each CENKey with reportID
func (backend *Backend) ProcessCENReport(cenReport *CENReport) (err error) {
	log.Println("ProcessCENReport 1")
	reportData, err := json.Marshal(cenReport)
	if err != nil {
		return err
	}
	log.Println("ProcessCENReport 2")

	/*
		// put the CENReport in CENKeys table
		sKeys := "insert into CENKeys (cenKey, reportID, reportTS) values ( ?, ?, ? ) on duplicate key update reportTS = values(reportTS)"
		stmtKeys, err := backend.db.Prepare(sKeys)
		if err != nil {
			return err
		}
	*/

	// put the CENReport in CENReport table
	sReport := "insert into CENReport (reportID, report, reportMimeType, reportTS, storeTS) values ( ?, ?, ?, ?, ? ) on duplicate key update report = values(report)"
	stmtReport, err := backend.db.Prepare(sReport)
	if err != nil {
		return err
	}

	curTS := uint64(time.Now().Unix())
	reportID := fmt.Sprintf("%x", Computehash(reportData))
	// store the cenreportID in cenkeys table, one row per key
	/*
		for _, cenKey := range cenKeys {
			cenKey := strings.Trim(cenKey, " \n")
			if len(cenKey) > 30 && len(cenKey) <= 32 {
				_, err = stmtKeys.Exec(cenKey, reportID, curTS)
				if err != nil {
					return err
				}
			}
		}
	*/

	// store the cenreportID in cenReport table, one row per key
	log.Println("ProcessCENReport 3")
	_, err = stmtReport.Exec(reportID, cenReport.Report, cenReport.ReportMimeType, cenReport.ReportTimeStamp, curTS)
	if err != nil {
		panic(5)
		return err
	}
	log.Println("ProcessCENReport 4")

	symptom := new(CENSymptom)
	symptom.ReportID = reportID
	symptom.ReportMimeType = cenReport.ReportMimeType
	symptom.ReportTimeStamp = cenReport.ReportTimeStamp
	symptom.StoredTimeStamp = curTS
	symptom.Symptoms, _ = backend.getSymptoms(cenReport.Report)
	backend.ProcessCENSymptom(symptom)
	log.Println("ProcessCENReport 5")

	log.Println("call status")
	status := new(CENStatus)
	status.ReportID = cenReport.ReportID
	status.ReportMimeType = cenReport.ReportMimeType
	status.ReportTimeStamp = cenReport.ReportTimeStamp
	status.StoredTimeStamp = curTS
	status.StatusID, _ = backend.getStatusID(cenReport.Report)
	backend.ProcessCENStatus(status)

	return nil
}

// ProcessCENReport manages the API Endpoint to POST /cenreport
//  Input: CENReport
//  Output: error
//  Behavior: write report bytes to "report" table; write row for each CENKey with reportID
func (backend *Backend) ProcessCENSymptom(cenSymptom *CENSymptom) (err error) {
	if cenSymptom.ReportID == "" {
		symptomData, err := json.Marshal(cenSymptom)
		if err != nil {
			return err
		}
		cenSymptom.ReportID = fmt.Sprintf("%x", Computehash(symptomData))
	}

	/*
		// put the CENReport in CENKeys table
		sKeys := "insert into CENKeys (cenKey, reportID, reportTS) values ( ?, ?, ? ) on duplicate key update reportTS = values(reportTS)"
		stmtKeys, err := backend.db.Prepare(sKeys)
		if err != nil {
			return err
		}
	*/

	// put the CENReport in CENReport table
	sReport := "insert into CENSymptom (reportID, symptomID, reportMimeType, reportTS, storeTS) values ( ?, ?, ?, ?, ?) "
	stmtReport, err := backend.db.Prepare(sReport)
	if err != nil {
		return err
	}

	/*
		cenKeys := strings.Split(cenSymptom.CENKeys, ",")
		// store the cenreportID in cenkeys table, one row per key
		for _, cenKey := range cenKeys {
			cenKey := strings.Trim(cenKey, " \n")
			if len(cenKey) > 30 && len(cenKey) <= 32 {
				_, err = stmtKeys.Exec(cenKey, reportID, curTS)
				if err != nil {
					return err
				}
			}
		}
	*/

	// store the cenreportID in cenReport table, one row per key
	if cenSymptom.StoredTimeStamp == 0 {
		cenSymptom.StoredTimeStamp = uint64(time.Now().Unix())
	}
	for _, symptomid := range cenSymptom.Symptoms {
		_, err = stmtReport.Exec(cenSymptom.ReportID, symptomid, cenSymptom.ReportMimeType, cenSymptom.ReportTimeStamp, cenSymptom.StoredTimeStamp)
		if err != nil {
			panic(5)
			return err
		}
	}

	return nil
}

func (backend *Backend) ProcessCENStatus(cenStatus *CENStatus) (err error) {
	if cenStatus.ReportID == "" {
		statusData, err := json.Marshal(cenStatus)
		if err != nil {
			return err
		}
		cenStatus.ReportID = fmt.Sprintf("%x", Computehash(statusData))
	}

	/*
		// put the CENReport in CENKeys table
		sKeys := "insert into CENKeys (cenKey, reportID, reportTS) values ( ?, ?, ? ) on duplicate key update reportTS = values(reportTS)"
		stmtKeys, err := backend.db.Prepare(sKeys)
		if err != nil {
			return err
		}
	*/

	// put the CENReport in CENReport table
	sReport := "insert into CENStatus (reportID, statusID, reportMimeType, reportTS) values ( ?, ?, ?, ?) "
	stmtReport, err := backend.db.Prepare(sReport)
	if err != nil {
		return err
	}

	/*
		cenKeys := strings.Split(cenStatus.CENKeys, ",")
		// store the cenreportID in cenkeys table, one row per key
		for _, cenKey := range cenKeys {
			cenKey := strings.Trim(cenKey, " \n")
			if len(cenKey) > 30 && len(cenKey) <= 32 {
				_, err = stmtKeys.Exec(cenKey, reportID, curTS)
				if err != nil {
					return err
				}
			}
		}
	*/

	if cenStatus.StoredTimeStamp == 0 {
		cenStatus.StoredTimeStamp = uint64(time.Now().Unix())
	}

	// store the cenreportID in cenReport table, one row per key
	_, err = stmtReport.Exec(cenStatus.ReportID, cenStatus.StatusID, cenStatus.ReportMimeType, cenStatus.ReportTimeStamp, cenStatus.StoredTimeStamp)
	if err != nil {
		panic(5)
		return err
	}

	return nil
}

// ProcessGetCENKeys manages the GET API endpoint /cenkeys
//  Input: timestamp
//  Output: array of CENKeys (in string form) for the last hour
func (backend *Backend) ProcessGetCENKeys(timestamp uint64) (cenKeys []string, err error) {
	cenKeys = make([]string, 0)

	s := "select cenKey From CENKeys where ReportTS >= 0" // TODO: ReportTS > ? and ReportTS <= ?"
	stmt, err := backend.db.Prepare(s)
	if err != nil {
		return cenKeys, err
	}
	rows, err := stmt.Query() // TODO: timestamp-3600, timestamp
	if err != nil {
		return cenKeys, err
	}
	for rows.Next() {
		var cenKey string
		err = rows.Scan(&cenKey)
		if err != nil {
			return cenKeys, err
		}
		cenKeys = append(cenKeys, cenKey)
	}
	return cenKeys, nil
}

// ProcessGetCENReport manages the POST API endpoint /cenreport
//  Input: cenKey
//  Output: array of CENReports
func (backend *Backend) ProcessGetCENReport(cenKey string) (reports []*CENReport, err error) {
	reports = make([]*CENReport, 0)

	s := fmt.Sprintf("select CENKeys.reportID, Report, reportMimeType, CENReport.reportTS From CENKeys, CENReport where CENKeys.CENKey = ? and CENKeys.reportID = CENReport.reportID")
	stmt, err := backend.db.Prepare(s)
	if err != nil {
		return reports, err
	}
	rows, err := stmt.Query(cenKey)
	if err != nil {
		return reports, err
	}
	for rows.Next() {
		var r CENReport
		err = rows.Scan(&(r.ReportID), &(r.Report), &(r.ReportMimeType), &(r.ReportTimeStamp))
		if err != nil {
			return reports, err
		}
		reports = append(reports, &r)
	}
	return reports, nil
}

func (backend *Backend) getStatusID(report []byte) (int, error) {
	var dat map[string]interface{}

	if err := json.Unmarshal(report, &dat); err != nil {
		return 0, err
	}
	/*
		    if ok := dat["reportMetadata"]; !ok{
			    return 0, nil
		    }
	*/
	log.Printf("getStatusID %v\n", dat)
	reportstr := strings.ToLower(dat["reportMetadata"].(string))
	log.Printf("status = %v", dat)
	if strings.Contains(reportstr, "negative") {
		return 2, nil
	} else if strings.Contains(reportstr, "positive") || strings.Contains(reportstr, "covid") {
		return 1, nil
	} else if strings.Contains(reportstr, "recovered") {
		return 4, nil
	}
	return 0, nil
}

func (backend *Backend) getSymptoms(report []byte) ([]int, error) {
	var dat map[string]interface{}

	if err := json.Unmarshal(report, &dat); err != nil {
		panic(err)
	}
	log.Printf("status = %v", dat)
	return []int{1, 2, 3}, nil
}

// Computehash returns the hash of its inputs
func Computehash(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, b := range data {
		_, err := hasher.Write(b)
		if err != nil {
			panic(1)
		}
	}
	return hasher.Sum(nil)
}

func makeCENKeyString() string {
	key := make([]byte, 16)
	rand.Read(key)
	encoded := fmt.Sprintf("%x", key)
	return encoded
}

// GetSampleCENReportAndCENKeys generates a CENReport and an array of CENKeys (in string form)
func GetSampleCENReportAndCENKeys(nKeys int) (cenReport *CENReport, cenKeys []string) {
	cenKeys = make([]string, nKeys)
	for i := 0; i < nKeys; i++ {
		cenKeys[i] = makeCENKeyString()
	}
	CENKeys := fmt.Sprintf("%s,%s", cenKeys[0], cenKeys[1])
	cenReport = new(CENReport)
	cenReport.ReportID = "1"
	cenReport.Report = []byte("severe fever,coughing,hard to breathe")
	cenReport.CENKeys = CENKeys
	return cenReport, cenKeys
}

// GetSampleCENReportAndCENKeys generates a CENReport and an array of CENKeys (in string form)
func GetSampleCENSymptomAndCENKeys(nKeys int) (cenSymptom *CENSymptom, cenKeys []string) {
	cenKeys = make([]string, nKeys)
	for i := 0; i < nKeys; i++ {
		cenKeys[i] = makeCENKeyString()
	}
	CENKeys := fmt.Sprintf("%s,%s", cenKeys[0], cenKeys[1])
	cenSymptom = new(CENSymptom)
	cenSymptom.ReportID = "1"
	cenSymptom.Symptoms = []int{1, 2, 4, 5, 7}
	cenSymptom.CENKeys = CENKeys
	return cenSymptom, cenKeys
}

// GetSampleCENReportAndCENKeys generates a CENReport and an array of CENKeys (in string form)
func GetSampleCENStatusAndCENKeys(nKeys int) (cenStatus *CENStatus, cenKeys []string) {
	cenKeys = make([]string, nKeys)
	for i := 0; i < nKeys; i++ {
		cenKeys[i] = makeCENKeyString()
	}
	CENKeys := fmt.Sprintf("%s,%s", cenKeys[0], cenKeys[1])
	cenStatus = new(CENStatus)
	cenStatus.ReportID = "1"
	cenStatus.StatusID = 1
	cenStatus.CENKeys = CENKeys
	return cenStatus, cenKeys
}

func (backend *Backend) ProcessSymptomReport(report *CENSymptomReport) (err error) {
}

func (backend *Backend) ProcessSymptomReport(prefixBitVector []byte) (reportid uint64) {
	return 1
}

func (backend *Backend) GetSymptomResult(reportid uint64) (results []CENSymptomReport, err error) {
}

func (backend *Backend) GetSymptomReportStatus(reportid uint64) (statusid uint64) {
}
