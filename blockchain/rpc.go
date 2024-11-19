package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RPCPool struct {
	clients []*ethclient.Client
	mu      sync.Mutex
	index   int
}

var rpcPool *RPCPool

func InitRpc(rpcUrl string) error {
	rpcUrls := strings.Split(rpcUrl, ",")
	clients := make([]*ethclient.Client, len(rpcUrls))
	log.Printf("init rpc pool with size:%d", len(rpcUrls))

	// init rpc clients with multiple rpc endpoint
	var err error
	for i := 0; i < len(rpcUrls); i++ {
		clients[i], err = ethclient.Dial(rpcUrls[i]) // assume all rpc init succ
		if err != nil {
			return fmt.Errorf("init RPC error:%v", err)
		}
	}

	rpcPool = &RPCPool{
		clients: clients,
		index:   0,
	}
	return nil
}

func (pool *RPCPool) getRPC() *ethclient.Client {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	client := pool.clients[pool.index]
	pool.index = (pool.index + 1) % len(pool.clients)
	return client
}

func (pool *RPCPool) closeRPC() {
	for _, c := range pool.clients {
		c.Close()
	}
}

func GetBlockByNumber(num uint64) (*types.Block, error) {
	blockNum := new(big.Int).SetUint64(num)
	if num == 0 {
		blockNum = nil
	}

	client := rpcPool.getRPC()
	header, err := client.HeaderByNumber(context.Background(), blockNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get block:%d header: %s", num, err)
	}

	// get block
	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block:%d %v", num, err)
	}

	return block, nil
}

func CloseRpc() {
	rpcPool.closeRPC()
	log.Println("RPC connection closed")
}
