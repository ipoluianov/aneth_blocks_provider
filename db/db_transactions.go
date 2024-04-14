package db

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipoluianov/aneth_blocks_provider/utils"
)

func (c *DB) LatestTransactions() string {
	c.mtx.Lock()
	type TransactionInfo struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Value    string `json:"value"`
		DataSize string `json:"data_size"`
		Color    string `json:"color"`
	}

	type Result struct {
		Items []TransactionInfo `json:"items"`
	}

	var s Result

	latestBlock := c.latestBlockNumber

	c.mtx.Unlock()
	for bn := latestBlock - 5; bn < latestBlock-3; bn++ {
		b, err := c.GetBlock(bn)
		if err != nil {
			continue
		}

		for _, tr := range b.Transactions {
			var ti TransactionInfo
			ti.From = utils.TrFrom(tr).Hex()
			ti.To = tr.To().Hex()
			ti.Value = utils.FormarValueToGWEI(tr.Value())
			ti.DataSize = fmt.Sprint(len(tr.Data()))
			ti.Color = colorForTransaction(tr)
			s.Items = append(s.Items, ti)
		}
	}

	bs, _ := json.MarshalIndent(s, "", " ")
	return string(bs)
}

func colorForTransaction(tr *types.Transaction) string {
	if len(tr.Data()) == 0 {
		return "#008000"
	}
	return "#888888"
}
