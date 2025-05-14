package salebus

import (
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/types/money"
)

// Sale represents an individual sale.
type Sale struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Discount  money.Money
	Amount    money.Money
	Items     []SaleItem
	UpdatedAt time.Time
	CreatedAt time.Time
}

type SaleItem struct {
	SaleID     uuid.UUID
	ProductID  uuid.UUID
	UnityPrice money.Money
	Quantity   int
	Amount     money.Money
	Discount   money.Money
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

// NewSale is what we require from clients when adding a sale.
type NewSale struct {
	UserID   uuid.UUID
	Discount money.Money
	Items    []NewSaleItem
}

// NewSaleItem is what we require from clients when adding a sale item.
type NewSaleItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     money.Money
}

// SaleItemValue is a helper type to calculate the amount and proportional discount, if any
type SaleItemValue struct {
	Amount   money.Money
	Discount money.Money
}
