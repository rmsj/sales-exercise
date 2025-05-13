package salebus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/role"

	"github.com/rmsj/service/business/sdk/dbtest"
	"github.com/rmsj/service/business/sdk/unitest"
)

func Test_Sale(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Sale")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsers(ctx, 1, role.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := productbus.TestGenerateSeedProducts(ctx, 3, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	var items []salebus.NewSaleItem
	for _, prd := range prds {
		items = append(items, salebus.NewSaleItem{
			ProductID: prd.ID,
			Quantity:  1,
			Price:     prd.Price,
		})
	}

	sales1, err := salebus.TestSeedSales(ctx, 1, usrs[0].ID, items, busDomain.Sale)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding sales : %w", err)
	}

	td1 := unitest.User{
		User: usrs[0],
	}
	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 1, role.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	items = nil
	for _, prd := range prds {
		items = append(items, salebus.NewSaleItem{
			ProductID: prd.ID,
			Quantity:  2,
			Price:     prd.Price,
		})
	}

	sales2, err := salebus.TestSeedSales(ctx, 1, usrs[0].ID, items, busDomain.Sale)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding sales : %w", err)
	}

	td2 := unitest.User{
		User: usrs[0],
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Users:    []unitest.User{td1, td2},
		Products: prds,
		Sales:    append(sales1, sales2...),
	}

	return sd, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	sls := make([]salebus.Sale, 0, len(sd.Sales))
	sls = append(sls, sd.Sales...)

	sort.Slice(sls, func(i, j int) bool {
		return sls[i].ID.String() <= sls[j].ID.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: sls,
			ExcFunc: func(ctx context.Context) any {
				filter := salebus.QueryFilter{}

				resp, err := busDomain.Sale.Query(ctx, filter, salebus.DefaultOrderBy, page.MustParse("1", "20"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]salebus.Sale)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]salebus.Sale)

				for i := range gotResp {
					if gotResp[i].ID == expResp[i].ID {
						expResp[i].UpdatedAt = gotResp[i].UpdatedAt
						expResp[i].CreatedAt = gotResp[i].CreatedAt
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Sales[1],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Sale.QueryByID(ctx, sd.Sales[1].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(salebus.Sale)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(salebus.Sale)

				if gotResp.ID == expResp.ID {
					expResp.UpdatedAt = gotResp.UpdatedAt
					expResp.CreatedAt = gotResp.CreatedAt
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: salebus.Sale{
				UserID: sd.Users[0].User.ID,
				Items: []salebus.SaleItem{
					{
						ProductID: sd.Products[0].ID,
						Quantity:  1,
						Amount:    sd.Products[0].Price,
					},
					{
						ProductID: sd.Products[0].ID,
						Quantity:  2,
						Amount:    money.MustParse(sd.Products[0].Price.Value() * 2),
					},
				},
			},
			ExcFunc: func(ctx context.Context) any {
				ng := salebus.NewSale{
					UserID: sd.Users[0].User.ID,
					Items: []salebus.NewSaleItem{
						{
							ProductID: sd.Products[0].ID,
							Quantity:  1,
							Price:     sd.Products[0].Price,
						},
						{
							ProductID: sd.Products[1].ID,
							Quantity:  2,
							Price:     sd.Products[1].Price,
						},
					},
				}

				resp, err := busDomain.Sale.Create(ctx, ng)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(salebus.Sale)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(salebus.Sale)

				expResp.ID = gotResp.ID
				expResp.UpdatedAt = gotResp.UpdatedAt
				expResp.CreatedAt = gotResp.CreatedAt

				for i := range gotResp.Items {
					if gotResp.Items[i].ProductID == expResp.Items[i].ProductID {
						expResp.Items[i].SaleID = gotResp.Items[i].SaleID
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

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "sale-delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Sale.Delete(ctx, sd.Sales[1]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
