package protocols

import (
	"context"

	"mantis/core/types"
)

type FileAttachment struct {
	FileName string
	MimeType string
	Data     []byte
	Caption  string
}

type DeliveryRequest struct {
	Text  string
	Steps []types.Step
	Files []FileAttachment
}

type ResponseTo interface {
	Execute(ctx context.Context, req DeliveryRequest) error
	Recipient() string
	Channel() string
}
