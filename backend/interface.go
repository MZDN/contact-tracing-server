package backend

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
	MysqlConn        string `json:"mysqlConn,omitempty"`
	BigtableProject  string `json:"bigtableProject,omitempty"`
	BigtableInstance string `json:"bigtableInstance,omitempty"`
}
