// Package saleapp maintains the app layer api for sale domain.
package saleapp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/app/sdk/mid"
	"github.com/rmsj/service/app/sdk/query"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/foundation/web"
)

type app struct {
	userBus    *userbus.Business
	productBus *productbus.Business
	saleBus    *salebus.Business
}

func newApp(user *userbus.Business, product *productbus.Business, sale *salebus.Business) *app {
	return &app{
		userBus:    user,
		productBus: product,
		saleBus:    sale,
	}
}

// newWithTx constructs a new Handlers value with the domain apis
// using a store transaction that was created via middleware.
func (a *app) newWithTx(ctx context.Context) (*app, error) {
	tx, err := mid.GetTran(ctx)

	if err != nil {
		return nil, err
	}

	userBus, err := a.userBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	productBus, err := a.productBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	saleBus, err := a.saleBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &app{
		userBus:    userBus,
		productBus: productBus,
		saleBus:    saleBus,
	}, nil

}

// Create adds a new sale to the system.
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var app NewSale
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	a, err := a.newWithTx(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "error while creating sales")
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "user missing in context: %s", err)
	}
	user, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		return errs.Newf(errs.Internal, "invalid user for sale: %s", err)
	}

	// validate products and get them if all good
	products, err := a.validateProductsInSale(ctx, app.Items)
	if err != nil {
		return errs.New(err.(*errs.Error).Code, err)
	}

	newSaleBus, err := toBusNewSale(user.ID, app, products)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	sl, err := a.saleBus.Create(ctx, newSaleBus)
	if err != nil {
		return errs.Newf(errs.Internal, "error creating sale: %s", err)
	}

	result, err := ToAppSale(sl, user, products)
	if err != nil {
		return errs.Newf(errs.Internal, "error parsing sale after creation - sale id[%s]: %s", sl.ID, err)
	}

	return result
}

// Delete removes a sale from the system.
func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	sID, err := a.saleID(r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	sl, err := a.saleBus.QueryByID(ctx, sID)
	if err != nil {
		if errors.Is(err, salebus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "invalid sale id: %s", sID)
		}
		return errs.Newf(errs.Internal, "error getting sale to delete: %s", err)
	}

	if err := a.saleBus.Delete(ctx, sl); err != nil {
		return errs.Newf(errs.Internal, "delete: saleID[%s]: %s", sl.ID, err)
	}

	return nil
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, salebus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	sls, err := a.saleBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.saleBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	//TODO: we could use a join or a view in the sale business layer / saledb OR
	// we could add this logic to the business layer

	var result []Sale
	for _, sl := range sls {
		saleUser, err := a.userBus.QueryByID(ctx, sl.UserID)
		if err != nil {
			return errs.Newf(errs.Internal, "error getting sale user: %s", err)
		}

		var pIDs []uuid.UUID
		for _, item := range sl.Items {
			pIDs = append(pIDs, item.ProductID)
		}
		products, err := a.productsForSale(ctx, pIDs)
		if err != nil {
			return errs.Newf(errs.Internal, "error getting products for sale: %s", err)
		}
		sale, err := ToAppSale(sl, saleUser, products)
		if err != nil {
			return errs.Newf(errs.Internal, "error parsing sale sale - sale id[%s]: %s", sl.ID, err)
		}
		result = append(result, sale)
	}

	return query.NewResult(result, total, pg)
}

func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	saleID, err := a.saleID(r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	sl, err := a.saleBus.QueryByID(ctx, saleID)
	if err != nil {
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	user, err := a.userBus.QueryByID(ctx, sl.UserID)
	if err != nil {
		return errs.Newf(errs.Internal, "invalid user for sale: %s - %v", sl.UserID, err)
	}

	var pIDs []uuid.UUID
	for _, item := range sl.Items {
		pIDs = append(pIDs, item.ProductID)
	}

	// get all products for this order
	products, err := a.productsForSale(ctx, pIDs)
	if err != nil {
		return errs.Newf(errs.Internal, "error getting products for sale: %s", err)
	}

	sale, err := ToAppSale(sl, user, products)
	if err != nil {
		return errs.Newf(errs.Internal, "error sale - sale id[%s]: %s", sl.ID, err)
	}
	return sale
}

func (a *app) validateProductsInSale(ctx context.Context, items []NewSaleItem) ([]productbus.Product, error) {
	// loop through items to get ids
	var pIDs []uuid.UUID
	for _, item := range items {
		id, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, errs.Newf(errs.InvalidArgument, "invalid product id: %s", item.ProductID)
		}
		pIDs = append(pIDs, id)
	}

	// get all products for this order
	products, err := a.productsForSale(ctx, pIDs)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "error getting products for sale: %s", err)
	}
	var itemsNotFound []string
	for _, item := range items {
		itemFound := false
		for _, p := range products {
			if item.ProductID == p.ID.String() {
				itemFound = true
			}
		}
		if !itemFound {
			itemsNotFound = append(itemsNotFound, item.ProductID)
		}
	}
	if len(itemsNotFound) > 0 {
		return nil, errs.Newf(errs.InvalidArgument, "invalid product id(s): %s", strings.Join(itemsNotFound, ", "))
	}

	return products, nil
}

func (a *app) productsForSale(ctx context.Context, pIDs []uuid.UUID) ([]productbus.Product, error) {

	// get all products for this order
	pg, err := page.Parse("1", "100") // this is the max we allow to simplify things
	if err != nil {
		return nil, errs.Newf(errs.Internal, "error parsing page to query products: %s", err)
	}
	products, err := a.productBus.Query(ctx, productbus.QueryFilter{
		IDs: pIDs,
	}, productbus.DefaultOrderBy, pg)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "error getting products for sale: %s", err)
	}

	return products, nil
}

func (a *app) saleID(r *http.Request) (uuid.UUID, error) {
	id := web.Param(r, "sale_id")
	if id == "" {
		return uuid.Nil, errs.Newf(errs.Internal, "sale id not in request")
	}
	return uuid.Parse(id)
}
