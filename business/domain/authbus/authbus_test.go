package authbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rmsj/fake"

	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/sdk/dbtest"
	"github.com/rmsj/service/business/sdk/unitest"
)

func Test_Auth(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Auth")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, queryPasswordReset(db.BusDomain, sd), "queryPasswordReset")
	unitest.Run(t, createPasswordReset(db.BusDomain), "createPasswordReset")
	unitest.Run(t, deletePasswordReset(db.BusDomain, sd), "deletePasswordReset")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	f, err := fake.New()
	if err != nil {
		panic(err)
	}

	tokenA, err := authbus.TestSeedPasswordResetToken(ctx, busDomain.Auth, f.Email())
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding keys : %w", err)
	}

	tokenB, err := authbus.TestSeedPasswordResetToken(ctx, busDomain.Auth, f.Email())
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding keys : %w", err)
	}

	// -------------------------------------------------------------------------

	sd := unitest.SeedData{
		PassResetTokens: []authbus.PasswordResetToken{tokenA, tokenB},
	}

	return sd, nil
}

// =============================================================================

func queryPasswordReset(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	table := []unitest.Table{
		{
			Name:    "bytoken",
			ExpResp: sd.PassResetTokens[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Auth.QueryPasswordResetByToken(ctx, sd.PassResetTokens[0].Token)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(authbus.PasswordResetToken)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(authbus.PasswordResetToken)

				// update db fields with default values
				if gotResp.Email == expResp.Email {
					expResp.ExpiryAt = gotResp.ExpiryAt
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byemail",
			ExpResp: sd.PassResetTokens[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Auth.QueryPasswordResetByEmail(ctx, sd.PassResetTokens[0].Email)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(authbus.PasswordResetToken)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(authbus.PasswordResetToken)

				// update db fields with default values
				if gotResp.Email == expResp.Email {
					expResp.ExpiryAt = gotResp.ExpiryAt
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createPasswordReset(busDomain dbtest.BusDomain) []unitest.Table {

	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: authbus.PasswordResetToken{
				Email: "user@email.com",
			},
			ExcFunc: func(ctx context.Context) any {
				nprt := authbus.NewPasswordResetToken{
					Email: "user@email.com",
				}

				resp, err := busDomain.Auth.CreatePasswordReset(ctx, nprt)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(authbus.PasswordResetToken)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(authbus.PasswordResetToken)
				expResp.Token = gotResp.Token
				expResp.ExpiryAt = gotResp.ExpiryAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func deletePasswordReset(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "key",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Auth.DeletePasswordReset(ctx, sd.PassResetTokens[1]); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
