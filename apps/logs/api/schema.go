package api

import "mantis/core/types"

type SessionLogOutput struct {
	Body types.SessionLog
}

type SessionLogsOutput struct {
	Body []types.SessionLog
}

type SessionLogIDInput struct {
	ID string `path:"id"`
}

type ListSessionLogsInput struct {
	ConnectionID string `query:"connectionId"`
	Limit        int    `query:"limit"`
	Offset       int    `query:"offset"`
}
