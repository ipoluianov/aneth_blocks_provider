package db

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Tx struct {
	BlNumber uint64          `json:"bl_n"`
	BlDT     uint64          `json:"bl_dt"`
	TxFrom   *common.Address `json:"tx_from"`
	TxTo     *common.Address `json:"tx_to"`
	TxValue  *big.Int        `json:"tx_value"`
	TxData   []byte          `json:"tx_data"`
}

type TxsByMinutes struct {
	Items []*TxsByMinute
}

type TxsByMinute struct {
	DT  uint64
	TXS []*Tx
}
