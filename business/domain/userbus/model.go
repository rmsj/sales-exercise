package userbus

import (
	"net/mail"
	"time"

	"github.com/google/uuid"

	"github.com/rmsj/service/business/types/name"
	"github.com/rmsj/service/business/types/role"
)

// User represents information about an individual user.
type User struct {
	ID           uuid.UUID
	Name         name.Name
	Email        mail.Address
	Mobile       string
	ProfileImage string
	Roles        []role.Role
	PasswordHash []byte
	Department   name.Null
	Enabled      bool
	RefreshToken string
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name         name.Name
	Email        mail.Address
	Mobile       *string
	ProfileImage *string
	Roles        []role.Role
	Department   name.Null
	Password     string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name            *name.Name
	Email           *mail.Address
	Mobile          *string
	ProfileImage    *string
	Roles           []role.Role
	Department      *name.Null
	RefreshToken    *string
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}
