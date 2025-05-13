package saledb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/domain/salebus"
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
	SaleID    uuid.UUID       `db:"sale_id"`
	ProductID uuid.UUID       `db:"product_id"`
	Quantity  int             `db:"quantity"`
	Discount  sql.NullFloat64 `db:"discount"`
	Amount    float64         `db:"amount"`
	UpdatedAt time.Time       `db:"updated_at"`
	CreatedAt time.Time       `db:"created_at"`
}

func toDBSale(bus salebus.Sale) dbSale {
	saleDB := dbSale{
		ID:        bus.ID,
		UserID:    bus.UserID,
		Discount:  sql.NullFloat64{Float64: bus.Discount, Valid: bus.Discount > 0},
		Amount:    bus.Amount,
		UpdatedAt: bus.UpdatedAt,
		CreatedAt: bus.CreatedAt,
	}

	return saleDB
}

//lint:ignore U1000 temp
func toBusSale(db dbSale, items []dbSaleItem) salebus.Sale {
	sl := salebus.Sale{
		ID:        db.ID,
		UserID:    db.UserID,
		Discount:  db.Discount.Float64,
		Amount:    db.Amount,
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
	sl.Items = toBusSaleItems(saleItems)

	return sl
}

//lint:ignore U1000 temp
func toBusSales(dbs []dbSale, dbItems []dbSaleItem) []salebus.Sale {
	bus := make([]salebus.Sale, len(dbs))

	for i, sl := range dbs {
		bus[i] = toBusSale(sl, dbItems)
	}

	return bus
}

func toDBSaleItem(bus salebus.SaleItem) dbSaleItem {
	saleItemDB := dbSaleItem{
		SaleID:    bus.SaleID,
		ProductID: bus.ProductID,
		Discount:  sql.NullFloat64{Float64: bus.Discount, Valid: bus.Discount > 0},
		Amount:    bus.Amount,
		UpdatedAt: bus.UpdatedAt,
		CreatedAt: bus.CreatedAt,
	}

	return saleItemDB
}

//lint:ignore U1000 temp
func toBusSaleItem(db dbSaleItem) salebus.SaleItem {
	slItem := salebus.SaleItem{
		SaleID:    db.SaleID,
		ProductID: db.ProductID,
		Discount:  db.Discount.Float64,
		Amount:    db.Amount,
		UpdatedAt: db.UpdatedAt,
		CreatedAt: db.CreatedAt,
	}

	return slItem
}

//lint:ignore U1000 temp
func toBusSaleItems(dbs []dbSaleItem) []salebus.SaleItem {
	bus := make([]salebus.SaleItem, len(dbs))

	for i, sli := range dbs {
		bus[i] = toBusSaleItem(sli)
	}

	return bus
}
