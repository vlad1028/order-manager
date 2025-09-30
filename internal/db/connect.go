package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vlad1028/order-manager/configs"
	"log"
)

func ConnectToDB(env string) (*pgxpool.Pool, error) {
	var dsn string

	switch env {
	case "prod":
		dsn = configs.MAIN_DB_DNS
	case "test":
		dsn = configs.TEST_DB_DNS
	default:
		log.Fatalf("Unknown environment: %s", env)
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return pool, nil
}
