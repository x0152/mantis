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
