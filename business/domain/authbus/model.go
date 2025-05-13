package authbus

import (
	"time"
)

// PasswordResetToken represents information about an individual key.
type PasswordResetToken struct {
	Email    string    `db:"email"`
	Token    string    `db:"token"`
	ExpiryAt time.Time `db:"expiry_at"`
}

// NewPasswordResetToken contains information needed to createPasswordReset a new key.
type NewPasswordResetToken struct {
	Email string
}

// Login is used for user login with email and password
type Login struct {
	Email    string
	Password string
}
