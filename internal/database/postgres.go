package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() error {
	// 🔥 1. Try production DB (Render)
	dsn := os.Getenv("DATABASE_URL")

	// 🔄 2. Fallback to local Docker (optional)
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("BLUEPRINT_DB_USERNAME"),
			os.Getenv("BLUEPRINT_DB_PASSWORD"),
			os.Getenv("BLUEPRINT_DB_HOST"),
			os.Getenv("BLUEPRINT_DB_PORT"),
			os.Getenv("BLUEPRINT_DB_DATABASE"),
		)
		fmt.Println("⚠️ Using LOCAL database")
	} else {
		// 🔥 Render requires SSL
		if !strings.Contains(dsn, "sslmode") {
			dsn += "?sslmode=require"
		}
		fmt.Println("🚀 Using RENDER database")
	}

	// ✅ Parse config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}

	// 🔥 Pool settings
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	DB = pool

	fmt.Println("✅ Connected to PostgreSQL")

	return nil
}

func CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password TEXT,
		fullname TEXT,
		provider TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);
	`

	_, err := DB.Exec(context.Background(), query)
	return err

}

func CreateReviewTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS reviews (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		rating INT CHECK (rating >= 1 AND rating <= 5),
		created_at TIMESTAMP DEFAULT NOW(),

		-- 🔗 Foreign key (IMPORTANT)
		CONSTRAINT fk_user
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
	);
	`

	_, err := DB.Exec(context.Background(), query)
	return err
}

