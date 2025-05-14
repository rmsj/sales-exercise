package apitest

import (
	"net/http/httptest"
	"testing"

	authbuild "github.com/rmsj/service/api/services/auth/build/all"
	salesbuild "github.com/rmsj/service/api/services/sales/build/all"
	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/app/sdk/authclient"
	"github.com/rmsj/service/app/sdk/mux"
	"github.com/rmsj/service/business/sdk/dbtest"
)

// New initialized the system to run a test.
func New(t *testing.T, testName string) *Test {
	db := dbtest.New(t, testName)

	// -------------------------------------------------------------------------

	ath, err := auth.New(auth.Config{
		Log:       db.Log,
		UserBus:   db.BusDomain.User,
		KeyLookup: &KeyStore{},
		APIKey:    "api_key",
		ActiveKID: "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1",
	})
	if err != nil {
		t.Fatal(err)
	}

	// -------------------------------------------------------------------------

	server := httptest.NewServer(mux.WebAPI(mux.Config{
		Log: db.Log,
		DB:  db.DB,
		BusConfig: mux.BusConfig{
			AuthBus: db.BusDomain.Auth,
			UserBus: db.BusDomain.User,
		},
		AuthConfig: mux.AuthConfig{
			Auth: ath,
		},
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, server.URL)

	// -------------------------------------------------------------------------

	tMux := mux.WebAPI(mux.Config{
		Log: db.Log,
		DB:  db.DB,
		BusConfig: mux.BusConfig{
			AuthBus:    db.BusDomain.Auth,
			UserBus:    db.BusDomain.User,
			ProductBus: db.BusDomain.Product,
			SaleBus:    db.BusDomain.Sale,
		},
		SalesConfig: mux.SalesConfig{
			AuthClient: authClient,
		},
	}, salesbuild.Routes())

	return &Test{
		DB:   db,
		Auth: ath,
		mux:  tMux,
	}
}
