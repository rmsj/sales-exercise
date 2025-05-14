package saleapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/rmsj/service/app/domain/saleapp"
	"github.com/rmsj/service/app/sdk/apitest"
	"github.com/rmsj/service/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &saleapp.NewSale{
				Items: []saleapp.NewSaleItem{
					{
						ProductID: sd.Products[0].ID.String(),
						Quantity:  1,
					},
					{
						ProductID: sd.Products[1].ID.String(),
						Quantity:  2,
					},
				},
			},
			GotResp: &saleapp.Sale{},
			ExpResp: &saleapp.Sale{
				Amount: sd.Products[0].Price.Value() + sd.Products[1].Price.Value()*2,
				Customer: saleapp.Customer{
					ID:    sd.Users[0].ID.String(),
					Name:  sd.Users[0].Name.String(),
					Email: sd.Users[0].Email.Address,
				},
				Items: []saleapp.Item{
					{
						ID:         sd.Products[0].ID.String(),
						Name:       sd.Products[0].Name.String(),
						UnityPrice: sd.Products[0].Price.Value(),
						Quantity:   1,
						Amount:     sd.Products[0].Price.Value(),
						Discount:   0,
					},
					{
						ID:         sd.Products[1].ID.String(),
						Name:       sd.Products[1].Name.String(),
						UnityPrice: sd.Products[1].Price.Value(),
						Quantity:   2,
						Amount:     sd.Products[1].Price.Value() * 2,
						Discount:   0,
					},
				},
			},
			CmpFunc: func(got any, exp any) string {
				gotResp := got.(*saleapp.Sale)
				expResp := exp.(*saleapp.Sale)

				expResp.ID = gotResp.ID
				expResp.UpdatedAt = gotResp.UpdatedAt
				expResp.CreatedAt = gotResp.CreatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-input",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &saleapp.NewSale{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"items\",\"error\":\"items is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-input-negative-discount",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &saleapp.NewSale{
				Discount: -10,
				Items: []saleapp.NewSaleItem{
					{
						ProductID: sd.Products[0].ID.String(),
						Quantity:  1,
					},
					{
						ProductID: sd.Products[1].ID.String(),
						Quantity:  2,
					},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"discount\",\"error\":\"discount must be 0 or greater\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-input-negative-discount",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &saleapp.NewSale{
				Discount: 1000001,
				Items: []saleapp.NewSaleItem{
					{
						ProductID: sd.Products[0].ID.String(),
						Quantity:  1,
					},
					{
						ProductID: sd.Products[1].ID.String(),
						Quantity:  2,
					},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"discount\",\"error\":\"discount must be 1,000,000 or less\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/sales",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badtoken",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        "/v1/sales",
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
