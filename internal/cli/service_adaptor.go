package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
	"strconv"
)

var _ OrderCLIAdaptor = (*OrderServiceAdaptor)(nil)

type OrderServiceAdaptor struct {
	orderService orderServise.Service
}

func NewOrderServiceAdaptor(orderService orderServise.Service) *OrderServiceAdaptor {
	return &OrderServiceAdaptor{
		orderService: orderService,
	}
}

func (a *OrderServiceAdaptor) AcceptOrder(req *AcceptOrderRequest) error {
	var parseErr error = nil

	orderID, err := parseID(req.ID)
	parseErr = errors.Join(parseErr, err)

	clientID, err := parseID(req.ClientID)
	parseErr = errors.Join(parseErr, err)

	weight, err := parseUnsigned(req.Weight)
	parseErr = errors.Join(parseErr, err)

	cost, err := parseUnsigned(req.Cost)
	parseErr = errors.Join(parseErr, err)

	pack, err := parsePackaging(req.Packaging)
	parseErr = errors.Join(parseErr, err)

	if parseErr != nil {
		return parseErr
	}

	r := &orderServise.AcceptOrderRequest{
		ID:        orderID,
		ClientID:  clientID,
		Weight:    weight,
		Cost:      cost,
		Packaging: pack,
		AddFilm:   req.AddFilm,
	}

	_, err = a.orderService.AcceptOrder(context.Background(), r)

	return err
}

func (a *OrderServiceAdaptor) CancelOrder(req *CancelOrderRequest) error {
	orderID, err := parseID(req.ID)
	if err != nil {
		return err
	}

	r := &orderServise.CancelOrderRequest{
		ID: orderID,
	}

	_, err = a.orderService.CancelOrder(context.Background(), r)

	return err
}

func (a *OrderServiceAdaptor) IssueOrder(req *IssueOrderRequest) ([]*order.Order, error) {
	var orderIDs []basetypes.ID
	for _, idStr := range req.IDs {
		id, err := parseID(idStr)
		if err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, id)
	}

	r := &orderServise.IssueOrderRequest{
		IDs: orderIDs,
	}

	resp, err := a.orderService.IssueOrder(context.Background(), r)

	return resp.Orders, err
}

func (a *OrderServiceAdaptor) GetOrders(req *GetOrdersRequest) ([]*order.Order, error) {
	clientID, err := parseID(req.ClientID)
	if err != nil {
		return nil, err
	}

	r := &orderServise.GetOrdersRequest{
		ClientID:  clientID,
		LocalOnly: req.LocalOnly,
	}

	resp, err := a.orderService.GetOrders(context.Background(), r)

	return resp.Orders, err
}

func (a *OrderServiceAdaptor) AcceptReturn(req *AcceptReturnRequest) error {
	clientID, err := parseID(req.ClientID)
	if err != nil {
		return err
	}
	orderID, err := parseID(req.OrderID)
	if err != nil {
		return err
	}

	r := &orderServise.AcceptReturnRequest{
		ClientID: clientID,
		OrderID:  orderID,
	}

	_, err = a.orderService.AcceptReturn(context.Background(), r)

	return err
}

func (a *OrderServiceAdaptor) GetReturned(req *GetReturnedRequest) ([]*order.Order, error) {
	r := &orderServise.GetReturnedRequest{Page: req.Page, PerPage: req.PerPage}

	resp, err := a.orderService.GetReturned(context.Background(), r)
	return resp.Orders, err
}

func parseUnsigned(str string) (uint, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, errors.New("non-negative number expected")
	}
	return uint(i), nil
}

func parseID(idStr string) (basetypes.ID, error) {
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %w", err)
	}
	return basetypes.ID(idInt), nil
}
func parsePackaging(p string) (order.Packaging, error) {
	switch p {
	case "":
		return nil, nil
	case "box":
		return order.NewBox(), nil
	case "bag":
		return order.NewBag(), nil
	case "film":
		return order.NewFilm(), nil
	default:
		return nil, fmt.Errorf("unknown package")
	}
}
