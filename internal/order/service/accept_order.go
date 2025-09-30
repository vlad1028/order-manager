package service

import (
	"context"
	"github.com/vlad1028/order-manager/internal/models/order"
	orderServise "github.com/vlad1028/order-manager/internal/order"
	"time"
)

func (s *Service) AcceptOrder(ctx context.Context, req *orderServise.AcceptOrderRequest) (resp *orderServise.AcceptOrderResponse, err error) {
	resp = &orderServise.AcceptOrderResponse{}

	o := order.NewOrder(req.ID, req.ClientID, s.ID, req.Weight, req.Cost)

	pack, err := newPackaging(req.Packaging, req.AddFilm)
	if err != nil {
		return resp, err
	}

	if err = o.ApplyPackaging(pack); err != nil {
		return resp, err
	}

	exists, err := s.addOrUpdate(ctx, o)
	if err != nil {
		return resp, err
	}

	if exists {
		return resp, orderServise.ErrOrderExists
	}

	event := order.Event{
		OrderID:   o.ID,
		Operation: "accept",
		Timestamp: time.Now().UTC(),
	}
	s.sendEvent(event)

	return resp, nil
}

func newPackaging(p order.Packaging, addFilm bool) (pack order.Packaging, err error) {
	if addFilm {
		p, err = applyAdditionalPack(p, order.NewFilm())
	}
	return p, err
}

func applyAdditionalPack(p1 order.Packaging, p2 order.Packaging) (order.Packaging, error) {
	if p1 == nil {
		return nil, orderServise.ErrNoPrimaryPack
	}

	if w, ok := p1.(order.Wrapper); !ok {
		return nil, orderServise.ErrAdditionalPackNotAllowed
	} else {
		w.Wrap(p2)
	}

	return p1, nil
}
