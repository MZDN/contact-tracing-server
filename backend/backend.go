package backend

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	//"strings"
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
	HashedPK   []byte `json:"hashedPK"`
	EncodedMsg []byte `json:"encodedMsg"`
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

func (backend *Backend) ProcessReport(reports []CENReport) (err error) {
	sReport := "insert into CENReport (hashedPK, encodedMsg, reportTS, prefixHashedPK) values ( ?, ?, ?, ? ) on duplicate key update report = values(report)"
	stmtReport, err := backend.db.Prepare(sReport)
	if err != nil {
		return err
	}

	curTS := uint64(time.Now().Unix())
	for _, report := range reports {
		var prefixHashedPK []byte
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[0])
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[1])
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[2]&0xC0)
		_, err = stmtReport.Exec(report.HashedPK, report.EncodedMsg, curTS, prefixHashedPK)
		if err != nil {
			// return or revart or continue
			//	return err
		}
	}
	return nil
}

func (backend *Backend) ProcessQuery(query []byte, timestamp uint64) (reports []CENReport, err error) {
	sQuery := "SELECT (hashedPK, encodedMsg) from CENReport where prefixHashedPK = ?"
	stmt, err := backend.db.Prepare(sQuery)
	if err != nil {
		return reports, err
	}
	querylen := len(query)
	posStartByte := 0
	posStartBit := 0
	for i := 0; ; i++ {
		if i != 0 {
			posStartByte = (18 * i) / 8
			posStartBit = (18 * i) % 8
		}
		if posStartByte >= querylen {
			break
		}

		var prefixedHashedKey []byte
		switch posStartBit {
		case 0:
		case 2:
		case 4:
		case 6:
		}

		rows, err := stmt.Query(prefixedHashedKey)
		if err != nil {
			return reports, err
		}
		for rows.Next() {
			var r CENReport
			err = rows.Scan(&(r.HashedPK), &(r.EncodedMsg))
			if err != nil {
				return reports, err
			}
			reports = append(reports, r)
		}
	}
	return reports, err
}
