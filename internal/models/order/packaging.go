package order

import (
	"fmt"
	"math"
)

type Packaging interface {
	ApplyPackaging(order *Order) error
}

type BasePackaging struct {
	weightLimit uint
	cost        uint
}

func (p *BasePackaging) validateWeight(weight uint) error {
	if weight > p.weightLimit {
		return fmt.Errorf("the order weight exceeds %d kg, choose another packaging", p.weightLimit)
	}
	return nil
}

func (p *BasePackaging) ApplyPackaging(order *Order) error {
	if err := p.validateWeight(order.Weight); err != nil {
		return err
	}
	order.Cost += p.cost
	return nil
}

type Wrapper interface {
	Packaging
	Wrap(p Packaging)
}

type PackageWrapper struct {
	BasePackaging
	wrapped Packaging
}

func (w *PackageWrapper) Wrap(p Packaging) {
	w.wrapped = p
}

func (w *PackageWrapper) ApplyPackaging(order *Order) error {
	if w.wrapped != nil {
		if err := w.wrapped.ApplyPackaging(order); err != nil {
			return err
		}
	}
	return w.BasePackaging.ApplyPackaging(order)
}

type Bag struct {
	PackageWrapper
}

func NewBag() *Bag {
	b := &Bag{}
	b.weightLimit = 10
	b.cost = 5
	return b
}

type Box struct {
	PackageWrapper
}

func NewBox() *Box {
	b := &Box{}
	b.weightLimit = 30
	b.cost = 20
	return b
}

type Film struct {
	BasePackaging
}

func NewFilm() *Film {
	f := &Film{}
	f.weightLimit = math.MaxInt
	f.cost = 1
	return f
}
