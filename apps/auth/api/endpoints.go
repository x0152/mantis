package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/auth/use_cases"
	"mantis/core/auth"
	"mantis/core/base"
)

type UseCases struct {
	Login *usecases.Login
	Me    *usecases.Me
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "auth-login", Method: http.MethodPost, Path: "/api/auth/login"}, e.login)
	huma.Register(api, huma.Operation{OperationID: "auth-logout", Method: http.MethodPost, Path: "/api/auth/logout"}, e.logout)
	huma.Register(api, huma.Operation{OperationID: "auth-me", Method: http.MethodGet, Path: "/api/auth/me"}, e.me)
}

func (e *Endpoints) login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	user, err := e.uc.Login.Execute(ctx, input.Body.Token)
	if err != nil {
		return nil, mapErr(err)
	}
	return &LoginOutput{
		SetCookie: sessionCookie(input.Body.Token),
		Body:      user,
	}, nil
}

func (e *Endpoints) logout(_ context.Context, _ *struct{}) (*LogoutOutput, error) {
	return &LogoutOutput{SetCookie: clearCookie()}, nil
}

func (e *Endpoints) me(ctx context.Context, _ *struct{}) (*UserOutput, error) {
	user, err := e.uc.Me.Execute(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &UserOutput{Body: user}, nil
}

func sessionCookie(token string) string {
	c := &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 365,
	}
	return c.String()
}

func clearCookie() string {
	c := &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
	return c.String()
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		return huma.NewError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, base.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, base.ErrValidation):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, fmt.Sprintf("%v", err))
	}
}
