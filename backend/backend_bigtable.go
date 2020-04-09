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
	table            *bigtable.Table
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
	backend.table = backend.client.Open(backend.tableName)
	return backend, nil
}

func (backend *Backend) ProcessReport(reports []FMReport) (err error) {

	timestamp := bigtable.Now()
	var keys []string
	var muts []*bigtable.Mutation
	for _, report := range reports {
		prefixHashedKey := fmt.Sprintf("%x", report.HashedPK[:3])
		keys = append(keys, prefixHashedKey)
		mut := bigtable.NewMutation()
		mut.Set(backend.columnFamilyName, "EncodedMsg", timestamp, report.EncodedMsg)
		mut.Set(backend.columnFamilyName, "HashedPK", timestamp, report.HashedPK)
		muts = append(muts, mut)
	}
	//up to a max of 100,000
	errs, err := backend.table.ApplyBulk(context.Background(), keys, muts)
	if err != nil {
		log.Printf("backend.table.ApplyBulk err %v %v\n", errs, err)
	} else {
		log.Printf("backend.table.ApplyBulk uploaded %d\n", len(keys))
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
