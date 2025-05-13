// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/sdk/migrate"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/docker"
	"github.com/rmsj/service/foundation/logger"
	"github.com/rmsj/service/foundation/otel"
)

// Database owns state for running and shutting down tests.
type Database struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	BusDomain BusDomain
}

// New creates a new test database inside the database that was started
// to handle testing. The database is migrated to the current version and
// a connection pool is provided with business domain packages.
func New(t *testing.T, testName string) *Database {
	image := "mysql:9.2.0"
	name := "servicetest"
	port := "3306"
	dockerArgs := []string{
		"-e", "MYSQL_ROOT_PASSWORD=root_password",
		"-e", "MYSQL_USER=db_user",
		"-e", "MYSQL_PASSWORD=db_password",
		"--health-cmd", "mysqladmin ping -h 127.0.0.1 --silent --wait=30",
	}

	c, err := docker.StartContainer(image, name, port, dockerArgs, nil)
	if err != nil {
		t.Fatalf("Starting database: %v", err)
	}

	t.Logf("Name    : %s\n", c.Name)
	t.Logf("HostPort: %s\n", c.HostPort)

	dbM, err := sqldb.Open(sqldb.Config{
		User:       "root",
		Password:   "root_password",
		Host:       c.HostPort,
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqldb.StatusCheck(ctx, dbM); err != nil {
		t.Fatalf("status check database: %v", err)
	}

	// -------------------------------------------------------------------------

	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	dbName := string(b)

	t.Logf("Create Database: %s\n", dbName)
	if _, err := dbM.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
		t.Fatalf("creating database %s: %v", dbName, err)
	}
	if _, err := dbM.ExecContext(context.Background(), "GRANT ALL PRIVILEGES ON "+dbName+".* TO 'db_user'@'%' "); err != nil {
		t.Fatalf("creating database %s: %v", dbName, err)
	}

	// -------------------------------------------------------------------------

	db, err := sqldb.Open(sqldb.Config{
		User:       "db_user",
		Password:   "db_password",
		Host:       c.HostPort,
		Name:       dbName,
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Logf("Migrate Database: %s\n", dbName)
	if err := migrate.Migrate(ctx, db); err != nil {
		t.Logf("Logs for %s\n%s:", c.Name, docker.DumpContainerLogs(c.Name))
		t.Fatalf("Migrating error: %s", err)
	}

	// -------------------------------------------------------------------------

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(ctx) })

	// -------------------------------------------------------------------------

	t.Cleanup(func() {
		t.Helper()

		t.Logf("Drop Database: %s\n", dbName)
		if _, err := dbM.ExecContext(context.Background(), "DROP DATABASE "+dbName); err != nil {
			t.Fatalf("dropping database %s: %v", dbName, err)
		}

		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = dbM.Close()
		if err != nil {
			return
		}

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	return &Database{
		DB:        db,
		Log:       log,
		BusDomain: newBusDomains(log, db),
	}
}
