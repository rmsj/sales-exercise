// Package authbus provides business access to key domain.
package authbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rmsj/service/business/sdk/ctxval"
	"github.com/rmsj/service/business/sdk/id"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
	"github.com/rmsj/service/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound      = errors.New("key not found")
	ErrEmailRequired = errors.New("email required to createPasswordReset token")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	CreatePasswordReset(ctx context.Context, token PasswordResetToken) error
	DeletePasswordReset(ctx context.Context, token PasswordResetToken) error
	QueryPasswordResetByEmail(ctx context.Context, email string) (PasswordResetToken, error)
	QueryPasswordResetByToken(ctx context.Context, token string) (PasswordResetToken, error)
}

// Business manages the set of APIs for key access.mi
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a key business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		b.log.Error(context.Background(), "business.authbus.newwithtx", "error", err)
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// CreatePasswordReset adds a new password reset token to the system.
func (b *Business) CreatePasswordReset(ctx context.Context, np NewPasswordResetToken) (PasswordResetToken, error) {
	ctx, span := otel.AddSpan(ctx, "business.authbus.createpasswordreset")
	defer span.End()

	now := ctxval.GetTime(ctx)

	randStr, err := id.NewRandomString(32)
	if err != nil {
		b.log.Error(ctx, "business.authbus.createpasswordreset", "error", err)
		return PasswordResetToken{}, fmt.Errorf("createPasswordReset: %w", err)
	}

	pst := PasswordResetToken{
		Email:    np.Email,
		Token:    randStr,
		ExpiryAt: now.Add(1 * time.Hour),
	}

	if err := b.storer.CreatePasswordReset(ctx, pst); err != nil {
		b.log.Error(ctx, "business.authbus.createpasswordreset", "error", err)
		return PasswordResetToken{}, fmt.Errorf("createPasswordReset: %w", err)
	}

	return pst, nil
}

// DeletePasswordReset removes the specified password reset token.
func (b *Business) DeletePasswordReset(ctx context.Context, key PasswordResetToken) error {
	ctx, span := otel.AddSpan(ctx, "business.authbus.deletepasswordreset")
	defer span.End()

	if err := b.storer.DeletePasswordReset(ctx, key); err != nil {
		b.log.Error(ctx, "business.authbus.deletepasswordreset", "error", err)
		return fmt.Errorf("deletePasswordReset: %w", err)
	}

	return nil
}

// QueryPasswordResetByEmail finds the key by the specified email.
func (b *Business) QueryPasswordResetByEmail(ctx context.Context, email string) (PasswordResetToken, error) {
	ctx, span := otel.AddSpan(ctx, "business.authbus.querypasswordresetbyemail")
	defer span.End()

	key, err := b.storer.QueryPasswordResetByEmail(ctx, email)
	if err != nil {
		b.log.Error(ctx, "business.authbus.querypasswordresetbyemail", "error", err)
		return PasswordResetToken{}, fmt.Errorf("queryPasswordReset password reset token: email[%s]: %w", email, err)
	}

	return key, nil
}

// QueryPasswordResetByToken finds the key by the specified token.
func (b *Business) QueryPasswordResetByToken(ctx context.Context, token string) (PasswordResetToken, error) {
	ctx, span := otel.AddSpan(ctx, "business.authbus.querypasswordresetbytoken")
	defer span.End()

	prt, err := b.storer.QueryPasswordResetByToken(ctx, token)
	if err != nil {
		b.log.Error(ctx, "business.authbus.querypasswordresetbytoken", "error", err)
		return PasswordResetToken{}, fmt.Errorf("queryPasswordReset password reset token: token[%s]: %w", token, err)
	}

	return prt, nil
}
