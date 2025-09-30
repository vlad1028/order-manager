package grpc

import (
	"context"
	"errors"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	orderServise "github.com/vlad1028/order-manager/internal/order"
	desc "github.com/vlad1028/order-manager/pkg/order-service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGrpcAdaptor struct {
	desc.UnimplementedOrderServiceServer
	service orderServise.Service
}

func NewOrderGrpcAdaptor(s orderServise.Service) *OrderGrpcAdaptor {
	return &OrderGrpcAdaptor{service: s}
}

func (s *OrderGrpcAdaptor) AcceptOrder(ctx context.Context, req *desc.AcceptOrderRequest) (*desc.AcceptOrderResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pack, err := ConvertPackagingFromProto(req.GetPackaging())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.AcceptOrderRequest{
		ID:        basetypes.ID(req.GetId()),
		ClientID:  basetypes.ID(req.GetClientId()),
		Weight:    uint(req.GetWeight()),
		Cost:      uint(req.GetCost()),
		Packaging: pack,
		AddFilm:   req.GetAddFilm(),
	}

	_, err = s.service.AcceptOrder(ctx, r)

	if err != nil {
		if errors.Is(err, orderServise.ErrOrderExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.AcceptOrderResponse{}, nil
}

func (s *OrderGrpcAdaptor) AcceptReturn(ctx context.Context, req *desc.AcceptReturnRequest) (*desc.AcceptReturnResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.AcceptReturnRequest{
		ClientID: basetypes.ID(req.GetClientId()),
		OrderID:  basetypes.ID(req.GetOrderId()),
	}

	_, err := s.service.AcceptReturn(ctx, r)

	if err != nil {
		if errors.Is(err, orderServise.ErrReturnExpired) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		} else if errors.Is(err, orderServise.ErrOrderNotIssued) || errors.Is(err, orderServise.ErrWrongClientID) || errors.Is(err, orderServise.ErrWrongPickupPoint) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.AcceptReturnResponse{}, nil
}

func (s *OrderGrpcAdaptor) CancelOrder(ctx context.Context, req *desc.CancelOrderRequest) (*desc.CancelOrderResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.CancelOrderRequest{
		ID: basetypes.ID(req.GetId()),
	}

	_, err := s.service.CancelOrder(ctx, r)

	if err != nil {
		if errors.Is(err, orderServise.ErrCantCancel) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &desc.CancelOrderResponse{}, nil
}

func (s *OrderGrpcAdaptor) GetOrders(ctx context.Context, req *desc.GetOrdersRequest) (*desc.GetOrdersResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.GetOrdersRequest{
		ClientID:  basetypes.ID(req.GetClientId()),
		LocalOnly: req.GetLocalOnly(),
	}

	resp, err := s.service.GetOrders(ctx, r)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orders, err := ConvertOrdersToProto(resp.Orders)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &desc.GetOrdersResponse{Orders: orders}, nil
}

func (s *OrderGrpcAdaptor) GetReturned(ctx context.Context, req *desc.GetReturnedRequest) (*desc.GetReturnedResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.GetReturnedRequest{Page: int(req.GetPage()), PerPage: int(req.GetPerPage())}

	resp, err := s.service.GetReturned(ctx, r)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orders, err := ConvertOrdersToProto(resp.Orders)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &desc.GetReturnedResponse{Orders: orders}, nil
}

func (s *OrderGrpcAdaptor) IssueOrder(ctx context.Context, req *desc.IssueOrderRequest) (*desc.IssueOrderResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r := &orderServise.IssueOrderRequest{
		IDs: ConvertIDsFromProto(req.GetIds()),
	}

	resp, err := s.service.IssueOrder(ctx, r)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orders, err := ConvertOrdersToProto(resp.Orders)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &desc.IssueOrderResponse{Orders: orders}, nil
}
