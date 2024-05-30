package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func State(c *gin.Context) {
	state := db.Instance.State()
	c.IndentedJSON(http.StatusOK, state)
}
