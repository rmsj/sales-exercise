package mid

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/app/sdk/authclient"
	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/types/role"
	"github.com/rmsj/service/foundation/web"
)

// Authenticate is a middleware function that integrates with an authentication client
// to validate user credentials and attach user data to the request context.
func Authenticate(client *authclient.Client) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			resp, err := client.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			ctx = setUserID(ctx, resp.UserID)
			ctx = setClaims(ctx, resp.Claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// Bearer processes JWT authentication logic.
func Bearer(ath *auth.Auth) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			claims, err := ath.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if claims.Subject == "" {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = setUserID(ctx, subjectID)
			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// Basic processes basic authentication logic.
func Basic(ath *auth.Auth, userBus *userbus.Business) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			email, pass, ok := parseBasicAuth(r.Header.Get("authorization"))
			if !ok {
				return errs.Newf(errs.Unauthenticated, "invalid Basic auth")
			}

			addr, err := mail.ParseAddress(email)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			usr, err := userBus.Authenticate(ctx, *addr, pass)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID.String(),
					Issuer:    ath.Issuer(),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
				},
				Roles: role.ParseToString(usr.Roles),
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = setUserID(ctx, subjectID)
			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// Login processes username/password auth logic
func Login(ath *auth.Auth, userBus *userbus.Business) web.MidFunc {

	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {

			var app auth.Login
			if err := web.Decode(r, &app); err != nil {
				return errs.New(errs.InvalidArgument, err)
			}

			addr, err := mail.ParseAddress(app.Email)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			usr, err := userBus.Authenticate(ctx, *addr, app.Password)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			now := GetTime(ctx)

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   usr.ID.String(),
					Issuer:    ath.Issuer(),
					ExpiresAt: jwt.NewNumericDate(now.UTC().Add(8 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(now.UTC()),
				},
				Roles: role.ParseToString(usr.Roles),
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			ctx = setUserID(ctx, subjectID)
			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}
		return h
	}

	return m
}

// RefreshToken processes refresh token auth logic
func RefreshToken(ath *auth.Auth, userBus *userbus.Business) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			// to refresh the token, the current one must still be valid
			claims, err := ath.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if claims.Subject == "" {
				return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return errs.Newf(errs.Unauthenticated, "parsing subject: %s", err)
			}

			var app auth.RefreshToken
			if err := web.Decode(r, &app); err != nil {
				return errs.New(errs.InvalidArgument, err)
			}

			usr, err := userBus.QueryByRefreshToken(ctx, app.Token)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			if usr.ID != subjectID {
				return errs.New(errs.Unauthenticated, errors.New("invalid token"))
			}
			ctx = setUserID(ctx, usr.ID)

			now := GetTime(ctx)
			claims.ExpiresAt = jwt.NewNumericDate(now.UTC().Add(8 * time.Hour))
			claims.IssuedAt = jwt.NewNumericDate(now.UTC())

			ctx = setClaims(ctx, claims)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// ResetToken processes reset password token auth logic
func ResetToken(authBus *authbus.Business, userBus *userbus.Business) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			resetToken := web.Param(r, "reset_token")
			token, err := authBus.QueryPasswordResetByToken(ctx, resetToken)
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}
			now := GetTime(ctx)
			if token.ExpiryAt.Before(now) {
				return errs.New(errs.Unauthenticated, errors.New("password reset token has expired. please request a new token"))
			}
			usr, err := userBus.QueryByEmail(ctx, mail.Address{Address: token.Email})
			if err != nil {
				return errs.New(errs.Unauthenticated, err)
			}

			ctx = setUserID(ctx, usr.ID)

			return next(ctx, r)
		}

		return h
	}

	return m
}

// APIKey processes API key authentication logic.
func APIKey(ath *auth.Auth) web.MidFunc {
	m := func(next web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			authKey := r.Header.Get("hg-api-key")
			if authKey != ath.APIKey() {
				return errs.New(errs.Unauthenticated, errs.Newf(errs.Unauthenticated, "invalid API Key"))
			}

			return next(ctx, r)
		}

		return h
	}

	return m
}

func parseBasicAuth(auth string) (string, string, bool) {
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}

	username, password, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}
