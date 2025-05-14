package productbus

import (
	"github.com/google/uuid"

	"github.com/rmsj/service/business/types/name"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID    *uuid.UUID
	IDs   []uuid.UUID
	Name  *name.Name
	Price *float64
}
