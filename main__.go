package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ipoluianov/aneth_blocks_provider/api"
	"github.com/ipoluianov/aneth_blocks_provider/db"
)

func main1() {
	//db := db.NewDB("POLYGON", "https://polygon-rpc.com/", 2000)
	//db.Start()
	db.CreateDB()
	return
	db.Instance.Start()

	router := gin.Default()
	router.GET("/state", api.State)
	router.GET("/blocks", api.Blocks)
	router.GET("/latest_block_number", api.LatestBlockNumber)
	router.GET("/block/:id", api.Block)
	router.Run(":8201")
}
