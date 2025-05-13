package auth

import (
	"encoding/json"

	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/authbus"
)

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Decode implements the decoder interface.
func (app *Login) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app Login) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func (app Login) ToBusLogin() authbus.Login {

	return authbus.Login{
		Email:    app.Email,
		Password: app.Password,
	}
}

type RefreshToken struct {
	Token string `json:"refreshToken" validate:"required"`
}

// Decode implements the decoder interface.
func (app *RefreshToken) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app RefreshToken) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
