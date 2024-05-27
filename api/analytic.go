package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/an"
)

func Analytic(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, an.Instance.GetResult("an"))
}
