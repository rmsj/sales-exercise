package saleapp

import (
	"github.com/rmsj/service/business/domain/salebus"
)

var orderByFields = map[string]string{
	"sale_id": salebus.OrderBySaleID,
	"amount":  salebus.OrderByAmount,
}
