package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("BLUEPRINT_DB_USERNAME"),
		os.Getenv("BLUEPRINT_DB_PASSWORD"),
		os.Getenv("BLUEPRINT_DB_HOST"),
		os.Getenv("BLUEPRINT_DB_PORT"),
		os.Getenv("BLUEPRINT_DB_DATABASE"),
	)

	// ✅ Config with pool tuning
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}

	// 🔥 Pool settings (IMPORTANT)
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// ✅ Timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	// ✅ Ping with timeout
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	DB = pool

	fmt.Println("✅ Connected to PostgreSQL")

	return nil
}