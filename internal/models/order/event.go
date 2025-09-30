package order

import (
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"time"
)

type Event struct {
	OrderID   basetypes.ID `json:"order_id"`
	Operation string       `json:"operation"`
	Timestamp time.Time    `json:"timestamp"`
}
