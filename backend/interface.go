package backend

// FMReport payload is sent by client to /fmreport when user reports symptoms
type FMReport struct {
	HashedPK   []byte `json:"hashedPK"`
	EncodedMsg []byte `json:"encodedMsg"`
}

type FMSymptom struct {
	ReportID        string `json:"reportID,omitempty"`
	Symptoms        []int  `json:"symptoms,omitempty"`
	FMKeys          string `json:"fmKeys,omitempty"` // comma separated list of base64 AES Keys
	ReportMimeType  string `json:"reportMimeType,omitempty"`
	ReportTimeStamp uint64 `json:"reportTimeStamp,omitempty"`
	StoredTimeStamp uint64 `json:"storedTimeStamp,omitempty"`
}

type FMStatus struct {
	ReportID        string `json:"reportID,omitempty"`
	StatusID        int    `json:"statusID,omitempty"`
	FMKeys          string `json:"fmKeys,omitempty"` // comma separated list of base64 AES Keys
	ReportMimeType  string `json:"reportMimeType,omitempty"`
	ReportTimeStamp uint64 `json:"reportTimeStamp,omitempty"`
	StoredTimeStamp uint64 `json:"storedTimeStamp,omitempty"`
}

type Config struct {
	MysqlConn        string `json:"mysqlConn,omitempty"`
	BigtableProject  string `json:"bigtableProject,omitempty"`
	BigtableInstance string `json:"bigtableInstance,omitempty"`
}
