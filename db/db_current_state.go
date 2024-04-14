package db

import "encoding/json"

func (c *DB) CurrentState() string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	type State struct {
		LatestBlock int64    `json:"latestBlock"`
		Network     string   `json:"network"`
		Log         []string `json:"log"`
	}
	var s State
	s.LatestBlock = c.latestBlockNumber
	s.Network = c.network
	bs, _ := json.MarshalIndent(s, "", " ")
	return string(bs)
}
