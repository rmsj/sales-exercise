package dbtest

import (
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/authbus/stores/authdb"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/productbus/stores/productdb"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/domain/userbus/stores/userdb"
	"github.com/rmsj/service/business/sdk/delegate"
	"github.com/rmsj/service/foundation/logger"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate
	Auth     *authbus.Business
	Product  *productbus.Business
	User     *userbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	dlg := delegate.New(log)
	authBus := authbus.NewBusiness(log, authdb.NewStore(log, db))
	userBus := userbus.NewBusiness(log, dlg, userdb.NewStore(log, db, time.Hour))
	productBus := productbus.NewBusiness(log, dlg, productdb.NewStore(log, db))

	return BusDomain{
		Delegate: dlg,
		Auth:     authBus,
		Product:  productBus,
		User:     userBus,
	}
}
