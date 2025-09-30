package db

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vlad1028/order-manager/internal/order"
	"github.com/vlad1028/order-manager/internal/order/repository/postgres"
)

func SetupOrderRepository(pool *pgxpool.Pool) order.Repository {
	txManager := postgres.NewTxManager(pool)
	repos := postgres.NewPgRepository()
	storage := postgres.NewStorageFacade(txManager, repos)

	return storage
}
