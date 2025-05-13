package productbus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/rmsj/service/business/types/money"
	"github.com/rmsj/service/business/types/name"
)

// TestGenerateNewProducts is a helper method for testing.
func TestGenerateNewProducts(n int) []NewProduct {
	newPrds := make([]NewProduct, n)

	idx := rand.Intn(10000)
	for i := range n {
		idx++

		np := NewProduct{
			Name:  name.MustParse(fmt.Sprintf("Name%d", idx)),
			Price: money.MustParse(float64(rand.Intn(500))),
		}

		newPrds[i] = np
	}

	return newPrds
}

// TestGenerateSeedProducts is a helper method for testing.
func TestGenerateSeedProducts(ctx context.Context, n int, api *Business) ([]Product, error) {
	newPrds := TestGenerateNewProducts(n)

	prds := make([]Product, len(newPrds))
	for i, np := range newPrds {
		prd, err := api.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding product: idx: %d : %w", i, err)
		}

		prds[i] = prd
	}

	return prds, nil
}
