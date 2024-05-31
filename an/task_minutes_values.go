package an

import (
	"math/big"
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/gomisc/logger"
)

func (c *An) taskMinutesValues(result *Result, txsByMin *db.TxsByMinutes, txs []*db.Tx) {
	logger.Println("An::anTrValue begin")
	for i := 0; i < len(txsByMin.Items); i++ {
		src := txsByMin.Items[i]
		var item ResultItemByMinutes
		item.Index = i
		item.DT = src.DT
		item.DTStr = time.Unix(int64(item.DT), 0).UTC().Format("2006-01-02 15:04:05")

		v := big.NewInt(0)
		for _, t := range src.TXS {
			v = v.Add(v, t.TxValue)
		}
		item.Value, _ = v.Float64()

		result.ItemsByMinutes = append(result.ItemsByMinutes, &item)
	}
	result.Count = len(result.ItemsByMinutes)
	result.CurrentDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	logger.Println("An::anTrValue end")
}
