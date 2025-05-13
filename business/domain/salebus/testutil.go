package salebus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/types/money"
)

// TestGenerateSales is a helper method for testing.
func TestGenerateSales(n int, userID uuid.UUID, items []NewSaleItem) []NewSale {
	newSls := make([]NewSale, n)

	idx := rand.Intn(10000)

	for i := 0; i < n; i++ {
		ns := NewSale{
			UserID:   userID,
			Discount: money.MustParse(float64(rand.Intn(10))),
			Items:    items,
		}

		newSls[i] = ns
		idx++
	}

	return newSls
}

// TestSeedSales is a helper method for testing.
func TestSeedSales(ctx context.Context, n int, userID uuid.UUID, items []NewSaleItem, api *Business) ([]Sale, error) {
	newSls := TestGenerateSales(n, userID, items)

	sls := make([]Sale, len(newSls))
	for i, ns := range newSls {
		sl, err := api.Create(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding sales: idx: %d : %w", i, err)
		}

		sls[i] = sl
	}

	return sls, nil
}
