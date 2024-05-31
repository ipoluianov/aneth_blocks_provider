package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/an"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

type MainState struct {
	DbState *db.DbState
	AnState *an.AnState
}

func State(c *gin.Context) {
	var mainState MainState
	mainState.DbState = db.Instance.GetState()
	mainState.AnState = an.Instance.GetState()
	c.IndentedJSON(http.StatusOK, mainState)
}
