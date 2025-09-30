package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
)

func (s *Service) GetOrders(ctx context.Context, req *orderServise.GetOrdersRequest) (resp *orderServise.GetOrdersResponse, err error) {
	stored := order.Stored
	filter := &order.Filter{
		ClientID: &req.ClientID,
		Status:   &stored,
	}

	if req.LocalOnly {
		filter.PickUpPointID = &s.ID
	}

	orders, err := s.repo.GetBy(ctx, filter)

	return &orderServise.GetOrdersResponse{Orders: orders}, err
}
