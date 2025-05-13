// Package userdb contains user related CRUD functionality.
package userdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/viccon/sturdyc"

	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/sdk/order"
	"github.com/rmsj/service/business/sdk/page"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log   *logger.Logger
	db    sqlx.ExtContext
	cache *sturdyc.Client[userbus.User]
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB, ttl time.Duration) *Store {
	const capacity = 10000
	const numShards = 10
	const evictionPercentage = 10

	return &Store{
		log:   log,
		db:    db,
		cache: sturdyc.New[userbus.User](capacity, numShards, 10*time.Second, evictionPercentage),
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log:   s.log,
		db:    ec,
		cache: s.cache,
	}

	return &store, nil
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr userbus.User) error {
	const q = `
	INSERT INTO users
		(id, name, email, password_hash, roles, department, enabled, created_at, updated_at)
	VALUES
		(:id, :name, :email, :password_hash, :roles, :department, :enabled, :created_at, :updated_at)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userbus.ErrUniqueEmail)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	s.writeCache(usr)

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr userbus.User) error {
	const q = `
	UPDATE
		users
	SET 
		name = :name,
		email = :email,
		roles = :roles,
		password_hash = :password_hash,
		department = :department,
		enabled = :enabled,
		updated_at = :updated_at
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return userbus.ErrUniqueEmail
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	s.writeCache(usr)

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr userbus.User) error {
	const q = `
	DELETE FROM
		users
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	s.deleteCache(usr)

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter userbus.QueryFilter, orderBy order.By, page page.Page) ([]userbus.User, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, email, password_hash, roles, department, enabled, created_at, updated_at
	FROM
		users`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")

	var dbUsrs []user
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUsers(dbUsrs)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter userbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = "SELECT COUNT(id) AS `count` FROM users"

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (userbus.User, error) {
	cached, ok := s.readCache(userID.String())
	if ok {
		return cached, nil
	}

	data := struct {
		ID string `db:"id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
        id, name, email, password_hash, roles, department, enabled, created_at, updated_at
	FROM
		users
	WHERE 
		id = :id`

	var dbUsr user
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userbus.User{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userbus.User{}, fmt.Errorf("db: %w", err)
	}
	bus, err := toBusUser(dbUsr)
	if err != nil {
		return userbus.User{}, err
	}
	s.writeCache(bus)

	return bus, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (userbus.User, error) {
	cached, ok := s.readCache(email.Address)
	if ok {
		return cached, nil
	}

	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	const q = `
	SELECT
        id, name, email, password_hash, roles, department, enabled, created_at, updated_at
	FROM
		users
	WHERE
		email = :email`

	var dbUsr user
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userbus.User{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userbus.User{}, fmt.Errorf("db: %w", err)
	}
	bus, err := toBusUser(dbUsr)
	if err != nil {
		return userbus.User{}, err
	}
	s.writeCache(bus)

	return bus, nil
}

// QueryByRefreshToken gets the specified user from the database by refresh token.
func (s *Store) QueryByRefreshToken(ctx context.Context, refreshToken string) (userbus.User, error) {
	cached, ok := s.readCache(refreshToken)
	if ok {
		return cached, nil
	}

	data := struct {
		Email string `db:"refresh_token"`
	}{
		Email: refreshToken,
	}

	const q = `
	SELECT
        *
	FROM
		users
	WHERE
		refresh_token = :refresh_token`

	var dbUsr user
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userbus.User{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userbus.User{}, fmt.Errorf("db: %w", err)
	}

	bus, err := toBusUser(dbUsr)
	if err != nil {
		return userbus.User{}, err
	}
	s.writeCache(bus)

	return bus, nil
}
