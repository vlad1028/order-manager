package order

import (
	"github.com/vlad1028/order-manager/internal/models/basetypes"
)

type Filter struct {
	ID            *basetypes.ID `db:"id"`
	ClientID      *basetypes.ID `db:"client_id"`
	PickUpPointID *basetypes.ID `db:"pickup_point_id"`
	Status        *Status       `db:"status"`
}
