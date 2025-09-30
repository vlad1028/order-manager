package order

import (
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"time"
)

type Order struct {
	ID            basetypes.ID `db:"id"`
	ClientID      basetypes.ID `db:"client_id"`
	PickupPointID basetypes.ID `db:"pickup_point_id"`
	Status        Status       `db:"status"`
	StatusUpdated time.Time    `db:"status_updated"` // only should db update this field
	Weight        uint         `db:"weight"`
	Cost          uint         `db:"cost"`
}

func NewOrder(ID, cID, ppID basetypes.ID, w, c uint) *Order {
	return &Order{
		ID:            ID,
		ClientID:      cID,
		PickupPointID: ppID,
		Status:        Stored,
		Weight:        w,
		Cost:          c,
	}
}

func (o *Order) SetStatus(newStatus Status) {
	o.Status = newStatus
}

func (o *Order) IsExpired(storeDuration time.Duration, now time.Time) bool {
	return o.Status != Stored || o.StatusUpdated.Add(storeDuration).Before(now)
}

func (o *Order) CanBeReturned(timeToMakeReturn time.Duration, now time.Time) bool {
	return o.Status == ReachedClient && o.StatusUpdated.Add(timeToMakeReturn).After(now)
}

func (o *Order) ApplyPackaging(p Packaging) error {
	if p != nil {
		return p.ApplyPackaging(o)
	}
	return nil
}
