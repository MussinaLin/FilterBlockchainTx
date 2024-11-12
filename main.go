package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"

	"mussinalin/interview_bedrock/blockchain"
	"mussinalin/interview_bedrock/database"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}

	ctx := context.Background()

	// read database dsn
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	// init database
	if err := database.InitDB(dsn); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer database.CloseDB()

	// init blockchain rpc
	rpcurl := os.Getenv("ETH_RPC")
	log.Printf("rpcurl:%s\n", rpcurl)

	if err := blockchain.InitRpc(rpcurl); err != nil {
		log.Fatalf("Failed to initialize rpc: %v\n", err)
	}
	defer blockchain.CloseRpc()

	block, _ := blockchain.GetBlockByNumber(big.NewInt(21150126))

	contractAddr := os.Getenv("TARGET_CONTRACT")
	fmt.Printf("contractAddr:%s\n", contractAddr)

	funcSelector := os.Getenv("FILTERED_FUNCTION_SELECTOR")
	fmt.Printf("funcSelector:%s\n", funcSelector)

	var wg sync.WaitGroup

	start := time.Now()

	for _, tx := range block.Transactions() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := blockchain.FilterTxByAddressAndFunSelector(contractAddr, funcSelector, tx)
			if res != nil {
				// database
				mintTx := &database.MintTx{
					TxHash:    tx.Hash().Hex(),
					BlockNum:  block.Number().Uint64(),
					BlockHash: block.Hash().Hex(),
					Sender:    getTransactionSender(tx),
				}
				database.InsertTx(ctx, mintTx)
			}
		}()

	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)

}

// func scanBlock()

func getTransactionSender(tx *types.Transaction) string {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Fatalf("Failed to get transaction sender: %v", err)
	}

	return from.Hex()
}
