package blockchain

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func FilterTxByAddressAndFunSelector(target string, targetSelector string, tx *types.Transaction) *types.Transaction {
	to := tx.To().Hex()

	if to == target { // `to` is the contract we want

		// check is `mint` function or not
		txdata := hexutil.Encode(tx.Data())
		selector := txdata[:10]

		if targetSelector == selector {
			fmt.Printf("get:%s\n", tx.Hash().Hex())

			return tx
		}
	}

	return nil
}
