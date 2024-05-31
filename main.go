package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/an"
	"github.com/ipoluianov/aneth_blocks_provider/api"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func main() {
	router := gin.Default()
	router.GET("/state", api.State)
	router.GET("/analytic/:code", api.Analytic)
	router.GET("/blocks", api.Blocks)
	router.GET("/latest_block_number", api.LatestBlockNumber)
	router.GET("/block/:id", api.Block)
	go router.Run(":8201")

	db.Instance.Start()
	an.Instance.Start()
	fmt.Scanln()
}
