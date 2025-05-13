// Package id provides easy and centralized access to generate an Nullable
package id

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
)

type Nullable struct {
	uuid.UUID
}

// Value implements sql.Valuer so that UUIDs can be written to databases
// transparently. Currently, UUIDs map to strings. Please consult
// database-specific driver documentation for matching types.
func (id Nullable) Value() (driver.Value, error) {
	if id.UUID == uuid.Nil {
		return nil, nil
	}
	return id.String(), nil
}

var attempts = 0

// New generates a uuid V7 and returns it
func New() uuid.UUID {
	newID, err := uuid.NewV7()
	if err != nil {
		attempts++
		if attempts < 4 {
			return New()
		}
	}

	return newID
}

// NewString generates a uuid V7 and returns it as string
func NewString() string {
	newID, err := uuid.NewV7()
	if err != nil {
		attempts++
		if attempts < 4 {
			return New().String()
		}
	}

	return newID.String()
}

// NewRandomString generates a random with a given length
func NewRandomString(length int) (string, error) {
	if length <= 0 || length > 64 {
		return "", errors.New("size must be between 1 and 64")
	}
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// NullableIDValue sets an ID field value to nil or to a valid uuid
func NullableIDValue(idStr *string) *uuid.UUID {
	if idStr == nil {
		return nil
	}

	if *idStr == "" {
		return &uuid.Nil
	}

	id, err := uuid.Parse(*idStr)
	if err == nil {
		return &id
	}

	return nil
}
