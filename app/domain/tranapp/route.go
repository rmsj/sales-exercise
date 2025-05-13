package tranapp

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/app/sdk/authclient"
	"github.com/rmsj/service/app/sdk/mid"
	"github.com/rmsj/service/business/domain/productbus"
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
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	transaction := mid.BeginCommitRollback(cfg.Log, sqldb.NewBeginner(cfg.DB))
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newApp(cfg.UserBus, cfg.ProductBus)

	app.HandlerFunc(http.MethodPost, version, "/tranexample", api.create, authen, ruleAdmin, transaction)
}
