// Package vproductdb provides access to the product view.
package vproductdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/domain/vproductbus"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
)

// Store manages the set of APIs for product view database access.
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

// Query retrieves a list of existing products from the database.
func (s *Store) Query(ctx context.Context, filter vproductbus.QueryFilter, orderBy order.By, page page.Page) ([]vproductbus.Product, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		product_id,
		user_id,
		name,
		cost,
		quantity,
		created_at,
		updated_at,
		user_name
	FROM
		view_products`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dnPrd []product
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dnPrd); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	prd, err := toBusProducts(dnPrd)
	if err != nil {
		return nil, err
	}

	return prd, nil
}

// Count returns the total number of products in the DB.
func (s *Store) Count(ctx context.Context, filter vproductbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = "SELECT COUNT(product_id) AS `count` FROM view_products"

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
