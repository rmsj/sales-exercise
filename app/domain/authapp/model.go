package authapp

import (
	"encoding/json"
	"time"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/authbus"
)

type token struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// Encode implements the encoder interface.
func (t token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}

// PasswordResetToken represents a password reset in the system
type PasswordResetToken struct {
	Email  string
	Token  string
	Expiry string
}

// Encode implements the encoder interface.
func (app PasswordResetToken) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewPasswordResetToken contains information needed to createPasswordReset a new key.
type NewPasswordResetToken struct {
	Email string `json:"email" validate:"required,email"`
}

// Decode implements the decoder interface.
func (app *NewPasswordResetToken) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewPasswordResetToken) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPasswordResetToken(app NewPasswordResetToken) authbus.NewPasswordResetToken {
	return authbus.NewPasswordResetToken{
		Email: app.Email,
	}
}

func toAppPasswordResetToken(bus authbus.PasswordResetToken) PasswordResetToken {
	return PasswordResetToken{
		Email:  bus.Email,
		Token:  bus.Token,
		Expiry: bus.ExpiryAt.Format(time.RFC3339),
	}
}

// ResetPassword contains information needed to reset user password
type ResetPassword struct {
	Password        string `json:"password" validate:"required,min=6"`
	PasswordConfirm string `json:"passwordConfirm" validate:"eqfield=Password"`
}

// Decode implements the decoder interface.
func (app *ResetPassword) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app ResetPassword) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
