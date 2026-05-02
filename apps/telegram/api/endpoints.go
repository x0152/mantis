package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/telegram/use_cases"
	"mantis/core/base"
)

type UseCases struct {
	Wizard *usecases.Wizard
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "telegram-wizard-verify", Method: http.MethodPost, Path: "/api/telegram/wizard/verify"}, e.verify)
	huma.Register(api, huma.Operation{OperationID: "telegram-wizard-status", Method: http.MethodPost, Path: "/api/telegram/wizard/status"}, e.status)
}

func (e *Endpoints) verify(ctx context.Context, input *WizardVerifyInput) (*WizardVerifyOutput, error) {
	bot, err := e.uc.Wizard.Verify(ctx, input.Body.Token)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &WizardVerifyOutput{}
	out.Body = *bot
	return out, nil
}

func (e *Endpoints) status(ctx context.Context, input *WizardStatusInput) (*WizardStatusOutput, error) {
	user, err := e.uc.Wizard.Status(ctx, input.Body.Token)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &WizardStatusOutput{}
	out.Body.User = user
	return out, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, base.ErrValidation):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, err.Error())
	}
}
