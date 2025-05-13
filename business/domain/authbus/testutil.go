package authbus

import (
	"context"
	"fmt"
)

// TestNewPasswordResetToken is a helper method for testing.
func TestNewPasswordResetToken(email string) NewPasswordResetToken {
	npr := NewPasswordResetToken{
		Email: email,
	}

	return npr
}

// TestSeedPasswordResetToken is a helper method for testing.
func TestSeedPasswordResetToken(ctx context.Context, api *Business, email string) (PasswordResetToken, error) {
	nu := NewPasswordResetToken{
		Email: email,
	}
	prt, err := api.CreatePasswordReset(ctx, nu)
	if err != nil {
		return PasswordResetToken{}, fmt.Errorf("seeding password reset token: %w", err)
	}

	return prt, nil
}
