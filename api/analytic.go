package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/an"
)

func Analytic(c *gin.Context) {
	code := c.Param("code")
	c.IndentedJSON(http.StatusOK, an.Instance.GetResult(code))
}
