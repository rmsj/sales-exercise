package authdb

import (
	"time"

	"github.com/rmsj/service/business/domain/authbus"
)

type PasswordResetToken struct {
	Email    string    `db:"email"`
	Token    string    `db:"token"`
	ExpiryAt time.Time `db:"expiry_at"`
}

func toDBPasswordResetToken(bus authbus.PasswordResetToken) PasswordResetToken {

	return PasswordResetToken{
		Email:    bus.Email,
		Token:    bus.Token,
		ExpiryAt: bus.ExpiryAt,
	}
}

func toBusPasswordResetToken(db PasswordResetToken) authbus.PasswordResetToken {
	return authbus.PasswordResetToken{
		Email:    db.Email,
		Token:    db.Token,
		ExpiryAt: db.ExpiryAt,
	}
}
