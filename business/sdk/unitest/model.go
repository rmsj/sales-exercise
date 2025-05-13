package unitest

import (
	"context"

	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users           []User
	Admins          []User
	Products        []productbus.Product
	Sales           []salebus.Sale
	PassResetTokens []authbus.PasswordResetToken
}

// Table represents fields needed for running an unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
