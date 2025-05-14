package saleapp

import (
	"github.com/rmsj/service/business/domain/salebus"

	"github.com/rmsj/service/business/sdk/order"
)

var defaultOrderBy = order.NewBy("sale_id", order.ASC)

var orderByFields = map[string]string{
	"sale_id": salebus.OrderByID,
}
