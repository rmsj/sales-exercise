package saleapp

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/app/sdk/authclient"
	"github.com/rmsj/service/app/sdk/mid"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/salebus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/sdk/sqldb"
	"github.com/rmsj/service/foundation/logger"
	"github.com/rmsj/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	DB         *sqlx.DB
	UserBus    *userbus.Business
	ProductBus *productbus.Business
	SaleBus    *salebus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authenticate := mid.Authenticate(cfg.AuthClient)
	transaction := mid.BeginCommitRollback(cfg.Log, sqldb.NewBeginner(cfg.DB))

	api := newApp(cfg.UserBus, cfg.ProductBus, cfg.SaleBus)
	app.HandlerFunc(http.MethodGet, version, "/sales", api.query, authenticate)
	app.HandlerFunc(http.MethodGet, version, "/sales/{ sale_id}", api.queryByID, authenticate)
	app.HandlerFunc(http.MethodPost, version, "/sales", api.create, authenticate, transaction)
	app.HandlerFunc(http.MethodDelete, version, "/sales/{sale_id}", api.delete, authenticate, transaction)
}
