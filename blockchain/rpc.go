package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"

	// "math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var client *ethclient.Client

func InitRpc(rpcUrl string) error {
	var err error = nil
	client, err = ethclient.Dial(rpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	return nil
}

func GetBlockByNumber(num *big.Int) (*types.Block, error) {
	// get latest block num
	header, err := client.HeaderByNumber(context.Background(), num)
	if err != nil {
		return nil, fmt.Errorf("failed to get block (%d) header: %v", err, num)
	}

	// get block
	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get the latest block: %v", err)
	}

	return block, nil
}

func CloseRpc() {
	client.Close()
	log.Println("RPC connection closed")
}