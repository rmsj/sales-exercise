package saledb

import (
	"fmt"

	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/sdk/order"
)

var orderByFields = map[string]string{
	salebus.OrderBySaleID: "id",
	salebus.OrderByAmount: "amount",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
