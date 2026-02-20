package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/logs/use_cases"
	"mantis/core/base"
)

type UseCases struct {
	ListSessionLogs *usecases.ListSessionLogs
	GetSessionLog   *usecases.GetSessionLog
	ClearLogs       *usecases.ClearLogs
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "list-session-logs", Method: http.MethodGet, Path: "/api/session-logs"}, e.listSessionLogs)
	huma.Register(api, huma.Operation{OperationID: "get-session-log", Method: http.MethodGet, Path: "/api/session-logs/{id}"}, e.getSessionLog)
	huma.Register(api, huma.Operation{OperationID: "clear-session-logs", Method: http.MethodDelete, Path: "/api/session-logs", DefaultStatus: 204}, e.clearLogs)
}

func (e *Endpoints) listSessionLogs(ctx context.Context, input *ListSessionLogsInput) (*SessionLogsOutput, error) {
	items, err := e.uc.ListSessionLogs.Execute(ctx, input.ConnectionID, input.Limit, input.Offset)
	if err != nil {
		return nil, mapErr(err)
	}
	return &SessionLogsOutput{Body: items}, nil
}

func (e *Endpoints) getSessionLog(ctx context.Context, input *SessionLogIDInput) (*SessionLogOutput, error) {
	s, err := e.uc.GetSessionLog.Execute(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &SessionLogOutput{Body: s}, nil
}

func (e *Endpoints) clearLogs(ctx context.Context, _ *struct{}) (*struct{}, error) {
	if err := e.uc.ClearLogs.Execute(ctx); err != nil {
		return nil, mapErr(err)
	}
	return nil, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, base.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, err.Error())
	}
}
