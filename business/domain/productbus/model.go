package productbus

import (
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
)

// Product represents an individual product.
type Product struct {
	ID          uuid.UUID
	Name        name.Name
	Price       money.Money
	DateCreated time.Time
	DateUpdated time.Time
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name  name.Name
	Price money.Money
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name  *name.Name
	Price *money.Money
}
