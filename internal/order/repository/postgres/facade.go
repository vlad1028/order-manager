package postgres

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderRepo "github.com/vlad1028/order-manager/internal/order"
)

var _ orderRepo.Repository = (*storageFacade)(nil)

type storageFacade struct {
	txManager    *TxManager
	pgRepository *PgRepository
}

func NewStorageFacade(txManager *TxManager, pgRepository *PgRepository) *storageFacade {
	return &storageFacade{
		txManager:    txManager,
		pgRepository: pgRepository,
	}
}

func (s *storageFacade) Get(ctx context.Context, id basetypes.ID) (o *order.Order, err error) {
	err = s.txManager.Run(ctx, func(tx pgx.Tx) error {
		o, err = s.pgRepository.Get(ctx, tx, id)
		return err
	})
	return
}

func (s *storageFacade) Delete(ctx context.Context, id basetypes.ID) error {
	return s.txManager.Run(ctx, func(tx pgx.Tx) error {
		return s.pgRepository.Delete(ctx, tx, id)
	})
}

func (s *storageFacade) AddOrUpdate(ctx context.Context, o *order.Order) (exists bool, err error) {
	err = s.txManager.Run(ctx, func(tx pgx.Tx) error {
		exists, err = s.pgRepository.AddOrUpdate(ctx, tx, o)
		return err
	})
	return
}

func (s *storageFacade) AddOrUpdateList(ctx context.Context, orders []*order.Order) error {
	return s.txManager.RunRepeatableRead(ctx, func(tx pgx.Tx) error {
		for _, o := range orders {
			if _, err := s.pgRepository.AddOrUpdate(ctx, tx, o); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *storageFacade) GetBy(ctx context.Context, filter *order.Filter) (orders []*order.Order, err error) {
	return s.GetByPaginated(ctx, filter, 0, -1)
}

func (s *storageFacade) GetByPaginated(ctx context.Context, filter *order.Filter, offset uint, limit int) (orders []*order.Order, err error) {
	err = s.txManager.Run(ctx, func(tx pgx.Tx) error {
		orders, err = s.pgRepository.GetByPaginated(ctx, tx, filter, offset, limit)
		return err
	})
	return
}

func (s *storageFacade) DeleteBy(ctx context.Context, filter *order.Filter) error {
	return s.txManager.Run(ctx, func(tx pgx.Tx) error {
		return s.pgRepository.DeleteBy(ctx, tx, filter)
	})
}
