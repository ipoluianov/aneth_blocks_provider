package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TrFrom(tx *types.Transaction) common.Address {
	from, _ := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
	return from
}
