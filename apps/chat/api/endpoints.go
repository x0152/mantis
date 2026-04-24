package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	usecases "mantis/apps/chat/use_cases"
	"mantis/core/protocols"
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
	StopGeneration    *usecases.StopGeneration
	RegenerateLast    *usecases.RegenerateLast
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
	huma.Register(api, huma.Operation{OperationID: "send-chat-message", Method: http.MethodPost, Path: "/api/chat/messages", DefaultStatus: 201, MaxBodyBytes: 64 * 1024 * 1024}, e.sendMessage)
	huma.Register(api, huma.Operation{OperationID: "regenerate-chat-last", Method: http.MethodPost, Path: "/api/chat/sessions/{id}/regenerate", DefaultStatus: 201}, e.regenerate)
	huma.Register(api, huma.Operation{OperationID: "stop-chat-session", Method: http.MethodPost, Path: "/api/chat/sessions/{id}/stop"}, e.stopSession)
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
	content := strings.TrimSpace(input.Body.Content)
	if content == "" && len(input.Body.Files) == 0 {
		return nil, huma.NewError(http.StatusBadRequest, "content or files are required")
	}

	files := make([]protocols.FileAttachment, 0, len(input.Body.Files))
	for _, f := range input.Body.Files {
		data, err := base64.StdEncoding.DecodeString(f.DataBase64)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "invalid base64 for file "+f.FileName)
		}
		if len(data) == 0 {
			continue
		}
		files = append(files, protocols.FileAttachment{
			FileName: f.FileName,
			MimeType: f.MimeType,
			Data:     data,
			Caption:  f.Caption,
		})
	}

	user, assistant, err := e.uc.SendMessage.Execute(ctx, input.Body.SessionID, input.Body.Content, files)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &SendMessageOutput{Body: SendMessageResponse{
		UserMessage: user, AssistantMessage: assistant,
	}}, nil
}

func (e *Endpoints) regenerate(ctx context.Context, input *RegenerateInput) (*RegenerateOutput, error) {
	if e.uc.RegenerateLast == nil {
		return nil, huma.NewError(http.StatusNotImplemented, "regeneration is not configured")
	}
	assistant, err := e.uc.RegenerateLast.Execute(ctx, input.ID)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &RegenerateOutput{Body: RegenerateResponse{AssistantMessage: assistant}}, nil
}

func (e *Endpoints) stopSession(ctx context.Context, input *StopSessionInput) (*StopSessionOutput, error) {
	stopped, err := e.uc.StopGeneration.Execute(ctx, input.ID)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return &StopSessionOutput{Body: StopSessionResponse{Stopped: stopped}}, nil
}

func (e *Endpoints) clearHistory(ctx context.Context, _ *struct{}) (*struct{}, error) {
	if err := e.uc.ClearHistory.Execute(ctx); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil, nil
}
