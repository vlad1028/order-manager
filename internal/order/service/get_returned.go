package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
)

func (s *Service) GetReturned(ctx context.Context, req *orderServise.GetReturnedRequest) (resp *orderServise.GetReturnedResponse, err error) {
	returned := order.Returned
	filter := &order.Filter{
		Status: &returned,
	}
	orders, err := s.repo.GetByPaginated(ctx, filter, uint(req.Page), req.PerPage)

	return &orderServise.GetReturnedResponse{Orders: orders}, err
}
