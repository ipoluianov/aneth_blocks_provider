package an

import (
	"math/big"
	"sync"
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/gomisc/logger"
)

type An struct {
	analytics map[string]*Result
	mtx       sync.Mutex
}

var Instance *An

func init() {
	Instance = NewAn()
	Instance.Start()
}

func NewAn() *An {
	var c An
	c.analytics = make(map[string]*Result)
	return &c
}

func (c *An) Start() {
	//go c.ThAn()
}

func (c *An) an(ts *db.TxsByMinutes) {
	count := 0
	for _, item := range ts.Items {
		count += len(item.TXS)
	}
	logger.Println("An::an txs:", count)
	c.anTrCount(ts)
	for i := 0; i < 1; i++ {
		c.anTrValue(ts)
	}
}

func (c *An) GetResult(code string) *Result {
	var res *Result
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if r, ok := c.analytics[code]; ok {
		res = r
	}
	return res
}

func (c *An) ThAn() {
	for {
		dt1 := time.Now()
		logger.Println("")
		logger.Println("---------- an ------------")
		logger.Println("reading transactions")
		ts := c.GetLatestTransactions()
		c.an(ts)
		dt2 := time.Now()
		logger.Println("execution time:", dt2.Sub(dt1).Milliseconds(), "ms")
		logger.Println("--------------------------")
		logger.Println("")

		time.Sleep(3 * time.Second)
	}
}

func (c *An) anTrCount(ts *db.TxsByMinutes) {
	logger.Println("An::anTrCount begin")
	var result Result
	for i := 0; i < len(ts.Items); i++ {
		src := ts.Items[i]
		var item ResultItem
		item.Index = i
		item.DT = src.DT
		item.DTStr = time.Unix(int64(item.DT), 0).UTC().Format("2006-01-02 15:04:05")
		item.Value = float64(len(src.TXS))
		result.Items = append(result.Items, &item)
	}
	result.Count = len(result.Items)
	result.CurrentDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	c.mtx.Lock()
	c.analytics["an"] = &result
	c.mtx.Unlock()
	logger.Println("An::anTrCount end")
}

func (c *An) anTrValue(ts *db.TxsByMinutes) {
	logger.Println("An::anTrValue begin")
	var result Result
	for i := 0; i < len(ts.Items); i++ {
		src := ts.Items[i]
		var item ResultItem
		item.Index = i
		item.DT = src.DT
		item.DTStr = time.Unix(int64(item.DT), 0).UTC().Format("2006-01-02 15:04:05")
		v := big.NewInt(0)
		for _, t := range src.TXS {
			v = v.Add(v, t.TxValue)
		}

		item.Value, _ = v.Float64()

		result.Items = append(result.Items, &item)
	}
	result.Count = len(result.Items)
	result.CurrentDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	c.mtx.Lock()
	c.analytics["vl"] = &result
	c.mtx.Unlock()
	logger.Println("An::anTrValue end")
}
