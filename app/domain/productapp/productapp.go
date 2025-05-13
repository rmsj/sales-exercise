// Package productapp maintains the app layer api for the product domain.
package productapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/app/sdk/query"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/foundation/web"
)

type app struct {
	productBus *productbus.Business
}

func newApp(productBus *productbus.Business) *app {
	return &app{
		productBus: productBus,
	}
}

func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
	var app NewProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	np, err := toBusNewProduct(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prd, err := a.productBus.Create(ctx, np)
	if err != nil {
		return errs.Newf(errs.Internal, "create: prd[%+v]: %s", prd, err)
	}

	return toAppProduct(prd)
}

func (a *app) update(ctx context.Context, r *http.Request) web.Encoder {
	var app UpdateProduct
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	up, err := toBusUpdateProduct(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pID, err := a.productID(r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	prd, err := a.productBus.QueryByID(ctx, pID)
	if err != nil {
		if errors.Is(err, productbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "invalid product id: %s", pID)
		}
		return errs.Newf(errs.Internal, "error getting product to update - please try again or contact support")
	}

	updPrd, err := a.productBus.Update(ctx, prd, up)
	if err != nil {
		return errs.Newf(errs.Internal, "update: productID[%s] up[%+v]: %s", prd.ID, app, err)
	}

	return toAppProduct(updPrd)
}

func (a *app) delete(ctx context.Context, r *http.Request) web.Encoder {
	pID, err := a.productID(r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	prd, err := a.productBus.QueryByID(ctx, pID)
	if err != nil {
		if errors.Is(err, productbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "invalid product id: %s", pID)
		}
		return errs.Newf(errs.Internal, "error getting product to delete - please try again or contact support")
	}

	if err := a.productBus.Delete(ctx, prd); err != nil {
		return errs.Newf(errs.Internal, "delete: productID[%s]: %s", prd.ID, err)
	}

	return nil
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, productbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	prds, err := a.productBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.productBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppProducts(prds), total, page)
}

func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	productID, err := a.productID(r)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	prd, err := a.productBus.QueryByID(ctx, productID)
	if err != nil {
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppProduct(prd)
}

func (a *app) productID(r *http.Request) (uuid.UUID, error) {
	id := web.Param(r, "product_id")
	if id == "" {
		return uuid.Nil, errs.Newf(errs.Internal, "product id not in request")
	}
	return uuid.Parse(id)
}
