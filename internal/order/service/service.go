package service

import (
	"context"
	"time"

	"github.com/vlad1028/order-manager/internal/models/basetypes"
	models "github.com/vlad1028/order-manager/internal/models/order"
	"github.com/vlad1028/order-manager/internal/order"
)

// MessageSender defines the interface for sending messages to a message broker (like Kafka).
type MessageSender interface {
	SendMessage(key, value []byte) error
}

// CachedOrders defines the interface for a key-value cache for orders.
type CachedOrders interface {
	Get(ctx context.Context, key string) (*models.Order, bool)
	Set(ctx context.Context, key string, value *models.Order) error
}

var _ order.Service = (*Service)(nil)

// Service implements the business logic for managing orders.
// It orchestrates interactions between the database, cache, and message broker.
type Service struct {
	ID               basetypes.ID     // Unique identifier for the pickup point (ПВЗ).
	timeToStore      time.Duration    // Default duration to store an order.
	timeToMakeReturn time.Duration    // Time window within which a customer can return an order.
	repo             order.Repository // Repository for database operations.
	kafkaProducer    MessageSender    // Producer to send events to Kafka.
	cache            CachedOrders     // Cache for frequently accessed orders.
}

// NewOrderService creates and returns a new Service instance.
func NewOrderService(id basetypes.ID, timeToStore, timeToMakeReturn time.Duration, r order.Repository, kafkaProducer MessageSender, cache CachedOrders) *Service {
	return &Service{
		ID:               id,
		timeToStore:      timeToStore,
		timeToMakeReturn: timeToMakeReturn,
		repo:             r,
		kafkaProducer:    kafkaProducer,
		cache:            cache,
	}
}
