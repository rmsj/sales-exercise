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
	SaleID    uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Discount  money.Money
	Amount    money.Money
	UpdatedAt time.Time
	CreatedAt time.Time
}

// NewSale is what we require from clients when adding a sale.
type NewSale struct {
	UserID   uuid.UUID
	Discount money.Money
	Items    []NewSaleItem
}

type NewSaleItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     money.Money
}
