package an

import (
	"fmt"
	"sort"
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/aneth_blocks_provider/utils"
)

type TxsByMinutes struct {
	Items []*TxsByMinute
}

type TxsByMinute struct {
	DT  uint64
	TXS []*Tx
}

func (c *An) GetTransactions(beginDT uint64, endDT uint64, secondsPerBlock uint64) []*Tx {

	unixTimeBegin := beginDT
	unixTimeEnd := endDT
	blocks := make([]*db.Block, 0)
	blNumber := db.Instance.LatestBlockNumber()
	blNumberBegin := blNumber - int64((endDT-beginDT)/secondsPerBlock)*2

	fmt.Println("Count of blocks to request:", blNumber-blNumberBegin)

	for blNumber > blNumberBegin {
		bl, err := db.Instance.GetBlock(blNumber)
		if err != nil || bl.Header.Time < unixTimeBegin || bl.Header.Time > unixTimeEnd {
			blNumber--
			continue
		}
		blocks = append(blocks, bl)
		blNumber--
	}

	var result []*Tx

	for _, bl := range blocks {
		for txIndex := 0; txIndex < len(bl.Transactions); txIndex++ {
			t := bl.Transactions[txIndex]
			var item Tx
			item.BlockNumber = bl.Header.Number.Uint64()
			item.BlockDT = bl.Header.Time
			item.TxFrom = utils.TrFrom(t)
			item.TxTo = t.To()
			item.TxData = t.Data()
			result = append(result, &item)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].BlockDT < result[j].BlockDT
	})

	return result
}

func (c *An) GroupByMinutes(beginDT uint64, endDT uint64, txs []*Tx) *TxsByMinutes {
	var res TxsByMinutes
	firstTxDt := beginDT
	lastTxDt := endDT

	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60

	lastTxDt = lastTxDt / 60
	lastTxDt = lastTxDt * 60

	countOfRanges := (lastTxDt - firstTxDt) / 60
	res.Items = make([]*TxsByMinute, countOfRanges)

	index := 0
	for dt := firstTxDt; dt < lastTxDt; dt += 60 {
		res.Items[index] = &TxsByMinute{}
		res.Items[index].DT = dt
		index++
	}

	for i := 0; i < len(txs); i++ {
		t := txs[i]
		rangeIndex := (t.BlockDT - firstTxDt) / 60
		if int(rangeIndex) >= len(res.Items) {
			fmt.Println("OVERFLOW")
		}
		res.Items[rangeIndex].TXS = append(res.Items[rangeIndex].TXS, t)
	}

	return &res
}

func (c *An) GetLatestTransactions() *TxsByMinutes {
	lastSeconds := uint64(24 * 3600)
	lastTxDt := uint64(time.Now().UTC().Unix())
	firstTxDt := uint64(lastTxDt - lastSeconds)
	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60
	lastTxDt = lastTxDt / 60
	lastTxDt = (lastTxDt + 1) * 60
	txs := c.GetTransactions(firstTxDt, lastTxDt, 12)
	if len(txs) < 1 {
		return &TxsByMinutes{}
	}
	byMinutes := c.GroupByMinutes(firstTxDt, lastTxDt, txs)
	return byMinutes
}
