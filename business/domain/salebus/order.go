package salebus

import "github.com/rmsj/service/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderBySaleID, order.ASC)

// Set of fields that the results can be ordered by. These are the names
// that should be used by the application layer.
const (
	OrderBySaleID = "a"
	OrderByAmount = "b"
)
