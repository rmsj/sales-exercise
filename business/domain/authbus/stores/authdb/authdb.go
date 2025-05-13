// Package authdb contains PasswordResetToken related CRUD functionality.
package authdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
)

// Store manages the set of APIs for PasswordResetToken database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (authbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// CreatePasswordReset inserts a new PasswordResetToken into the database.
func (s *Store) CreatePasswordReset(ctx context.Context, token authbus.PasswordResetToken) error {
	const q = `
	INSERT INTO password_reset_tokens
		(email, token, expiry_at)
	VALUES
		(:email, :token, :expiry_at)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPasswordResetToken(token)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// DeletePasswordReset removes a PasswordResetToken from the database.
func (s *Store) DeletePasswordReset(ctx context.Context, prt authbus.PasswordResetToken) error {
	const q = `
	DELETE FROM
		password_reset_tokens
	WHERE
		token = :token`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPasswordResetToken(prt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryPasswordResetByEmail gets the specified PasswordResetToken from the database.
func (s *Store) QueryPasswordResetByEmail(ctx context.Context, email string) (authbus.PasswordResetToken, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `SELECT * FROM password_reset_tokens WHERE email = :email`

	var dbKey PasswordResetToken
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbKey); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return authbus.PasswordResetToken{}, fmt.Errorf("namedquerystruct: %w", authbus.ErrNotFound)
		}
		return authbus.PasswordResetToken{}, fmt.Errorf("db: %w", err)
	}

	return toBusPasswordResetToken(dbKey), nil
}

// QueryPasswordResetByToken gets the specified PasswordResetToken from the database.
func (s *Store) QueryPasswordResetByToken(ctx context.Context, token string) (authbus.PasswordResetToken, error) {
	data := struct {
		Token string `db:"token"`
	}{
		Token: token,
	}

	const q = `SELECT * FROM password_reset_tokens WHERE token = :token`

	var dbKey PasswordResetToken
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbKey); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return authbus.PasswordResetToken{}, fmt.Errorf("namedquerystruct: %w", authbus.ErrNotFound)
		}
		return authbus.PasswordResetToken{}, fmt.Errorf("db: %w", err)
	}

	return toBusPasswordResetToken(dbKey), nil
}
