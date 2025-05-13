// Package authapp maintains the web based api for auth access.
package authapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/google/uuid"

	"github.com/rmsj/service/app/sdk/auth"
	"github.com/rmsj/service/app/sdk/authclient"
	"github.com/rmsj/service/app/sdk/errs"
	"github.com/rmsj/service/app/sdk/mid"
	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/foundation/web"
)

type app struct {
	auth    *auth.Auth
	authBus *authbus.Business
	userBus *userbus.Business
}

func newApp(ath *auth.Auth, authBus *authbus.Business, userBus *userbus.Business) *app {
	return &app{
		auth:    ath,
		authBus: authBus,
		userBus: userBus,
	}
}

func (a *app) token(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		return errs.NewFieldErrors("kid", errors.New("missing kid"))
	}

	// The BearerBasic middleware function generates the claims.
	claims := mid.GetClaims(ctx)

	tkn, refreshToken, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	return token{Token: tkn, RefreshToken: refreshToken}
}

func (a *app) authenticate(ctx context.Context, r *http.Request) web.Encoder {
	// The middleware is actually handling the authentication. So if the code
	// gets to this handler, authentication passed.

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	resp := authclient.AuthenticateResp{
		UserID: userID,
		Claims: mid.GetClaims(ctx),
	}

	return resp
}

func (a *app) authorize(ctx context.Context, r *http.Request) web.Encoder {
	var auth authclient.Authorize
	if err := web.Decode(r, &auth); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.auth.Authorize(ctx, auth.Claims, auth.UserID, auth.Rule); err != nil {
		return errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", auth.Claims.Roles, auth.Rule, err)
	}

	return nil
}

// login handles user login with username and password
func (a *app) login(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		kid = a.auth.ActiveKID()
		if kid == "" {
			return errs.New(errs.FailedPrecondition, errs.NewFieldErrors("kid", errors.New("missing kid")))
		}
	}

	// if we get to this point, we already have claims as the login itself happens in the middleware

	// The BearerBasic middleware function generates the claims.
	claims := mid.GetClaims(ctx)

	tkn, refreshToken, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	// update user with new refresh token
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.DataLoss, err)
	}

	usr, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "invalid user id: %s", userID)
		}
		return errs.Newf(errs.Internal, "error getting user to login")
	}
	_, err = a.userBus.Update(ctx, usr, userbus.UpdateUser{RefreshToken: &refreshToken})
	if err != nil {
		return errs.Newf(errs.Internal, "error updating user refresh token: email[%s]", usr.Email)
	}

	return token{Token: tkn, RefreshToken: refreshToken}
}

// refresh creates a new token for the logged in user, using the provided refresh token
func (a *app) refresh(ctx context.Context, r *http.Request) web.Encoder {
	kid := web.Param(r, "kid")
	if kid == "" {
		kid = a.auth.ActiveKID()
		if kid == "" {
			return errs.New(errs.FailedPrecondition, errs.NewFieldErrors("kid", errors.New("missing kid")))
		}
	}

	var app auth.RefreshToken
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// if we get to this point, we already have claims as the login itself happens in the middleware

	// The BearerBasic middleware function generates the claims.
	claims := mid.GetClaims(ctx)

	fmt.Printf("claims: %+v \n", claims)

	tkn, refreshToken, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		return errs.New(errs.Internal, err)
	}

	if err := a.updateRefreshToken(ctx, app.Token, refreshToken); err != nil {
		return errs.Newf(errs.FailedPrecondition, "error updating token")
	}

	return token{
		Token:        tkn,
		RefreshToken: refreshToken,
	}
}

// forgotPassword creates a forgot password token and sends via email to the user, if a valid email is provided, otherwise, do nothing
func (a *app) forgotPassword(ctx context.Context, r *http.Request) web.Encoder {

	var app NewPasswordResetToken
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// this creates the password reset and sends the email to the user
	_, err := a.createPasswordReset(ctx, app)
	if err != nil {
		// TODO: log error, always return empty - message in FE is if it's a valid email, should receive the reset link.
		return nil
	}

	return nil
}

// resetPassword is the actual reset endpoint - it is validated with middleware that validates the reset token.
func (a *app) resetPassword(ctx context.Context, r *http.Request) web.Encoder {

	var app ResetPassword
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	// this creates the password reset and sends the email to the user
	if err := a.updateUserPassword(ctx, userID, userbus.UpdateUser{
		Password:        &app.Password,
		PasswordConfirm: &app.PasswordConfirm,
	}); err != nil {
		return errs.Newf(errs.InvalidArgument, "error updating user password")
	}

	return nil
}

func (a *app) authenticateAPI(ctx context.Context, r *http.Request) web.Encoder {
	// The middleware is actually handling the authentication. So if the code
	// gets to this handler, authentication passed.

	return nil
}

// createPasswordReset adds a new password reset to the system.
func (a *app) createPasswordReset(ctx context.Context, app NewPasswordResetToken) (PasswordResetToken, error) {

	_, err := a.userBus.QueryByEmail(ctx, mail.Address{Address: app.Email})
	if err != nil {
		if errors.Is(err, userbus.ErrNotFound) {
			return PasswordResetToken{}, nil
		}
		return PasswordResetToken{}, errs.Newf(errs.Internal, "error checking user to create token: %s", err)
	}

	rt, err := a.authBus.CreatePasswordReset(ctx, toBusNewPasswordResetToken(app))
	if err != nil {
		return PasswordResetToken{}, errs.Newf(errs.Internal, "create password reset token: email[%s]: %s", app.Email, err)
	}

	// send the email - TODO
	//err = a.mailer.Send(email.Mail{
	//	To: mail.Address{
	//		Name:    usr.Name.String(),
	//		Address: app.Email,
	//	},
	//	Subject:  "Password Reset Request",
	//	Template: email.TmplResetPassword,
	//	Data:     rt,
	//})
	//if err != nil {
	//	fmt.Printf("error sending reset password email: %s \n", err)
	//	return PasswordResetToken{}, err
	//}

	return toAppPasswordResetToken(rt), nil
}

// deletePasswordResetByEmail removes a password reset token from the system.
//
//lint:ignore U1000 temp
func (a *app) deletePasswordResetByEmail(ctx context.Context, email string) error {

	pr, err := a.authBus.QueryPasswordResetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, authbus.ErrNotFound) {
			return nil
		}
		return errs.Newf(errs.Internal, "error deleting password reset token")
	}

	if err := a.authBus.DeletePasswordReset(ctx, pr); err != nil {
		return errs.Newf(errs.Internal, "error deleting password reset token")
	}

	return nil
}

// deletePasswordResetByToken removes a password reset token from the system.
//
//lint:ignore U1000 temp
func (a *app) deletePasswordResetByToken(ctx context.Context, token string) error {

	pr, err := a.authBus.QueryPasswordResetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, authbus.ErrNotFound) {
			return nil
		}
		return errs.Newf(errs.Internal, "error deleting password reset token")
	}

	if err := a.authBus.DeletePasswordReset(ctx, pr); err != nil {
		return errs.Newf(errs.Internal, "error deleting password reset token")
	}

	return nil
}

// queryPasswordResetByEmail gets a password reset token from the system using email
//
//lint:ignore U1000 temp
func (a *app) queryPasswordResetByEmail(ctx context.Context, email string) (PasswordResetToken, error) {

	pr, err := a.authBus.QueryPasswordResetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, authbus.ErrNotFound) {
			return PasswordResetToken{}, nil
		}
		return PasswordResetToken{}, errs.Newf(errs.Internal, "error querying password reset token")
	}

	return toAppPasswordResetToken(pr), nil
}

// queryPasswordResetByToken gets a password reset token from the system using token.
//
//lint:ignore U1000 temp
func (a *app) queryPasswordResetByToken(ctx context.Context, token string) (PasswordResetToken, error) {

	pr, err := a.authBus.QueryPasswordResetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, authbus.ErrNotFound) {
			return PasswordResetToken{}, nil
		}
		return PasswordResetToken{}, errs.Newf(errs.Internal, "error querying password reset token")
	}

	return toAppPasswordResetToken(pr), nil
}

// updateRefreshToken updates user refresh token after a new token is generated
func (a *app) updateRefreshToken(ctx context.Context, currentToken, newToken string) error {

	usr, err := a.userBus.QueryByRefreshToken(ctx, currentToken)
	if err != nil {
		return errs.Newf(errs.Internal, "error updating user token: %s", err)
	}

	_, err = a.userBus.Update(ctx, usr, userbus.UpdateUser{RefreshToken: &newToken})
	if err != nil {
		return errs.Newf(errs.Internal, "updating refresh token: email[%s]: %s", usr.Email, err)
	}

	return nil
}

// updateUserPassword updates user password using the reset password flow
func (a *app) updateUserPassword(ctx context.Context, userID uuid.UUID, up userbus.UpdateUser) error {

	usr, err := a.userBus.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "invalid contact id: %s", userID)
		}
		return errs.Newf(errs.Internal, "error getting user to update refresh token: email[%s]", usr.Email)
	}

	_, err = a.userBus.Update(ctx, usr, up)
	if err != nil {
		return errs.Newf(errs.Internal, "error updating refresh token: email[%s]", usr.Email)
	}

	return nil
}
