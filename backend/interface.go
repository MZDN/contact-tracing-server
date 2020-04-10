package backend

// CTReport payload is sent by client to /fmreport when user reports symptoms
type CTReport struct {
	HashedPK   []byte `json:"hashedPK"`
	EncodedMsg []byte `json:"encodedMsg"`
}

type Config struct {
	MysqlConn        string `json:"mysqlConn,omitempty"`
	BigtableProject  string `json:"bigtableProject,omitempty"`
	BigtableInstance string `json:"bigtableInstance,omitempty"`
}
