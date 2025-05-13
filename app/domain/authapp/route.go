package authapp

import (
	"net/http"

	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/app/sdk/mid"
	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	AuthBus *authbus.Business
	UserBus *userbus.Business
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	bearer := mid.Bearer(cfg.Auth)
	apiKey := mid.APIKey(cfg.Auth)
	basic := mid.Basic(cfg.Auth, cfg.UserBus)
	login := mid.Login(cfg.Auth, cfg.UserBus)
	refresh := mid.RefreshToken(cfg.Auth, cfg.UserBus)
	resetPass := mid.ResetToken(cfg.AuthBus, cfg.UserBus)

	api := newApp(cfg.Auth, cfg.AuthBus, cfg.UserBus)

	app.HandlerFunc(http.MethodGet, version, "/auth/token/{kid}", api.token, basic)
	app.HandlerFunc(http.MethodPost, version, "/auth/login", api.login, login)
	app.HandlerFunc(http.MethodPost, version, "/auth/refresh", api.refresh, refresh)
	app.HandlerFunc(http.MethodPost, version, "/auth/forgot", api.forgotPassword)
	app.HandlerFunc(http.MethodPost, version, "/auth/reset-password/{reset_token}", api.resetPassword, resetPass)
	app.HandlerFunc(http.MethodGet, version, "/auth/authenticate", api.authenticate, bearer)
	app.HandlerFunc(http.MethodGet, version, "/auth/authenticate-api", api.authenticateAPI, apiKey)
	app.HandlerFunc(http.MethodPost, version, "/auth/authorize", api.authorize)
}
