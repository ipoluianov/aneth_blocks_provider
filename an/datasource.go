package an

import (
	"fmt"
	"sort"
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/gomisc/logger"
)

func (c *An) GetTransactions(beginDT uint64, endDT uint64, secondsPerBlock uint64) []*db.Tx {
	logger.Println("An::GetTransactions begin")
	unixTimeBegin := beginDT
	unixTimeEnd := endDT
	blocks := make([]*db.Block, 0)
	blNumber := db.Instance.LatestBlockNumber()
	blNumberBegin := blNumber - uint64((endDT-beginDT)/secondsPerBlock)

	logger.Println("An::GetTransactions expected blocks count:", blNumber-blNumberBegin)

	for blNumber > blNumberBegin {
		bl, err := db.Instance.GetBlockFromCache(blNumber)
		if err != nil || bl.Time < unixTimeBegin || bl.Time > unixTimeEnd {
			blNumber--
			continue
		}
		blocks = append(blocks, bl)
		blNumber--
	}

	result := make([]*db.Tx, 0, len(blocks)*300)

	logger.Println("An::GetTransactions blocks:", len(blocks))

	for _, bl := range blocks {
		result = append(result, bl.Txs...)
	}

	logger.Println("An::GetTransactions sorting")

	sort.Slice(result, func(i, j int) bool {
		return result[i].BlDT < result[j].BlDT
	})

	logger.Println("An::GetTransactions end")

	return result
}

func (c *An) GroupByMinutes(beginDT uint64, endDT uint64, txs []*db.Tx) *db.TxsByMinutes {
	logger.Println("An::GroupByMinutes begin")
	var res db.TxsByMinutes
	firstTxDt := beginDT
	lastTxDt := endDT

	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60

	lastTxDt = lastTxDt / 60
	lastTxDt = lastTxDt * 60

	countOfRanges := (lastTxDt - firstTxDt) / 60
	res.Items = make([]*db.TxsByMinute, countOfRanges)

	index := 0
	for dt := firstTxDt; dt < lastTxDt; dt += 60 {
		res.Items[index] = &db.TxsByMinute{}
		res.Items[index].DT = dt
		index++
	}

	for i := 0; i < len(txs); i++ {
		t := txs[i]
		rangeIndex := (t.BlDT - firstTxDt) / 60
		if int(rangeIndex) >= len(res.Items) {
			fmt.Println("OVERFLOW")
		}
		res.Items[rangeIndex].TXS = append(res.Items[rangeIndex].TXS, t)
	}

	logger.Println("An::GroupByMinutes end")

	return &res
}

func (c *An) GetLatestTransactions() *db.TxsByMinutes {
	lastSeconds := uint64(24 * 3600)
	lastTxDt := uint64(time.Now().UTC().Unix())
	firstTxDt := uint64(lastTxDt - lastSeconds)
	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60
	lastTxDt = lastTxDt / 60
	lastTxDt = (lastTxDt + 1) * 60
	txs := c.GetTransactions(firstTxDt, lastTxDt, 12)
	if len(txs) < 1 {
		return &db.TxsByMinutes{}
	}
	byMinutes := c.GroupByMinutes(firstTxDt, lastTxDt, txs)
	return byMinutes
}
