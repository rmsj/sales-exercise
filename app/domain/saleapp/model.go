package saleapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/types/money"
)

type Customer struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type Item struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	UnityPrice float64   `json:"unity_price"`
	Quantity   int       `json:"quantity"`
	Amount     float64   `json:"amount"`
	Discount   float64   `json:"discount"`
}

// Sale represents information about an individual sale.
type Sale struct {
	ID        uuid.UUID `json:"id"`
	Customer  Customer  `json:"customer"`
	Items     []Item    `json:"items"`
	UpdatedAt string    `json:"updatedAt"`
	CreatedAt string    `json:"createdAt"`
}

// Encode implements the encoder interface.
func (app Sale) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppSale(bus salebus.Sale, user userbus.User, productsInSale []productbus.Product) (Sale, error) {
	saleApp := Sale{
		ID: bus.ID,
		Customer: Customer{
			ID:    bus.UserID,
			Name:  user.Name.String(),
			Email: user.Email.Address,
		},
		UpdatedAt: bus.UpdatedAt.Format(time.RFC3339),
		CreatedAt: bus.CreatedAt.Format(time.RFC3339),
	}

	for _, item := range bus.Items {
		var product productbus.Product
		for _, prd := range productsInSale {
			if prd.ID.String() == item.ProductID.String() {
				product = prd
			}
		}
		saleApp.Items = append(saleApp.Items, Item{
			ID:         item.ProductID,
			Name:       product.Name.String(),
			UnityPrice: item.UnityPrice.Value(),
			Quantity:   item.Quantity,
			Amount:     item.Amount.Value(),
			Discount:   item.Discount.Value(),
		})
	}

	return saleApp, nil
}

// NewSale defines the data needed to add a new sale.
type NewSale struct {
	Discount float64       `json:"discount" validate:"omitempty,gte=0,lte=1000000"`
	Items    []NewSaleItem `json:"items" validate:"required"`
}

type NewSaleItem struct {
	ProductID string `json:"productId" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gte=0,lte=100"`
}

func toBusNewSale(userID uuid.UUID, app NewSale, productsInSale []productbus.Product) (salebus.NewSale, error) {

	discount, err := money.Parse(app.Discount)
	if err != nil {
		return salebus.NewSale{}, fmt.Errorf("parse discount: %w", err)
	}

	bus := salebus.NewSale{
		UserID:   userID,
		Discount: discount,
	}

	// far from ideal - we can use a join instead
	var saleItems []salebus.NewSaleItem
	for _, item := range app.Items {
		var product productbus.Product
		for _, prd := range productsInSale {
			if prd.ID.String() == item.ProductID {
				product = prd
			}
		}
		newItem, err := toBusNewSaleItem(item, product.Price)
		if err != nil {
			return salebus.NewSale{}, fmt.Errorf("parse items: %w", err)
		}
		saleItems = append(saleItems, newItem)
	}
	bus.Items = saleItems

	return bus, nil
}

func toBusNewSaleItem(app NewSaleItem, itemPrice money.Money) (salebus.NewSaleItem, error) {

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return salebus.NewSaleItem{}, fmt.Errorf("parse product id: %w", err)
	}

	slItem := salebus.NewSaleItem{
		ProductID: productID,
		Quantity:  app.Quantity,
		Price:     itemPrice,
	}

	return slItem, nil
}

// Decode implements the decoder interface.
func (app *NewSale) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewSale) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
