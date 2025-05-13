package productapp

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/types/name"
)

type queryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
	Price   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("product_id"),
		Name:    values.Get("name"),
		Price:   values.Get("price"),
	}

	return filter
}

func parseFilter(qp queryParams) (productbus.QueryFilter, error) {
	var fieldErrors errs.FieldErrors
	var filter productbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		switch err {
		case nil:
			filter.ID = &id
		default:
			fieldErrors.Add("product_id", err)
		}
	}

	if qp.Name != "" {
		name, err := name.Parse(qp.Name)
		switch err {
		case nil:
			filter.Name = &name
		default:
			fieldErrors.Add("name", err)
		}
	}

	if qp.Price != "" {
		cst, err := strconv.ParseFloat(qp.Price, 64)
		switch err {
		case nil:
			filter.Price = &cst
		default:
			fieldErrors.Add("price", err)
		}
	}

	if fieldErrors != nil {
		return productbus.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}
