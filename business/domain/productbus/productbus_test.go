package productbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/sdk/dbtest"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/sdk/unitest"
	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
)

func Test_Product(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Product")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	prds1, err := productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	// -------------------------------------------------------------------------

	prds2, err := productbus.TestGenerateSeedProducts(ctx, 2, busDomain.Product)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		Products: append(prds1, prds2...),
	}

	return sd, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	prds := make([]productbus.Product, 0, len(sd.Products))
	prds = append(prds, sd.Products...)

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() <= prds[j].ID.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: prds,
			ExcFunc: func(ctx context.Context) any {
				filter := productbus.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := busDomain.Product.Query(ctx, filter, productbus.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]productbus.Product)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Products[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Product.QueryByID(ctx, sd.Products[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(productbus.Product)

				if gotResp.DateCreated.Format(time.RFC3339) == expResp.DateCreated.Format(time.RFC3339) {
					expResp.DateCreated = gotResp.DateCreated
				}

				if gotResp.DateUpdated.Format(time.RFC3339) == expResp.DateUpdated.Format(time.RFC3339) {
					expResp.DateUpdated = gotResp.DateUpdated
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
			ExpResp: productbus.Product{
				Name:  name.MustParse("Guitar"),
				Price: money.MustParse(10.34),
			},
			ExcFunc: func(ctx context.Context) any {
				np := productbus.NewProduct{
					Name:  name.MustParse("Guitar"),
					Price: money.MustParse(10.34),
				}

				resp, err := busDomain.Product.Create(ctx, np)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(productbus.Product)

				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: productbus.Product{
				ID:          sd.Products[0].ID,
				Name:        name.MustParse("Guitar"),
				Price:       money.MustParse(10.34),
				DateCreated: sd.Products[0].DateCreated,
				DateUpdated: sd.Products[0].DateCreated,
			},
			ExcFunc: func(ctx context.Context) any {
				up := productbus.UpdateProduct{
					Name:  dbtest.NamePointer("Guitar"),
					Price: dbtest.MoneyPointer(10.34),
				}

				resp, err := busDomain.Product.Update(ctx, sd.Products[0], up)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(productbus.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(productbus.Product)

				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.Delete(ctx, sd.Products[1]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Product.Delete(ctx, sd.Products[1]); err != nil {
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
