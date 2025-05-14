package saledb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/types/money"
)

type dbSale struct {
	ID        uuid.UUID       `db:"id"`
	UserID    uuid.UUID       `db:"user_id"`
	Discount  sql.NullFloat64 `db:"discount"`
	Amount    float64         `db:"amount"`
	UpdatedAt time.Time       `db:"updated_at"`
	CreatedAt time.Time       `db:"created_at"`
}

type dbSaleItem struct {
	SaleID     uuid.UUID       `db:"sale_id"`
	ProductID  uuid.UUID       `db:"product_id"`
	UnityPrice float64         `db:"unity_price"`
	Quantity   int             `db:"quantity"`
	Discount   sql.NullFloat64 `db:"discount"`
	Amount     float64         `db:"amount"`
	UpdatedAt  time.Time       `db:"updated_at"`
	CreatedAt  time.Time       `db:"created_at"`
}

func toDBSale(bus salebus.Sale) dbSale {

	saleDB := dbSale{
		ID:        bus.ID,
		UserID:    bus.UserID,
		Discount:  sql.NullFloat64{Float64: bus.Discount.Value(), Valid: bus.Discount.Value() > 0},
		Amount:    bus.Amount.Value(),
		UpdatedAt: bus.UpdatedAt,
		CreatedAt: bus.CreatedAt,
	}

	return saleDB
}

//lint:ignore U1000 temp
func toBusSale(db dbSale, items []dbSaleItem) (salebus.Sale, error) {

	discount, err := money.Parse(db.Discount.Float64)
	if err != nil {
		return salebus.Sale{}, fmt.Errorf("parse discount: %w", err)
	}

	amount, err := money.Parse(db.Amount)
	if err != nil {
		return salebus.Sale{}, fmt.Errorf("parse amount: %w", err)
	}

	sl := salebus.Sale{
		ID:        db.ID,
		UserID:    db.UserID,
		Discount:  discount,
		Amount:    amount,
		UpdatedAt: db.UpdatedAt,
		CreatedAt: db.CreatedAt,
	}

	// far from ideal - we can use a join instead
	var saleItems []dbSaleItem
	for _, item := range items {
		if item.SaleID == db.ID {
			saleItems = append(saleItems, item)
		}
	}

	sl.Items, err = toBusSaleItems(saleItems)
	if err != nil {
		return salebus.Sale{}, fmt.Errorf("parse items: %w", err)
	}

	return sl, nil
}

//lint:ignore U1000 temp
func toBusSales(dbs []dbSale, dbItems []dbSaleItem) ([]salebus.Sale, error) {
	bus := make([]salebus.Sale, len(dbs))

	for i, sl := range dbs {
		var err error
		bus[i], err = toBusSale(sl, dbItems)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}

func toDBSaleItem(bus salebus.SaleItem) dbSaleItem {
	saleItemDB := dbSaleItem{
		SaleID:     bus.SaleID,
		ProductID:  bus.ProductID,
		Quantity:   bus.Quantity,
		Discount:   sql.NullFloat64{Float64: bus.Discount.Value(), Valid: bus.Discount.Value() > 0},
		UnityPrice: bus.UnityPrice.Value(),
		Amount:     bus.Amount.Value(),
		UpdatedAt:  bus.UpdatedAt,
		CreatedAt:  bus.CreatedAt,
	}

	return saleItemDB
}

//lint:ignore U1000 temp
func toBusSaleItem(db dbSaleItem) (salebus.SaleItem, error) {

	discount, err := money.Parse(db.Discount.Float64)
	if err != nil {
		return salebus.SaleItem{}, fmt.Errorf("parse discount: %w", err)
	}

	amount, err := money.Parse(db.Amount)
	if err != nil {
		return salebus.SaleItem{}, fmt.Errorf("parse amount: %w", err)
	}

	unityPrice, err := money.Parse(db.UnityPrice)
	if err != nil {
		return salebus.SaleItem{}, fmt.Errorf("parse unity price: %w", err)
	}

	slItem := salebus.SaleItem{
		SaleID:     db.SaleID,
		ProductID:  db.ProductID,
		Quantity:   db.Quantity,
		Discount:   discount,
		UnityPrice: unityPrice,
		Amount:     amount,
		UpdatedAt:  db.UpdatedAt,
		CreatedAt:  db.CreatedAt,
	}

	return slItem, nil
}

//lint:ignore U1000 temp
func toBusSaleItems(dbs []dbSaleItem) ([]salebus.SaleItem, error) {
	bus := make([]salebus.SaleItem, len(dbs))

	for i, sli := range dbs {
		var err error
		bus[i], err = toBusSaleItem(sli)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
