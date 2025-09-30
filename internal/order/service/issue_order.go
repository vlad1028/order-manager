package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/vlad1028/order-manager/internal/metrics"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderService "github.com/vlad1028/order-manager/internal/order"
	"slices"
	"time"
)

func (s *Service) IssueOrder(ctx context.Context, req *orderService.IssueOrderRequest) (resp *orderService.IssueOrderResponse, err error) {
	resp = &orderService.IssueOrderResponse{}

	o, err := s.getOrder(ctx, req.IDs[0])
	if err != nil {
		return resp, err
	}

	orders, err := s.repo.GetBy(ctx, &order.Filter{ClientID: &o.ClientID})
	if err != nil {
		return resp, err
	}

	issuedOrders := filterByIds(orders, req.IDs)
	issuedOrders, err = filterStored(issuedOrders, s.ID)
	if err != nil {
		return resp, err
	}

	setIssueDate(issuedOrders)
	err = s.addOrUpdateList(ctx, issuedOrders)
	if err != nil {
		return resp, err
	}

	sendIssueEvents(s, issuedOrders)
	metrics.AddIssuedOrdersTotal(len(issuedOrders), "issued")

	resp.Orders = issuedOrders
	return resp, nil
}

func sendIssueEvents(s *Service, orders []*order.Order) {
	for _, o := range orders {
		event := order.Event{
			OrderID:   o.ID,
			Operation: "issue",
			Timestamp: time.Now().UTC(),
		}
		s.sendEvent(event)
	}
}

func filterByIds(orders []*order.Order, ids []basetypes.ID) []*order.Order {
	filtered := make([]*order.Order, 0, len(ids))
	for _, o := range orders {
		if slices.Contains(ids, o.ID) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

func setIssueDate(orders []*order.Order) {
	for _, o := range orders {
		o.SetStatus(order.ReachedClient)
	}
}

func filterStored(orders []*order.Order, ppid basetypes.ID) (filtered []*order.Order, err error) {
	err = nil

	filtered = make([]*order.Order, 0, len(orders))
	for _, o := range orders {
		if o.Status == order.Stored && o.PickupPointID == ppid {
			filtered = append(filtered, o)
		} else {
			err = errors.Join(err, fmt.Errorf("order[%v] is not stored", o.ID))
		}
	}
	return filtered, err
}
