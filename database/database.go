package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MintTx struct {
	TxHash    string
	BlockNum  int
	BlockHash string
	Sender    string
}

// global PGX Pool
var db *pgxpool.Pool

func InitDB(dsn string) error {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	// connection config
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.HealthCheckPeriod = 5 * time.Minute
	cfg.MaxConnIdleTime = 30 * time.Minute

	// create connection pool
	db, err = pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	log.Println("Database connection pool initialized")
	return nil
}

func CloseDB() {
	db.Close()
	log.Println("Database connection pool closed")
}

func InsertTx(ctx context.Context, txHash string, blockNum int, blockHash string, sender string) error {
	query := `INSERT INTO mint_tx (tx_hash, block_num, block_hash, sender) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(ctx, query, txHash, blockNum, blockHash, sender)
	if err != nil {
		return fmt.Errorf("failed to insert tx: %w", err)
	}
	log.Println("Inserted tx:", txHash)
	return nil
}
