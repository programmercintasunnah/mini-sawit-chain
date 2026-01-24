package domain

import "errors"

var (
	ErrInvalidWeight   = errors.New("weight must be greater than zero")
	ErrHarvestNotFound = errors.New("harvest not found")
)
