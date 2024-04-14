package db

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ipoluianov/aneth_blocks_provider/utils"
)

func (c *DB) LatestBlocksTable() string {
	c.mtx.Lock()
	type BlockInfo struct {
		BlockNumber      int64  `json:"blockNumber"`
		Loaded           bool   `json:"loaded"`
		TransactionCount int    `json:"transactionCount"`
		Cast             string `json:"cast"`
		Time             string `json:"time"`
	}

	type Result struct {
		Items []BlockInfo `json:"items"`
	}

	var s Result

	latestBlock := c.latestBlockNumber

	c.mtx.Unlock()
	for bn := latestBlock - 25; bn < latestBlock; bn++ {
		b, err := c.GetBlock(bn)
		if err != nil {
			continue
		}
		var bi BlockInfo
		bi.Loaded = len(b.Header.Hash()) > 0
		bi.BlockNumber = bn
		bi.TransactionCount = b.Transactions.Len()
		bi.Time = fmt.Sprint(b.Header.Time)

		cast := big.NewInt(0)

		for _, tr := range b.Transactions {
			cast = cast.Add(cast, tr.Cost())
		}

		castFloat := new(big.Float)
		castFloat.SetInt(cast)

		ethValue := new(big.Float).Quo(castFloat, big.NewFloat(math.Pow10(9)))
		ethValueInt64, _ := ethValue.Int64()

		//bi.Cast = ethValue.String()
		bi.Cast = fmt.Sprint(ethValueInt64)

		bi.Cast = utils.FormatIntString(bi.Cast) + " GWEI"

		s.Items = append(s.Items, bi)
	}

	bs, _ := json.MarshalIndent(s, "", " ")
	return string(bs)
}
