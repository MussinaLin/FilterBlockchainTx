package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

	// read env
	contractAddr := os.Getenv("TARGET_CONTRACT")
	fmt.Printf("contractAddr:%s\n", contractAddr)

	funcSelector := os.Getenv("FILTERED_FUNCTION_SELECTOR")
	fmt.Printf("funcSelector:%s\n", funcSelector)

	n := os.Getenv("BLOCK_START_FROM_LATEST")
	blockFromLast, _ := strconv.ParseUint(n, 10, 64)
	fmt.Printf("blockFromLatest:%d\n", blockFromLast)

	// get current block
	curBlock, err := blockchain.GetBlockByNumber(nil)
	if err != nil {
		log.Fatalf("get current block fail. %v", err)
	}
	startBlock := curBlock.Number().Uint64() - blockFromLast

	var wg sync.WaitGroup

	start := time.Now()

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)

}

func scanBlockByToAddrAndFuncSelector(wg sync.WaitGroup, ctx context.Context, blockNum *big.Int, toAddr string, selector string) {
	block, _ := blockchain.GetBlockByNumber(blockNum)

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

}

func getTransactionSender(tx *types.Transaction) string {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Fatalf("Failed to get transaction sender: %v", err)
	}

	return from.Hex()
}
