package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/vlad1028/order-manager/internal/grpc"
	"github.com/vlad1028/order-manager/internal/models/order"
	desc "github.com/vlad1028/order-manager/pkg/order-service/v1"
)

var _ OrderCLIAdaptor = (*OrderGrpcAdaptor)(nil)

type OrderGrpcAdaptor struct {
	orderService desc.OrderServiceClient
}

func NewOrderGrpcAdaptor(orderService desc.OrderServiceClient) *OrderGrpcAdaptor {
	return &OrderGrpcAdaptor{
		orderService: orderService,
	}
}

func (a *OrderGrpcAdaptor) AcceptOrder(req *AcceptOrderRequest) error {
	var parseErr error = nil

	orderID, err := a.parseID(req.ID)
	parseErr = errors.Join(parseErr, err)

	clientID, err := a.parseID(req.ClientID)
	parseErr = errors.Join(parseErr, err)

	weight, err := a.parseUnsigned(req.Weight)
	parseErr = errors.Join(parseErr, err)

	cost, err := a.parseUnsigned(req.Cost)
	parseErr = errors.Join(parseErr, err)

	pack, err := a.parsePackaging(req.Packaging)
	parseErr = errors.Join(parseErr, err)

	if parseErr != nil {
		return parseErr
	}

	r := &desc.AcceptOrderRequest{
		Id:        orderID,
		ClientId:  clientID,
		Weight:    weight,
		Cost:      cost,
		Packaging: pack,
		AddFilm:   req.AddFilm,
	}

	_, err = a.orderService.AcceptOrder(context.Background(), r)

	return err
}

func (a *OrderGrpcAdaptor) CancelOrder(req *CancelOrderRequest) error {
	orderID, err := a.parseID(req.ID)
	if err != nil {
		return err
	}

	r := &desc.CancelOrderRequest{
		Id: orderID,
	}

	_, err = a.orderService.CancelOrder(context.Background(), r)

	return err
}

func (a *OrderGrpcAdaptor) IssueOrder(req *IssueOrderRequest) ([]*order.Order, error) {
	var orderIDs []uint64
	for _, idStr := range req.IDs {
		id, err := a.parseID(idStr)
		if err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, id)
	}

	r := &desc.IssueOrderRequest{
		Ids: orderIDs,
	}

	resp, err := a.orderService.IssueOrder(context.Background(), r)
	if err != nil {
		return nil, err
	}

	return grpc.ConvertOrdersFromProto(resp.Orders)
}

func (a *OrderGrpcAdaptor) GetOrders(req *GetOrdersRequest) ([]*order.Order, error) {
	clientID, err := a.parseID(req.ClientID)
	if err != nil {
		return nil, err
	}

	r := &desc.GetOrdersRequest{
		ClientId:  clientID,
		LocalOnly: req.LocalOnly,
	}

	resp, err := a.orderService.GetOrders(context.Background(), r)
	if err != nil {
		return nil, err
	}

	return grpc.ConvertOrdersFromProto(resp.Orders)
}

func (a *OrderGrpcAdaptor) AcceptReturn(req *AcceptReturnRequest) error {
	clientID, err := a.parseID(req.ClientID)
	if err != nil {
		return err
	}
	orderID, err := a.parseID(req.OrderID)
	if err != nil {
		return err
	}

	r := &desc.AcceptReturnRequest{
		ClientId: clientID,
		OrderId:  orderID,
	}

	_, err = a.orderService.AcceptReturn(context.Background(), r)

	return err
}

func (a *OrderGrpcAdaptor) GetReturned(req *GetReturnedRequest) ([]*order.Order, error) {
	r := &desc.GetReturnedRequest{Page: uint32(req.Page), PerPage: uint32(req.PerPage)}

	resp, err := a.orderService.GetReturned(context.Background(), r)
	if err != nil {
		return nil, err
	}
	return grpc.ConvertOrdersFromProto(resp.Orders)
}

func (a *OrderGrpcAdaptor) parseUnsigned(str string) (uint32, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, errors.New("non-negative number expected")
	}
	return uint32(i), nil
}

func (a *OrderGrpcAdaptor) parseID(idStr string) (uint64, error) {
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %w", err)
	}
	return uint64(idInt), nil
}

func (a *OrderGrpcAdaptor) parsePackaging(p string) (*desc.OrderPackaging, error) {
	pack := desc.OrderPackaging_ORDER_PACKAGING_UNSPECIFIED

	switch p {
	case "":
		return nil, nil
	case "box":
		pack = desc.OrderPackaging_ORDER_PACKAGING_BOX
	case "bag":
		pack = desc.OrderPackaging_ORDER_PACKAGING_BAG
	case "film":
		pack = desc.OrderPackaging_ORDER_PACKAGING_FILM
	default:
		return &pack, fmt.Errorf("unknown package")
	}
	return &pack, nil
}
