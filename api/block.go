package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func Block(c *gin.Context) {
	blockNumberStr := c.Param("id")
	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 64)
	if err != nil {
		c.AbortWithError(500, err)
	}

	b, err := db.Instance.GetBlock(blockNumber)
	if err != nil {
		type Result struct {
			Error string
		}
		var result Result
		result.Error = err.Error()
		c.IndentedJSON(http.StatusInternalServerError, result)
		return
	}
	c.IndentedJSON(http.StatusOK, b)
}
