package salebus

import (
	"time"

	"github.com/google/uuid"
)

// Sale represents an individual sale.
type Sale struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Discount  float64
	Amount    float64
	Items     []SaleItem
	UpdatedAt time.Time
	CreatedAt time.Time
}

type SaleItem struct {
	SaleID    uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Discount  float64
	Amount    float64
	UpdatedAt time.Time
	CreatedAt time.Time
}

// NewSale is what we require from clients when adding a sale.
type NewSale struct {
	Discount float64
	Items    []NewSaleItem
}

type NewSaleItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     float64
	Discount  float64
}
