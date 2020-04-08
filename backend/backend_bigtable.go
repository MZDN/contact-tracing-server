package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/bigtable"
)

type Backend struct {
	client           *bigtable.Client
	tableName        string
	columnFamilyName string
}

func NewBackend(conf *Config) (backend *Backend, err error) {
	ctx := context.Background()
	backend = new(Backend)

	client, err := bigtable.NewClient(ctx, conf.BigtableProject, conf.BigtableInstance)
	if err != nil {
		log.Printf("bigtable err %v\n", err)
		return backend, err
	}
	backend.columnFamilyName = "CENReport"
	backend.tableName = "CENReport"
	backend.client = client
	return backend, nil
}

func (backend *Backend) ProcessReport(reports []CENReport) (err error) {

	// when should open is called?
	table := backend.client.Open(backend.tableName)

	timestamp := bigtable.Now()
	for _, report := range reports {
		PrefixHashedKey := report.HashedPK[:3]
		mut := bigtable.NewMutation()
		mut.Set(backend.columnFamilyName, "EncodedMsg", timestamp, report.EncodedMsg)
		mut.Set(backend.columnFamilyName, "HashedPK", timestamp, report.HashedPK)
		err = table.Apply(context.Background(), fmt.Sprintf("%x", PrefixHashedKey), mut)
		if err != nil {
			log.Println("ProcessReport:bigTable Apply Error", err)
		}
	}

	return err
}

func (backend *Backend) ProcessQuery(query []byte, timestamp int64) (reports []CENReport, err error) {
	prefixkeys := make(bigtable.RowList, 0)
	// TODO: split query into H(PK) prefixes
	for q := 0; q < len(query); q += 3 {
		prefixkey := fmt.Sprintf("%x", query[q:(q+3)])
		prefixkeys = append(prefixkeys, prefixkey)
	}
	/* 18 bit support
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
	}
	*/

	ctx := context.Background()
	startTime := time.Unix(timestamp, 0)
	fmt.Println(startTime)
	endTime := time.Now()
	fmt.Println(endTime)

	filter := bigtable.ChainFilters(bigtable.FamilyFilter(backend.columnFamilyName), bigtable.TimestampRangeFilter(startTime, endTime))

	table := backend.client.Open(backend.tableName)
	err = table.ReadRows(ctx, prefixkeys,
		func(row bigtable.Row) bool {
			//for columnFamily, cols := range row {
			for _, cols := range row {
				var report CENReport
				for _, col := range cols {
					dt := strings.Split(col.Column, ":")
					fmt.Println("dt", dt)
					fmt.Println("col", col.Value)
					switch dt[1] {
					case "EncodedMsg":
						report.EncodedMsg = []byte(col.Value)
					case "HashedPK":
						report.HashedPK = col.Value
					case "timestamp":
					default:
					}
				}
				reports = append(reports, report)
			}
			return true
			//		}, bigtable.RowFilter(bigtable.TimestampRangeFilter(startTime, endTime)))
		}, bigtable.RowFilter(filter))

	return reports, err
}
