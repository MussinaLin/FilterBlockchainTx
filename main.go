package main

import (
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

	// log
	if err := initLog(); err != nil {
		log.Fatalln(err)
	}

	// init database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer database.CloseDB()

	// init blockchain rpc
	if err := initRpc(); err != nil {
		log.Fatalf("Failed to initialize rpc: %v\n", err)
	}
	defer blockchain.CloseRpc()

	// read env
	n := os.Getenv("BLOCK_START_FROM_LATEST")
	blockFromLast, _ := strconv.ParseUint(n, 10, 64)
	log.Printf("blockFromLatest:%d\n", blockFromLast)

	// get processing block range
	curBlock, err := blockchain.GetBlockByNumber(0)
	if err != nil {
		log.Fatalf("get current block fail. %v", err)
	}
	startBlock := curBlock.Number().Uint64() - blockFromLast

	var wg sync.WaitGroup
	mintTxs := make(chan *database.MintTx) // each mint tx will push into mintTxs channel
	start := time.Now()

	// handle mint tx
	go handleMintTx(mintTxs)

	wg.Add(1) // for processBlocks()
	go processBlocks(&wg, mintTxs, startBlock, curBlock.Number().Uint64())

	wg.Wait()
	close(mintTxs)

	elapsed := time.Since(start)
	log.Printf("Execution time: %s\n", elapsed)
	log.Println("===== Finish =====")

}

func initLog() error {
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

	// output log to file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// set log format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("===== transaction filter start =====")
	return nil
}

func initDB() error {
	// read database dsn
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	// init database
	return database.InitDB(dsn)
}

func initRpc() error {
	rpcUrl := os.Getenv("ETH_RPC")
	log.Printf("rpcUrl:%s\n", rpcUrl)
	return blockchain.InitRpc(rpcUrl)
}

func processBlocks(wg *sync.WaitGroup, mintTxs chan *database.MintTx, start uint64, end uint64) {
	defer wg.Done()

	// read env
	contractAddr := os.Getenv("TARGET_CONTRACT")
	log.Printf("contractAddr:%s\n", contractAddr)

	funcSelector := os.Getenv("FILTERED_FUNCTION_SELECTOR")
	log.Printf("funcSelector:%s\n", funcSelector)

	for i := start; i <= end; i++ {
		wg.Add(1)
		go scanBlockByToAddrAndFuncSelector(wg, mintTxs, i, contractAddr, funcSelector)

		// add random delay to avoid free rpc rate limit
		// at least delay 100ms
		delay := time.Duration(rng.Intn(10)+30) * time.Millisecond
		time.Sleep(delay)
	}
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
			// database obj
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

func handleMintTx(mintTxs chan *database.MintTx) {
	for tx := range mintTxs {
		fmt.Println("receive tx from channel:", tx.TxHash)
		err := database.InsertTx(tx)
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
