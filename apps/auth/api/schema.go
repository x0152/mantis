package api

import "mantis/core/types"

type UserOutput struct {
	Body types.User
}

type LoginInput struct {
	Body struct {
		Token string `json:"token" required:"true" minLength:"1"`
	}
}

type LoginOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      types.User
}

type LogoutOutput struct {
	SetCookie string `header:"Set-Cookie"`
}
