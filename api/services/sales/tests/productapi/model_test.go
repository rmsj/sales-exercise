package product_test

import (
	"time"

	"github.com/rmsj/service/app/domain/productapp"
	"github.com/rmsj/service/business/domain/productbus"
)

func toAppProduct(prd productbus.Product) productapp.Product {
	return productapp.Product{
		ID:          prd.ID.String(),
		Name:        prd.Name.String(),
		Price:       prd.Price.Value(),
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProductPtr(prd productbus.Product) *productapp.Product {
	appPrd := toAppProduct(prd)
	return &appPrd
}

func toAppProducts(prds []productbus.Product) []productapp.Product {
	items := make([]productapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}
