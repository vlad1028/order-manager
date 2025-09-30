package order

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
)

type BasicRepository interface {
	Get(context.Context, basetypes.ID) (*order.Order, error)
	Delete(context.Context, basetypes.ID) error
	AddOrUpdate(context.Context, *order.Order) (exists bool, err error)
	AddOrUpdateList(context.Context, []*order.Order) error
}

type RepositoryWithFilters interface {
	GetBy(context.Context, *order.Filter) ([]*order.Order, error)
	GetByPaginated(ctx context.Context, filter *order.Filter, offset uint, limit int) ([]*order.Order, error)
	DeleteBy(context.Context, *order.Filter) error
}

type Repository interface {
	BasicRepository
	RepositoryWithFilters
}
