package cache

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/order"
)

type Mock struct {
}

func NewCacheMock() *Mock {
	return &Mock{}
}

func (c *Mock) Get(ctx context.Context, key string) (*order.Order, bool) {
	return nil, false
}

func (c *Mock) Set(ctx context.Context, key string, value *order.Order) error {
	return nil
}
