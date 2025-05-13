package productdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
)

type product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Price       float64   `db:"price"`
	DateCreated time.Time `db:"created_at"`
	DateUpdated time.Time `db:"updated_at"`
}

func toDBProduct(bus productbus.Product) product {
	db := product{
		ID:          bus.ID,
		Name:        bus.Name.String(),
		Price:       bus.Price.Value(),
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}

	return db
}

func toBusProduct(db product) (productbus.Product, error) {
	name, err := name.Parse(db.Name)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse name: %w", err)
	}

	price, err := money.Parse(db.Price)
	if err != nil {
		return productbus.Product{}, fmt.Errorf("parse cost: %w", err)
	}

	bus := productbus.Product{
		ID:          db.ID,
		Name:        name,
		Price:       price,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusProducts(dbs []product) ([]productbus.Product, error) {
	bus := make([]productbus.Product, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusProduct(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
