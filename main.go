package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"

	"mussinalin/interview_bedrock/blockchain"
	"mussinalin/interview_bedrock/database"
)

var rng *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}

	// 10s timeout context for database operation
	//ctx , cancel:= context.WithTimeout(context.Background(), 10 * time.Second)
	ctx := context.Background()
	// defer cancel()

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
	curBlock, err := blockchain.GetBlockByNumber(0)
	if err != nil {
		log.Fatalf("get current block fail. %v", err)
	}
	startBlock := curBlock.Number().Uint64() - blockFromLast

	var wg sync.WaitGroup

	start := time.Now()
	mintTxs := make(chan *database.MintTx)

	go handleMintTx(ctx, mintTxs)

	for i := startBlock; i <= curBlock.Number().Uint64(); i++ {
		wg.Add(1)
		go scanBlockByToAddrAndFuncSelector(&wg, mintTxs, i, contractAddr, funcSelector)

		// add random delay to avoid free rpc rate limit
		// at least delay 100ms
		delay := time.Duration(rng.Intn(50)+100) * time.Millisecond
		time.Sleep(delay)
	}

	wg.Wait()
	close(mintTxs)

	elapsed := time.Since(start)
	fmt.Printf("Execution time: %s\n", elapsed)

}

func scanBlockByToAddrAndFuncSelector(wg *sync.WaitGroup, mintTxs chan *database.MintTx, blockNum uint64, toAddr string, selector string) {
	defer wg.Done()
	fmt.Printf("process block:%d", blockNum)

	block, err := blockchain.GetBlockByNumber(blockNum)
	if err != nil {
		fmt.Printf("process block:%d error. %v\n", blockNum, err)
		return
	}
	if block.Transactions() == nil {
		fmt.Printf("process block:%d  Transactions is nil.\n", blockNum)
		return
	}

	fmt.Printf("process block:%d  tx len:%d\n", blockNum, block.Transactions().Len())
	for _, tx := range block.Transactions() {
		res := blockchain.FilterTxByAddressAndFunSelector(toAddr, selector, tx)
		if res != nil {
			// database
			mintTx := &database.MintTx{
				TxHash:    tx.Hash().Hex(),
				BlockNum:  block.Number().Uint64(),
				BlockHash: block.Hash().Hex(),
				Sender:    getTransactionSender(tx),
			}
			mintTxs <- mintTx
		}
	}

}

func handleMintTx(ctx context.Context, mintTxs chan *database.MintTx) {
	for tx := range mintTxs {
		fmt.Println("receive tx from channel:", tx.TxHash)
		err := database.InsertTx(ctx, tx)
		if err != nil {
			fmt.Println("insert tx fail:", tx.TxHash)
		}
	}

}

func getTransactionSender(tx *types.Transaction) string {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Fatalf("Failed to get transaction sender: %v", err)
	}

	return from.Hex()
}
