package vproductbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
	"github.com/rmsj/service/business/types/quantity"
)

// Product represents an individual product with extended information.
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        name.Name
	Cost        money.Money
	Quantity    quantity.Quantity
	DateCreated time.Time
	DateUpdated time.Time
	UserName    name.Name
}
