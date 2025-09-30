package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	models "github.com/vlad1028/order-manager/internal/models/order"
	"log"
)

func (s *Service) genCacheKey(orderID basetypes.ID) string {
	return "order:" + orderID.String()
}

func (s *Service) getOrder(ctx context.Context, orderID basetypes.ID) (*models.Order, error) {
	cachedOrder, found := s.cache.Get(ctx, s.genCacheKey(orderID))
	if found {
		return cachedOrder, nil
	}

	o, err := s.repo.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}

	s.setOrderCache(ctx, o)
	return o, nil
}

func (s *Service) addOrUpdate(ctx context.Context, o *models.Order) (exists bool, err error) {
	exists, err = s.repo.AddOrUpdate(ctx, o)
	if err != nil {
		return exists, err
	}

	s.setOrderCache(ctx, o)
	return exists, nil
}

func (s *Service) addOrUpdateList(ctx context.Context, orders []*models.Order) error {
	err := s.repo.AddOrUpdateList(ctx, orders)
	if err != nil {
		return err
	}

	for _, o := range orders {
		s.setOrderCache(ctx, o)
	}
	return nil
}

func (s *Service) setOrderCache(ctx context.Context, o *models.Order) {
	err := s.cache.Set(ctx, s.genCacheKey(o.ID), o)
	if err != nil {
		log.Printf("Failed to set order to cache: %v", err)
	}
}
