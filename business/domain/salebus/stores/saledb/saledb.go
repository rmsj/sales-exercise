// Package saledb contains events related CRUD functionality.
package saledb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
)

// Store manages the set of APIs for sale database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (salebus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create adds a Sale to the sqldb. It returns the created Sale with
// fields like ID and CreatedAt populated.
func (s *Store) Create(ctx context.Context, sale salebus.Sale) error {
	const q = `
	INSERT INTO sales
		(id, user_id, discount, amount, updated_at, created_at)
	VALUES
		(:id, :user_id, :discount, :amount, :updated_at, :created_at)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBSale(sale)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	// insert items
	for _, item := range sale.Items {
		const qi = `
		INSERT INTO sale_items
			(sale_id, product_id, quantity, unity_price, discount, amount, created_at, updated_at)
		VALUES
			(:sale_id, :product_id, :quantity, :unity_price, :discount, :amount, :created_at, :updated_at)`

		if err := sqldb.NamedExecContext(ctx, s.log, s.db, qi, toDBSaleItem(item)); err != nil {
			return fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	return nil
}

// Delete removes the sale identified by a given ID.
func (s *Store) Delete(ctx context.Context, sl salebus.Sale) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: sl.ID.String(),
	}

	const q = `DELETE FROM sales WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query gets all sales from the database.
func (s *Store) Query(ctx context.Context, filter salebus.QueryFilter, orderBy order.By, page page.Page) ([]salebus.Sale, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `SELECT * FROM sales`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbSales []dbSale
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSales); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	// get the ids for the sales to get the items
	var salesIDs []uuid.UUID
	for _, dbsl := range dbSales {
		salesIDs = append(salesIDs, dbsl.ID)
	}
	items, err := s.getSaleItems(ctx, salesIDs)
	if err != nil {
		return nil, fmt.Errorf("getSaleItems: %w", err)
	}

	return toBusSales(dbSales, items)
}

// Count returns the total number of sales in the DB.
func (s *Store) Count(ctx context.Context, filter salebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = "SELECT COUNT(id) AS `count` FROM sales"

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the sale identified by a given ID.
func (s *Store) QueryByID(ctx context.Context, slID uuid.UUID) (salebus.Sale, error) {

	data := struct {
		ID string `db:"id"`
	}{
		ID: slID.String(),
	}

	const q = `SELECT * FROM sales WHERE id = :id`

	var dbsl dbSale
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbsl); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return salebus.Sale{}, fmt.Errorf("namedquerystruct: %w", salebus.ErrNotFound)
		}
		return salebus.Sale{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	items, err := s.getSaleItems(ctx, []uuid.UUID{dbsl.ID})
	if err != nil {
		return salebus.Sale{}, fmt.Errorf("getSaleItems: %w", err)
	}

	return toBusSale(dbsl, items)
}

// QueryByID finds the sale identified by a given ID.
func (s *Store) getSaleItems(ctx context.Context, slID []uuid.UUID) ([]dbSaleItem, error) {

	data := struct {
		IDS []uuid.UUID `db:"sale_ids"`
	}{
		IDS: slID,
	}

	const q = `SELECT * FROM sale_items WHERE sale_id IN (:sale_ids) ORDER BY sale_id, product_id ASC`

	var dbItems []dbSaleItem
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbItems); err != nil {
		return dbItems, fmt.Errorf("namedquerystruct: %w", err)
	}

	return dbItems, nil
}
