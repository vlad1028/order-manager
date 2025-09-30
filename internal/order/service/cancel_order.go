package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
	"time"
)

func (s *Service) CancelOrder(ctx context.Context, req *orderServise.CancelOrderRequest) (resp *orderServise.CancelOrderResponse, err error) {
	resp = &orderServise.CancelOrderResponse{}

	o, err := s.getOrder(ctx, req.ID)
	if err != nil {
		return resp, err
	}

	if err = s.validateCancelOperation(o, time.Now().UTC()); err != nil {
		return resp, err
	}

	o.SetStatus(order.Canceled)
	_, err = s.addOrUpdate(ctx, o)

	return resp, err
}

func (s *Service) validateCancelOperation(o *order.Order, now time.Time) error {
	if o.Status != order.Returned && (!o.IsExpired(s.timeToStore, now) || o.Status == order.ReachedClient) {
		return orderServise.ErrCantCancel
	}
	return nil
}
