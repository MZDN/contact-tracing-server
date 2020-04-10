package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/bigtable"
)

const threadsPerRequest = 10

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

func (backend *Backend) ProcessReport(reports []CTReport) (err error) {
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
	// will move this process to Loop
	errs, err := backend.table.ApplyBulk(context.Background(), keys, muts)
	if err != nil {
		log.Printf("backend.table.ApplyBulk err %v %v\n", errs, err)
	} else {
		log.Printf("backend.table.ApplyBulk uploaded %d\n", len(keys))
	}
	return err
}

type reportResult struct {
	results []CTReport
	err     error
}

func (backend *Backend) ProcessQuery(query []byte, timestamp int64) (reports []CTReport, err error) {
	fmt.Println("******** ProcessQuery ***********", len(query))
	var prefixKeyList []bigtable.RowList
	prefixkeys := make(bigtable.RowList, 0)
	// TODO: split query into H(PK) prefixes
	keyLength := len(query) / 3
	threadNum := keyLength/1000 + 1
	for q := 0; q < len(query); q += 3 {
		prefixkey := fmt.Sprintf("%x", query[q:(q+3)])
		prefixkeys = append(prefixkeys, prefixkey)
		if len(prefixkeys)%1000 == 1 && q != 0 {
			prefixKeyList = append(prefixKeyList, prefixkeys)
			prefixkeys = make(bigtable.RowList, 0)
		}
		fmt.Println("len list", len(prefixkeys))
	}
	fmt.Println("len list b", len(prefixkeys))
	if len(prefixkeys) > 0 {
		prefixKeyList = append(prefixKeyList, prefixkeys)
	}

	ctx := context.Background()
	startTime := time.Unix(timestamp, 0)
	fmt.Println(startTime)
	endTime := time.Now()
	fmt.Println(endTime)

	filter := bigtable.ChainFilters(bigtable.FamilyFilter(backend.columnFamilyName), bigtable.TimestampRangeFilter(startTime, endTime))

	resCh := make(chan *reportResult)
	threadLimit := make(chan struct{}, threadsPerRequest)
	for i := 0; i < threadNum; i++ {
		fmt.Println("creating thread", i)
		threadLimit <- struct{}{}
		go func(prefixkeys bigtable.RowList) {
			threadResult := new(reportResult)
			threadResult.err = backend.table.ReadRows(ctx, prefixkeys,
				func(row bigtable.Row) bool {
					//for columnFamily, cols := range row {
					for _, cols := range row {
						var report CTReport
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
						threadResult.results = append(threadResult.results, report)
					}
					return true
					//		}, bigtable.RowFilter(bigtable.TimestampRangeFilter(startTime, endTime)))
				}, bigtable.RowFilter(filter))
			resCh <- threadResult
			<-threadLimit
		}(prefixKeyList[i])
	}

	for i := 0; i < threadNum; i++ {
		r := <-resCh
		if r.err != nil {
			// TODO: how to handle errors
			err = r.err
		}
		if r.results != nil {
			reports = append(reports, r.results...)
		}
	}

	return reports, err
}

func (backend *Backend) ProcessSync(timestamp int64) (reports []CTReport, err error) {
	// is there any limitation of the size of data?
	startTime := time.Unix(timestamp, 0)
	fmt.Println(startTime)
	endTime := time.Now()
	fmt.Println(endTime)
	filter := bigtable.ChainFilters(bigtable.FamilyFilter(backend.columnFamilyName), bigtable.TimestampRangeFilter(startTime, endTime))

	resCh := make(chan *reportResult)
	ctx := context.Background()
	for i := 0; i < 16; i++ {
		fmt.Println("creating thread", i)
		go func(pos int) {
			threadResult := new(reportResult)
			threadResult.err = backend.table.ReadRows(ctx, bigtable.PrefixRange(fmt.Sprintf("%x", pos)),
				func(row bigtable.Row) bool {
					//for columnFamily, cols := range row {
					for _, cols := range row {
						var report CTReport
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
						threadResult.results = append(threadResult.results, report)
					}
					return true
				}, bigtable.RowFilter(filter))
			resCh <- threadResult
		}(i)
	}

	for i := 0; i < 16; i++ {
		r := <-resCh
		if r.err != nil {
			// TODO: how to handle errors
			err = r.err
		}
		if r.results != nil {
			reports = append(reports, r.results...)
		}
	}

	return
}
