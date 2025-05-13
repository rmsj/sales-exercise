package productapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
)

// Product represents information about an individual product.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppProduct(prd productbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		Name:        prd.Name.String(),
		Price:       prd.Price.Value(),
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProducts(prds []productbus.Product) []Product {
	app := make([]Product, len(prds))
	for i, prd := range prds {
		app[i] = toAppProduct(prd)
	}

	return app
}

// =============================================================================

// NewProduct defines the data needed to add a new product.
type NewProduct struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required,gte=0"`
}

// Decode implements the decoder interface.
func (app *NewProduct) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusNewProduct(app NewProduct) (productbus.NewProduct, error) {

	name, err := name.Parse(app.Name)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse name: %w", err)
	}

	price, err := money.Parse(app.Price)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse price: %w", err)
	}

	bus := productbus.NewProduct{
		Name:  name,
		Price: price,
	}

	return bus, nil
}

// =============================================================================

// UpdateProduct defines the data needed to update a product.
type UpdateProduct struct {
	Name  *string  `json:"name"`
	Price *float64 `json:"price" validate:"omitempty,gte=0"`
}

// Decode implements the decoder interface.
func (app *UpdateProduct) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func toBusUpdateProduct(app UpdateProduct) (productbus.UpdateProduct, error) {
	var nme *name.Name
	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		nme = &nm
	}

	var price *money.Money
	if app.Price != nil {
		prc, err := money.Parse(*app.Price)
		if err != nil {
			return productbus.UpdateProduct{}, fmt.Errorf("parse: %w", err)
		}
		price = &prc
	}

	bus := productbus.UpdateProduct{
		Name:  nme,
		Price: price,
	}

	return bus, nil
}
