package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"mantis/core/protocols"
	"mantis/core/types"
	adapter "mantis/infrastructure/adapters/channel"
)

type HandleModelCommand struct {
	modelStore   protocols.Store[string, types.Model]
	channelStore protocols.Store[string, types.Channel]
}

func NewHandleModelCommand(
	modelStore protocols.Store[string, types.Model],
	channelStore protocols.Store[string, types.Channel],
) *HandleModelCommand {
	return &HandleModelCommand{
		modelStore:   modelStore,
		channelStore: channelStore,
	}
}

func (uc *HandleModelCommand) Execute(ctx context.Context, channelID string, args string) (adapter.Reply, error) {
	if uc.modelStore == nil {
		return adapter.Reply{Text: "Models unavailable: model store is not configured."}, nil
	}
	if uc.channelStore == nil {
		return adapter.Reply{Text: "Channel unavailable: channel store is not configured."}, nil
	}

	arg := strings.TrimSpace(args)
	if arg == "" || arg == "list" {
		currentID, err := uc.getChannelModelID(ctx, channelID)
		if err != nil {
			return adapter.Reply{}, err
		}
		return uc.modelListReply(ctx, currentID), nil
	}

	newID := strings.Fields(arg)[0]
	existing, err := uc.modelStore.Get(ctx, []string{newID})
	if err != nil {
		return adapter.Reply{}, err
	}
	m, ok := existing[newID]
	if !ok {
		return adapter.Reply{Text: "Model not found: " + newID}, nil
	}

	if err := uc.updateChannelModelID(ctx, channelID, newID); err != nil {
		return adapter.Reply{}, err
	}
	return adapter.Reply{Text: fmt.Sprintf("Model switched: %s (%s)", m.Name, m.ID)}, nil
}

func (uc *HandleModelCommand) modelListReply(ctx context.Context, currentID string) adapter.Reply {
	items, err := uc.modelStore.List(ctx, types.ListQuery{})
	if err != nil {
		return adapter.Reply{Text: fmt.Sprintf("Error: %v", err)}
	}
	if items == nil {
		items = []types.Model{}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

	currentName := ""
	if currentID != "" {
		if m, ok := mapByID(items)[currentID]; ok {
			currentName = m.Name
		}
	}

	var sb strings.Builder
	if currentID == "" {
		sb.WriteString("Select a model.\n")
	} else if currentName != "" {
		sb.WriteString(fmt.Sprintf("Current model: %s\n", currentName))
	} else {
		sb.WriteString(fmt.Sprintf("Current model: %s\n", currentID))
	}
	if len(items) == 0 {
		sb.WriteString("\nNo models available. Create one in the web panel.")
		return adapter.Reply{Text: strings.TrimSpace(sb.String())}
	}
	sb.WriteString("\nTap a button to switch model.")

	type btn map[string]string
	var keyboard [][]btn
	row := []btn{}
	for _, m := range items {
		label := m.Name
		if m.ID == currentID {
			label = "✅ " + label
		}
		row = append(row, btn{
			"text":          label,
			"callback_data": "model:" + m.ID,
		})
		if len(row) == 2 {
			keyboard = append(keyboard, row)
			row = nil
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	markup, _ := json.Marshal(map[string]any{
		"inline_keyboard": keyboard,
	})

	return adapter.Reply{
		Text:        strings.TrimSpace(sb.String()),
		ReplyMarkup: markup,
	}
}

func mapByID(items []types.Model) map[string]types.Model {
	out := make(map[string]types.Model, len(items))
	for _, m := range items {
		out[m.ID] = m
	}
	return out
}

func (uc *HandleModelCommand) getChannelModelID(ctx context.Context, channelID string) (string, error) {
	existing, err := uc.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return "", err
	}
	ch, ok := existing[channelID]
	if !ok {
		return "", fmt.Errorf("channel %q not found", channelID)
	}
	return strings.TrimSpace(ch.ModelID), nil
}

func (uc *HandleModelCommand) updateChannelModelID(ctx context.Context, channelID, modelID string) error {
	existing, err := uc.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return err
	}
	ch, ok := existing[channelID]
	if !ok {
		return fmt.Errorf("channel %q not found", channelID)
	}
	ch.ModelID = modelID
	_, err = uc.channelStore.Update(ctx, []types.Channel{ch})
	return err
}
