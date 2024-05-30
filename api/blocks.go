package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Blocks(c *gin.Context) {
	type ResultBlockInfo struct {
		BlockNumber      int64
		Hash             string
		Time             int
		TimeString       string
		TransactionCount int
	}
	type Result struct {
		Count  int
		Blocks []ResultBlockInfo
	}
	var result Result
	//latestBlockNumber := db.Instance.LatestBlockNumber()
	/*for i := latestBlockNumber - 1000; i < latestBlockNumber; i++ {
		b, err := db.Instance.GetBlock(i)
		if err == nil {
			var item ResultBlockInfo
			item.BlockNumber = b.Number
			item.Hash = b.Header.Hash().Hex()
			item.TransactionCount = len(b.Txs)
			item.Time = int(b.Header.Time)
			item.TimeString = time.Unix(int64(b.Header.Time), 0).String()
			result.Blocks = append(result.Blocks, item)
		}
	}*/
	result.Count = len(result.Blocks)

	c.IndentedJSON(http.StatusOK, result)
}
