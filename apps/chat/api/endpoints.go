package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/chat/use_cases"
)

type UseCases struct {
	GetCurrentSession *usecases.GetCurrentSession
	ResetContext      *usecases.ResetContext
	ListSessions      *usecases.ListSessions
	CreateSession     *usecases.CreateSession
	UpdateSession     *usecases.UpdateSession
	DeleteSession     *usecases.DeleteSession
	ListMessages      *usecases.ListMessages
	SendMessage       *usecases.SendMessage
	ClearHistory      *usecases.ClearHistory
}

type Endpoints struct {
	uc UseCases
}

func NewEndpoints(uc UseCases) *Endpoints {
	return &Endpoints{uc: uc}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{OperationID: "get-chat-session", Method: http.MethodGet, Path: "/api/chat/session"}, e.getSession)
	huma.Register(api, huma.Operation{OperationID: "reset-chat-context", Method: http.MethodPost, Path: "/api/chat/reset", DefaultStatus: 201}, e.resetContext)
	huma.Register(api, huma.Operation{OperationID: "list-chat-sessions", Method: http.MethodGet, Path: "/api/chat/sessions"}, e.listSessions)
	huma.Register(api, huma.Operation{OperationID: "create-chat-session", Method: http.MethodPost, Path: "/api/chat/sessions", DefaultStatus: 201}, e.createSession)
	huma.Register(api, huma.Operation{OperationID: "update-chat-session", Method: http.MethodPut, Path: "/api/chat/sessions/{id}"}, e.updateSession)
	huma.Register(api, huma.Operation{OperationID: "delete-chat-session", Method: http.MethodDelete, Path: "/api/chat/sessions/{id}", DefaultStatus: 204}, e.deleteSession)
	huma.Register(api, huma.Operation{OperationID: "list-chat-messages", Method: http.MethodGet, Path: "/api/chat/messages"}, e.listMessages)
	huma.Register(api, huma.Operation{OperationID: "send-chat-message", Method: http.MethodPost, Path: "/api/chat/messages", DefaultStatus: 201}, e.sendMessage)
	huma.Register(api, huma.Operation{OperationID: "clear-chat-history", Method: http.MethodDelete, Path: "/api/chat/history", DefaultStatus: 204}, e.clearHistory)
}

func (e *Endpoints) getSession(ctx context.Context, _ *struct{}) (*SessionOutput, error) {
	s, err := e.uc.GetCurrentSession.Execute(ctx)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SessionOutput{Body: s}, nil
}

func (e *Endpoints) resetContext(ctx context.Context, _ *struct{}) (*SessionOutput, error) {
	s, err := e.uc.ResetContext.Execute(ctx)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SessionOutput{Body: s}, nil
}

func (e *Endpoints) listSessions(ctx context.Context, input *ListSessionsInput) (*SessionsOutput, error) {
	sessions, err := e.uc.ListSessions.Execute(ctx, input.Limit, input.Offset)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SessionsOutput{Body: sessions}, nil
}

func (e *Endpoints) createSession(ctx context.Context, input *CreateSessionInput) (*SessionOutput, error) {
	s, err := e.uc.CreateSession.Execute(ctx, input.Body.Title)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SessionOutput{Body: s}, nil
}

func (e *Endpoints) updateSession(ctx context.Context, input *UpdateSessionInput) (*SessionOutput, error) {
	s, err := e.uc.UpdateSession.Execute(ctx, input.ID, input.Body.Title)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SessionOutput{Body: s}, nil
}

func (e *Endpoints) deleteSession(ctx context.Context, input *DeleteSessionInput) (*struct{}, error) {
	if err := e.uc.DeleteSession.Execute(ctx, input.ID); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil, nil
}

func (e *Endpoints) listMessages(ctx context.Context, input *ListMessagesInput) (*MessagesOutput, error) {
	msgs, err := e.uc.ListMessages.Execute(ctx, input.Limit, input.Offset, input.SessionID, input.Source)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &MessagesOutput{Body: msgs}, nil
}

func (e *Endpoints) sendMessage(ctx context.Context, input *SendMessageInput) (*SendMessageOutput, error) {
	user, assistant, err := e.uc.SendMessage.Execute(ctx, input.Body.SessionID, input.Body.Content)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SendMessageOutput{Body: SendMessageResponse{
		UserMessage: user, AssistantMessage: assistant,
	}}, nil
}

func (e *Endpoints) clearHistory(ctx context.Context, _ *struct{}) (*struct{}, error) {
	if err := e.uc.ClearHistory.Execute(ctx); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil, nil
}
