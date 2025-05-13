// Package ctxval provides a way to get some specific values from the context
package ctxval

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// this needs to be the same as the ones in app/sdk/mid
const (
	timeKey   = "time"
	userIDKey = "auth-user"
)

// GetTime returns the start time of the current request from the context.
func GetTime(ctx context.Context) time.Time {
	v, ok := ctx.Value(timeKey).(time.Time)
	if !ok {
		return time.Now()
	}
	return v
}

// GetAuthUserID returns the logged-in user saved in the context, or empty string
func GetAuthUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return uuid.Nil, errors.New("no user id in context")
	}
	userID, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}
