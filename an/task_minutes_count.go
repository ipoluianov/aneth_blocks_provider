package an

import (
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/gomisc/logger"
)

func (c *An) taskMinutesCount(result *Result, txsByMin *db.TxsByMinutes, txs []*db.Tx) {
	logger.Println("An::anTrCount begin")
	for i := 0; i < len(txsByMin.Items); i++ {
		src := txsByMin.Items[i]
		var item ResultItemByMinutes
		item.Index = i
		item.DT = src.DT
		item.DTStr = time.Unix(int64(item.DT), 0).UTC().Format("2006-01-02 15:04:05")
		item.Value = float64(len(src.TXS))
		result.ItemsByMinutes = append(result.ItemsByMinutes, &item)
	}
	result.Count = len(result.ItemsByMinutes)
	result.CurrentDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	logger.Println("An::anTrCount end")
}
