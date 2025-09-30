package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/basetypes"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
	"time"
)

func (s *Service) AcceptReturn(ctx context.Context, req *orderServise.AcceptReturnRequest) (resp *orderServise.AcceptReturnResponse, err error) {
	resp = &orderServise.AcceptReturnResponse{}

	o, err := s.getOrder(ctx, req.OrderID)
	if err != nil {
		return resp, err
	}

	err = s.validateReturnOperation(o, req.ClientID, time.Now().UTC())
	if err != nil {
		return resp, err
	}

	o.SetStatus(order.Returned)
	_, err = s.addOrUpdate(ctx, o)
	if err != nil {
		return resp, err
	}

	event := order.Event{
		OrderID:   o.ID,
		Operation: "return",
		Timestamp: time.Now().UTC(),
	}
	s.sendEvent(event)

	return resp, nil
}

func (s *Service) validateReturnOperation(o *order.Order, clientID basetypes.ID, now time.Time) error {
	if o == nil {
		return orderServise.ErrOrderNotIssued
	}
	if o.ClientID != clientID {
		return orderServise.ErrWrongClientID
	}
	if o.PickupPointID != s.ID {
		return orderServise.ErrWrongPickupPoint
	}
	if !o.CanBeReturned(s.timeToMakeReturn, now) {
		return orderServise.ErrReturnExpired
	}
	return nil
}
