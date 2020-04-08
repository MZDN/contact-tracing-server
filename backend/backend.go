package backend

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	//"flag"
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
	//mysqlConnectionString := flag.String("conn", mysqlconn, "MySQL Connection String")

	backend = new(Backend)
	backend.db, err = sql.Open("mysql", mysqlconn)
	if err != nil {
		log.Println("err")
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

func makeCENKey() string {
        key := make([]byte, 16)
        rand.Read(key)
        encoded := fmt.Sprintf("%x", key)
        return encoded
}

func (backend *Backend) ProcessReport(reports []CENReport) (err error) {
	sReport := "insert into CENReport (hashedPK, encodedMsg, reportTS, prefixHashedPK) values ( ?, ?, ?, ? ) on duplicate key update hashedPK = values(hashedPK)"
	stmtReport, err := backend.db.Prepare(sReport)
	if err != nil {
		log.Println("ProcessReport sql error", err)
		return err
	}

	curTS := uint64(time.Now().Unix())
	for _, report := range reports {
		var prefixHashedPK []byte
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[0])
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[1])
		prefixHashedPK = append(prefixHashedPK, report.HashedPK[2]&0xC0)
		_, err = stmtReport.Exec(fmt.Sprintf("%x", report.HashedPK), fmt.Sprintf("%x", report.EncodedMsg), curTS, fmt.Sprintf("%x", prefixHashedPK))
		if err != nil {
			log.Println("ProcessReport sql error ", err)
			// return or revart or continue
			//	return err
		}
	}
	return nil
}

func (backend *Backend) ProcessQuery(query []byte, timestamp uint64) (reports []CENReport, err error) {
	sQuery := "SELECT hashedPK, encodedMsg from CENReport where prefixHashedPK = ? and reportTS < ?"
	stmt, err := backend.db.Prepare(sQuery)
	if err != nil {
		log.Printf("ProcessQuery DB Prepare error %v", err)
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
		if posStartByte+2 >= querylen {
			break
		}

		var prefixedHashedKey []byte
		switch posStartBit {
		case 0:
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte])
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte+1])
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte+2]&0xC0)
		case 2:
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte]&0x3F<<2)|(query[posStartByte+1]&0xC0>>6))
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte+1]&0x3F<<2)|(query[posStartByte+2]&0xC0>>6))
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte+2]&0x30<<2)
		case 4:
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte]&0x0F<<4)|(query[posStartByte+1]&0xF0>>4))
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte+1]&0x0F<<4)|(query[posStartByte+2]&0xF0>>4))
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte+2]&0x0C<<4)
		case 6:
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte]&0x03<<6)|(query[posStartByte+1]&0xFC>>2))
			prefixedHashedKey = append(prefixedHashedKey, (query[posStartByte+1]&0x03<<6)|(query[posStartByte+2]&0xFC>>2))
			prefixedHashedKey = append(prefixedHashedKey, query[posStartByte+2]&0x03<<6)
		}

		strHashedKey := fmt.Sprintf("%x", prefixedHashedKey)
		rows, err := stmt.Query(strHashedKey, timestamp)
		if err != nil {
			log.Printf("ProcessQuery Query Error %v", err)
			return reports, err
		}
		for rows.Next() {
			var r CENReport
			var hashedPK, encodedMsg string
			err = rows.Scan(&(hashedPK), &(encodedMsg))
			if err != nil {
				log.Printf("ProcessQuery Next Error %v", err)
				return reports, err
			}
			r.HashedPK, err = hex.DecodeString(hashedPK)
			if err != nil {
				log.Printf("ProcessQuery DecodeString Error %v", err)
				continue
			}
			r.EncodedMsg, err = hex.DecodeString(encodedMsg)
			if err != nil {
				log.Printf("ProcessQuery DecodeString Error %v", err)
				continue
			}
			reports = append(reports, r)
		}
	}
	return reports, err
}
