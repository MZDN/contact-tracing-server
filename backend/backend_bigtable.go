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
	backend.columnFamilyName = "report"
	backend.tableName = "report"
	backend.client = client
	return backend, nil
}

func (backend *Backend) ProcessReport(reports []FMReport) (err error) {

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

func (backend *Backend) ProcessQuery(query []byte, timestamp int64) (reports []FMReport, err error) {
	prefixkeys := make(bigtable.RowList, 0)
	// TODO: split query into H(PK) prefixes
	for q := 0; q < len(query); q += 3 {
		prefixkey := fmt.Sprintf("%x", query[q:(q+3)])
		prefixkeys = append(prefixkeys, prefixkey)
	}

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
				var report FMReport
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
