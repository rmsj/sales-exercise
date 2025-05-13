// Package salebus provides business access to device sale domain.
package salebus

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/sdk/id"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
	"github.com/rmsj/service/foundation/otel"
)

// ErrNotFound is the error variables for CRUD operations.
var (
	ErrNotFound = errors.New("sale not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, sale Sale) error
	Delete(ctx context.Context, sale Sale) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Sale, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, saleID uuid.UUID) (Sale, error)
}

// Business manages the set of APIs for sale access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a sale domain API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	b := Business{
		log:    log,
		storer: storer,
	}

	return &b
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new sale to the system.
func (b *Business) Create(ctx context.Context, ns NewSale) (Sale, error) {
	ctx, span := otel.AddSpan(ctx, "business.salebus.create")
	defer span.End()

	// helper function to round the value of the discount
	roundToTwoDecimals := func(value float64) float64 {
		return math.Round(value*100) / 100
	}

	now := time.Now()

	slDB := Sale{
		ID:        id.New(),
		Discount:  ns.Discount,
		UpdatedAt: now,
		CreatedAt: now,
	}

	var saleAmount float64
	for _, item := range ns.Items {
		saleAmount += float64(item.Quantity) * item.Price
	}
	slDB.Amount = saleAmount

	if ns.Discount > slDB.Amount {
		return Sale{}, fmt.Errorf("discount[%f] is greater than total sale amount[%f]", ns.Discount, slDB.Amount)
	}

	// items
	var distributedDiscount float64
	for _, item := range ns.Items {
		itemAmount := float64(item.Quantity) * item.Price
		proportion := itemAmount / saleAmount
		itemDiscount := roundToTwoDecimals(proportion * ns.Discount)

		saleItem := SaleItem{
			SaleID:    slDB.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Discount:  itemDiscount,
			Amount:    itemAmount,
			UpdatedAt: time.Time{},
			CreatedAt: time.Time{},
		}
		distributedDiscount += itemDiscount
		slDB.Items = append(slDB.Items, saleItem)
	}

	if ns.Discount != distributedDiscount {
		slDB.Items[0].Discount += ns.Discount - distributedDiscount
	}

	if err := b.storer.Create(ctx, slDB); err != nil {
		return Sale{}, fmt.Errorf("create sale: %w", err)
	}

	return slDB, nil
}

// Delete removes the specified sale.
func (b *Business) Delete(ctx context.Context, sl Sale) error {
	ctx, span := otel.AddSpan(ctx, "business.salebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, sl); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing sales.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Sale, error) {
	ctx, span := otel.AddSpan(ctx, "business.salebus.query")
	defer span.End()

	sls, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return sls, nil
}

// Count returns the total number of sales.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.salebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the sale by the specified ID.
func (b *Business) QueryByID(ctx context.Context, slID uuid.UUID) (Sale, error) {
	ctx, span := otel.AddSpan(ctx, "business.salebus.querybyid")
	defer span.End()

	sl, err := b.storer.QueryByID(ctx, slID)
	if err != nil {
		return Sale{}, fmt.Errorf("query: slID[%d]: %w", slID, err)
	}

	return sl, nil
}
