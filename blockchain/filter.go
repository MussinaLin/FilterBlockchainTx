package blockchain

import (
	"github.com/ethereum/go-ethereum/core/types"
)

func FilterTxByAddress(target string, tx *types.Transaction) string {
	to := tx.To().Hex()

	if to == target {
		return tx.Hash().Hex()
	}

	return ""
}
