package domain

import "time"

type Shipment struct {
	ID        string
	HarvestID string
	MillID    string
	ShipDate  time.Time
	WeightKg  float64
	Status    ShipmentStatus
}

type ShipmentStatus string

const (
	ShipmentPending   ShipmentStatus = "PENDING"
	ShipmentDelivered ShipmentStatus = "DELIVERED"
)
