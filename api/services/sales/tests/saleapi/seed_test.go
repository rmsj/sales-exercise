package saleapi_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/rmsj/service/app/sdk/apitest"
	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/sdk/dbtest"
	"github.com/rmsj/service/business/types/role"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsers(ctx, 1, role.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := productbus.TestGenerateSeedProducts(ctx, 3, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}
	sort.Slice(prds, func(i, j int) bool {
		return prds[i].ID.String() < prds[j].ID.String()
	})
	var items []salebus.NewSaleItem
	for _, prd := range prds {
		items = append(items, salebus.NewSaleItem{
			ProductID: prd.ID,
			Quantity:  1,
			Price:     prd.Price,
		})
	}

	sales1, err := salebus.TestSeedSales(ctx, 5, usrs[0].ID, items, busDomain.Sale)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding sales : %w", err)
	}

	td1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 1, role.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	sales2, err := salebus.TestSeedSales(ctx, 5, usrs[0].ID, items, busDomain.Sale)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding sales : %w", err)
	}

	td2 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(db.BusDomain.User, ath, usrs[0].Email.Address),
	}

	sd := apitest.SeedData{
		Users:    []apitest.User{td1, td2},
		Products: prds,
		Sales:    append(sales1, sales2...),
	}

	return sd, nil
}
