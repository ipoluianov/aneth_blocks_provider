package an

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type An struct {
	analytics map[string]*Result
	mtx       sync.Mutex
}

type Tx struct {
	BlockNumber uint64
	BlockDT     uint64
	TxFrom      *common.Address
	TxTo        *common.Address
	TxData      []byte
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
	go c.ThAn()
}

func (c *An) an(ts *TxsByMinutes) {
	c.anTrCount(ts)
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
		fmt.Println("---------- an ------------")
		ts := c.GetLatestTransactions()
		c.an(ts)
		dt2 := time.Now()
		fmt.Println("execution time:", dt2.Sub(dt1).Milliseconds(), "ms")
		fmt.Println("--------------------------")

		time.Sleep(3 * time.Second)
	}
}

func (c *An) anTrCount(ts *TxsByMinutes) {
	var result Result
	for i := 0; i < len(ts.Items); i++ {
		src := ts.Items[i]
		var item ResultItem
		item.Index = i
		item.DT = src.DT
		item.DTStr = time.Unix(int64(item.DT), 0).Format("2006-01-02 15:04:05")
		item.Value = float64(len(src.TXS))
		result.Items = append(result.Items, &item)
	}
	result.Count = len(result.Items)
	result.CurrentDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	c.mtx.Lock()
	c.analytics["an"] = &result
	c.mtx.Unlock()
}
