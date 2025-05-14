package productapp

import (
	"net/http"
	"strconv"
	"strings"

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
	IDs     []string
	Name    string
	Price   string
}

func parseQueryParams(r *http.Request) queryParams {
	values := r.URL.Query()

	productIds := values.Get("product_ids")
	var ids []string
	if productIds != "" {
		ids = strings.Split(productIds, ",")
	}

	filter := queryParams{
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
		ID:      values.Get("product_id"),
		IDs:     ids,
		Name:    values.Get("name"),
		Price:   values.Get("price"),
	}

	return filter
}

func parseFilter(qp queryParams) (productbus.QueryFilter, error) {
	var filter productbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldErrors("product_id", err)
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		pName, err := name.Parse(qp.Name)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldErrors("name", err)
		}
		filter.Name = &pName
	}

	if qp.Price != "" {
		price, err := strconv.ParseFloat(qp.Price, 64)
		if err != nil {
			return productbus.QueryFilter{}, errs.NewFieldErrors("price", err)
		}
		filter.Price = &price
	}

	if len(qp.IDs) > 0 {
		for _, id := range qp.IDs {
			parsedID, err := uuid.Parse(id)
			if err != nil {
				return productbus.QueryFilter{}, errs.NewFieldErrors("price", err)
			}
			filter.IDs = append(filter.IDs, parsedID)
		}
	}

	return filter, nil
}
