package main

import (
	"context"
	"fmt"
	"log"

	// "math/big"
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

	rpcurl := os.Getenv("ETH_RPC")
	log.Printf("rpcurl:%s\n", rpcurl)

	// ctx := context.Background()
	if err := blockchain.InitRpc(rpcurl); err != nil {
		log.Fatalf("Failed to initialize rpc: %v\n", err)
	}
	defer blockchain.CloseRpc()

	block, _ := blockchain.GetBlockByNumber(nil)

	fmt.Printf("Block Hash: %s\n", block.Hash().Hex())
	fmt.Printf("Block Number: %d\n", block.Number().Uint64())
	fmt.Printf("Block Time: %d\n", block.Time())
	fmt.Printf("Block Nonce: %d\n", block.Nonce())
	fmt.Printf("Block Transactions Count: %d\n", len(block.Transactions()))

	// read each tx
	// for _, tx := range block.Transactions() {
	// 	fmt.Printf("Transaction Hash: %s\n", tx.Hash().Hex())
	// 	fmt.Printf("From: %s\n", getTransactionSender(client, tx))
	// 	fmt.Printf("To: %s\n", tx.To().Hex())
	// 	fmt.Printf("Value: %s\n", tx.Value().String())
	// 	fmt.Printf("Gas: %d\n", tx.Gas())
	// 	fmt.Printf("Gas Price: %s\n", tx.GasPrice().String())
	// 	fmt.Println("===================================")
	// }
}

func getTransactionSender(tx *types.Transaction) string {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Fatalf("Failed to get transaction sender: %v", err)
	}

	return from.Hex()
}

func databaseFn(ctx context.Context) {
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

	// test insert
	if err := database.InsertTx(ctx, "0x1234", 999, "0x2222", "0xmussinaeth"); err != nil {
		log.Fatalf("Error inserting user: %v\n", err)
	}

}
