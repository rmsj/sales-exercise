package saleapi_test

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"

	"github.com/rmsj/service/app/domain/saleapp"
	"github.com/rmsj/service/app/sdk/apitest"
	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/app/sdk/query"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
)

func query200(sd apitest.SeedData) []apitest.Table {
	sls := make([]salebus.Sale, 0, len(sd.Sales))
	sls = append(sls, sd.Sales...)

	var result []saleapp.Sale
	for _, sl := range sls {

		var saleUser userbus.User
		for _, user := range sd.Users {
			if user.ID == sl.UserID {
				saleUser = user.User
			}
		}

		sale, err := saleapp.ToAppSale(sl, saleUser, sd.Products)
		if err != nil {
			panic(err)
		}
		result = append(result, sale)
	}

	sort.Slice(sls, func(i, j int) bool {
		return sls[i].ID.String() < sls[j].ID.String()
	})
	for i := range sls {
		sort.Slice(sls[i].Items, func(k, l int) bool {
			return sls[i].Items[k].ProductID.String() < sls[i].Items[l].ProductID.String()
		})
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales?page=1&rows=10&order_by=sale_id,ASC&name=Name",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[saleapp.Sale]{},
			ExpResp: &query.Result[saleapp.Sale]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sls),
				Items:       result,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*query.Result[saleapp.Sale])
				expResp := exp.(*query.Result[saleapp.Sale])

				for i := range gotResp.Items {
					// update db fields with default values
					if gotResp.Items[i].ID == expResp.Items[i].ID {
						expResp.Items[i].UpdatedAt = gotResp.Items[i].UpdatedAt
						expResp.Items[i].CreatedAt = gotResp.Items[i].CreatedAt
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func query400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/v1/sales?page=1&rows=10&sale_id=invalid_uuid",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"id\",\"error\":\"invalid UUID length: 12\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-orderby-value",
			URL:        "/v1/sales?page=1&rows=10&order_by=ale_id,ASC",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: ale_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {

	var saleUser userbus.User
	for _, user := range sd.Users {
		if user.ID == sd.Sales[0].UserID {
			saleUser = user.User
		}
	}

	sale, err := saleapp.ToAppSale(sd.Sales[0], saleUser, sd.Products)
	if err != nil {
		panic(err)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/sales/%s", sd.Sales[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &saleapp.Sale{},
			ExpResp:    &sale,
			CmpFunc: func(got any, exp any) string {
				resp := got.(*saleapp.Sale)
				expResp := exp.(*saleapp.Sale)

				expResp.ID = resp.ID
				// fields with default values, database or otherwise...
				expResp.UpdatedAt = resp.UpdatedAt
				expResp.CreatedAt = resp.CreatedAt

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
