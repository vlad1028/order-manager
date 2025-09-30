package order

import (
	"errors"
)

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderExists              = errors.New("order already exists")
	ErrOrderNotIssued           = errors.New("order is not issued")
	ErrWrongPickupPoint         = errors.New("order was issued from another Pick Up Point")
	ErrWrongClientID            = errors.New("wrong clientID")
	ErrReturnExpired            = errors.New("the deadline for making a return has expired")
	ErrCantCancel               = errors.New("order cannot be cancelled")
	ErrNoPrimaryPack            = errors.New("you need to provide primary packaging to use additional packaging")
	ErrAdditionalPackNotAllowed = errors.New("you can't add additional packaging to that primary packaging")
)
