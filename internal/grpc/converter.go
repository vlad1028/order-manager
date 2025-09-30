package grpc

import (
	"fmt"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	desc "github.com/vlad1028/order-manager/pkg/order-service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertIDsFromProto(ids []uint64) []basetypes.ID {
	res := make([]basetypes.ID, len(ids))

	for i, id := range ids {
		res[i] = basetypes.ID(id)
	}

	return res
}

func ConvertOrdersFromProto(orders []*desc.Order) ([]*order.Order, error) {
	res := make([]*order.Order, len(orders))

	for i, o := range orders {
		r, err := ConvertOrderFromProto(o)
		if err != nil {
			return res, err
		}
		res[i] = r
	}

	return res, nil
}

func ConvertOrdersToProto(orders []*order.Order) ([]*desc.Order, error) {
	res := make([]*desc.Order, len(orders))

	for i, o := range orders {
		r, err := ConvertOrderToProto(o)
		if err != nil {
			return res, err
		}
		res[i] = r
	}

	return res, nil
}

func ConvertOrderFromProto(o *desc.Order) (*order.Order, error) {
	res := &order.Order{}

	res.ID = basetypes.ID(o.Id)
	res.ClientID = basetypes.ID(o.ClientId)
	res.PickupPointID = basetypes.ID(o.PickupPointId)

	status, err := ConvertStatusFromProto(o.Status)
	if err != nil {
		return res, err
	}
	res.Status = status

	res.StatusUpdated = o.StatusUpdated.AsTime()
	res.Weight = uint(o.Weight)
	res.Cost = uint(o.Weight)

	return res, nil
}

func ConvertOrderToProto(o *order.Order) (*desc.Order, error) {
	res := &desc.Order{}

	res.Id = uint64(o.ID)
	res.ClientId = uint64(o.ClientID)
	res.PickupPointId = uint64(o.PickupPointID)

	status, err := ConvertStatusToProto(o.Status)
	if err != nil {
		return res, err
	}
	res.Status = status

	res.StatusUpdated = timestamppb.New(o.StatusUpdated)
	res.Weight = uint32(o.Weight)
	res.Cost = uint32(o.Weight)

	return res, nil
}

func ConvertStatusFromProto(s desc.OrderStatus) (order.Status, error) {
	switch s {
	case desc.OrderStatus_ORDER_STATUS_RETURNED:
		return order.Returned, nil
	case desc.OrderStatus_ORDER_STATUS_STORED:
		return order.Stored, nil
	case desc.OrderStatus_ORDER_STATUS_REACHED_CLIENT:
		return order.ReachedClient, nil
	case desc.OrderStatus_ORDER_STATUS_CANCELED:
		return order.Canceled, nil
	default:
		return "", fmt.Errorf("unknown order status: %v", s)
	}
}

func ConvertStatusToProto(s order.Status) (desc.OrderStatus, error) {
	switch s {
	case order.Canceled:
		return desc.OrderStatus_ORDER_STATUS_CANCELED, nil
	case order.ReachedClient:
		return desc.OrderStatus_ORDER_STATUS_REACHED_CLIENT, nil
	case order.Returned:
		return desc.OrderStatus_ORDER_STATUS_RETURNED, nil
	case order.Stored:
		return desc.OrderStatus_ORDER_STATUS_STORED, nil
	default:
		return desc.OrderStatus_ORDER_STATUS_UNSPECIFIED, fmt.Errorf("unknown order status: %v", s)
	}
}

func ConvertPackagingFromProto(packaging desc.OrderPackaging) (order.Packaging, error) {
	switch packaging {
	case desc.OrderPackaging_ORDER_PACKAGING_UNSPECIFIED:
		return nil, nil
	case desc.OrderPackaging_ORDER_PACKAGING_BOX:
		return order.NewBox(), nil
	case desc.OrderPackaging_ORDER_PACKAGING_BAG:
		return order.NewBag(), nil
	case desc.OrderPackaging_ORDER_PACKAGING_FILM:
		return order.NewFilm(), nil
	default:
		return nil, fmt.Errorf("unknown package")
	}
}
