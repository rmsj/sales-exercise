package saleapp

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/salebus"
)

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("order_by"),
		ID:      values.Get("sale_id"),
	}

	return filter
}

func parseFilter(qp queryParams) (salebus.QueryFilter, error) {

	var filter salebus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return salebus.QueryFilter{}, errs.NewFieldErrors("id", err)
		}
		filter.ID = &id
	}

	return filter, nil
}
