package order

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
)

type Service interface {
	AcceptOrder(context.Context, *AcceptOrderRequest) (*AcceptOrderResponse, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error)
	IssueOrder(context.Context, *IssueOrderRequest) (*IssueOrderResponse, error)
	GetOrders(context.Context, *GetOrdersRequest) (*GetOrdersResponse, error)
	AcceptReturn(context.Context, *AcceptReturnRequest) (*AcceptReturnResponse, error)
	GetReturned(context.Context, *GetReturnedRequest) (*GetReturnedResponse, error)
}

type (
	AcceptOrderRequest struct {
		ID        basetypes.ID
		ClientID  basetypes.ID
		Weight    uint
		Cost      uint
		Packaging order.Packaging
		AddFilm   bool
	}
	AcceptOrderResponse struct {
	}

	AcceptReturnRequest struct {
		ClientID basetypes.ID
		OrderID  basetypes.ID
	}
	AcceptReturnResponse struct {
	}

	CancelOrderRequest struct {
		ID basetypes.ID
	}
	CancelOrderResponse struct {
	}

	GetOrdersRequest struct {
		ClientID  basetypes.ID
		LocalOnly bool
	}
	GetOrdersResponse struct {
		Orders []*order.Order
	}

	GetReturnedRequest struct {
		Page    int
		PerPage int
	}
	GetReturnedResponse struct {
		Orders []*order.Order
	}

	IssueOrderRequest struct {
		IDs []basetypes.ID
	}
	IssueOrderResponse struct {
		Orders []*order.Order
	}
)
