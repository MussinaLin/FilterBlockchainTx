package main

import (
	"context"
	"fmt"
	"io"
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
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	if err := setupLog(); err != nil {
		fmt.Println(err)
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
	rpcUrl := os.Getenv("ETH_RPC")
	log.Printf("rpcUrl:%s\n", rpcUrl)

	if err := blockchain.InitRpc(rpcUrl); err != nil {
		log.Fatalf("Failed to initialize rpc: %v\n", err)
	}
	defer blockchain.CloseRpc()

	// read env
	contractAddr := os.Getenv("TARGET_CONTRACT")
	log.Printf("contractAddr:%s\n", contractAddr)

	funcSelector := os.Getenv("FILTERED_FUNCTION_SELECTOR")
	log.Printf("funcSelector:%s\n", funcSelector)

	n := os.Getenv("BLOCK_START_FROM_LATEST")
	blockFromLast, _ := strconv.ParseUint(n, 10, 64)
	log.Printf("blockFromLatest:%d\n", blockFromLast)

	// get current block
	curBlock, err := blockchain.GetBlockByNumber(0)
	if err != nil {
		log.Fatalf("get current block fail. %v", err)
	}
	startBlock := curBlock.Number().Uint64() - blockFromLast

	var wg sync.WaitGroup
	mintTxs := make(chan *database.MintTx)
	start := time.Now()

	// handle mint tx goroutine
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
	log.Printf("Execution time: %s\n", elapsed)
	log.Println("===== Finish =====")

}

func setupLog() error {
	logFilePath := "./logs/txFilter.log"

	// make sure logs folder exist
	if err := os.MkdirAll("./logs", os.ModePerm); err != nil {
		return fmt.Errorf("error creating log directory: %v", err)
	}

	// set read/write permission
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer logFile.Close()

	// output log to file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// set log format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("===== transaction filter start =====")
	return nil
}

func scanBlockByToAddrAndFuncSelector(wg *sync.WaitGroup, mintTxs chan *database.MintTx, blockNum uint64, toAddr string, selector string) {
	defer wg.Done()
	log.Printf("process block:%d\n", blockNum)

	block, err := blockchain.GetBlockByNumber(blockNum)
	if err != nil {
		log.Printf("process block:%d error. %v\n", blockNum, err)
		return
	}

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
