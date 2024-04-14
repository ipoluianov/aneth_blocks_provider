package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func State(c *gin.Context) {
	type Result struct {
		MinBlock      int64
		MaxBlock      int64
		CountOfBlocks int
		Network       string
	}
	var result Result
	min, max, count, network := db.Instance.State()
	result.MinBlock = min
	result.MaxBlock = max
	result.CountOfBlocks = count
	result.Network = network
	c.IndentedJSON(http.StatusOK, result)
}
