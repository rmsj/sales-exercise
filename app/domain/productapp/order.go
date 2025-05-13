package productapp

import (
	"github.com/rmsj/service/business/domain/productbus"
)

var orderByFields = map[string]string{
	"product_id": productbus.OrderByProductID,
	"name":       productbus.OrderByName,
	"price":      productbus.OrderByPrice,
	"user_id":    productbus.OrderByUserID,
}
