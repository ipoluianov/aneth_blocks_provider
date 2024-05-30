package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func LatestBlockNumber(c *gin.Context) {
	type Result struct {
		LatestBlockNumber uint64
	}
	var result Result
	result.LatestBlockNumber = db.Instance.LatestBlockNumber()
	c.IndentedJSON(http.StatusOK, result)
}
