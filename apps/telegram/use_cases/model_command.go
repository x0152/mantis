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
	presetStore  protocols.Store[string, types.Preset]
	channelStore protocols.Store[string, types.Channel]
}

func NewHandleModelCommand(
	presetStore protocols.Store[string, types.Preset],
	channelStore protocols.Store[string, types.Channel],
) *HandleModelCommand {
	return &HandleModelCommand{
		presetStore:  presetStore,
		channelStore: channelStore,
	}
}

func (uc *HandleModelCommand) Execute(ctx context.Context, channelID string, args string) (adapter.Reply, error) {
	if uc.presetStore == nil {
		return adapter.Reply{Text: "Presets unavailable."}, nil
	}
	if uc.channelStore == nil {
		return adapter.Reply{Text: "Channel unavailable."}, nil
	}

	arg := strings.TrimSpace(args)
	if arg == "" || arg == "list" {
		currentID, err := uc.getChannelPresetID(ctx, channelID)
		if err != nil {
			return adapter.Reply{}, err
		}
		return uc.presetListReply(ctx, currentID), nil
	}
	if arg == "inherit" || arg == "default" || arg == "clear" {
		if err := uc.updateChannelPresetID(ctx, channelID, ""); err != nil {
			return adapter.Reply{}, err
		}
		return adapter.Reply{Text: "Preset switched: Inherit (global default)"}, nil
	}

	newID := strings.Fields(arg)[0]
	existing, err := uc.presetStore.Get(ctx, []string{newID})
	if err != nil {
		return adapter.Reply{}, err
	}
	p, ok := existing[newID]
	if !ok {
		return adapter.Reply{Text: "Preset not found: " + newID}, nil
	}

	if err := uc.updateChannelPresetID(ctx, channelID, newID); err != nil {
		return adapter.Reply{}, err
	}
	return adapter.Reply{Text: fmt.Sprintf("Preset switched: %s", p.Name)}, nil
}

func (uc *HandleModelCommand) presetListReply(ctx context.Context, currentID string) adapter.Reply {
	items, err := uc.presetStore.List(ctx, types.ListQuery{})
	if err != nil {
		return adapter.Reply{Text: fmt.Sprintf("Error: %v", err)}
	}
	if items == nil {
		items = []types.Preset{}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

	currentName := ""
	if currentID != "" {
		for _, p := range items {
			if p.ID == currentID {
				currentName = p.Name
				break
			}
		}
	}

	var sb strings.Builder
	if currentID == "" {
		sb.WriteString("Using global default.\n")
	} else if currentName != "" {
		sb.WriteString(fmt.Sprintf("Current preset: %s\n", currentName))
	} else {
		sb.WriteString(fmt.Sprintf("Current preset: %s\n", currentID))
	}
	if len(items) == 0 {
		sb.WriteString("\nNo presets available. Create one in the web panel.")
		return adapter.Reply{Text: strings.TrimSpace(sb.String())}
	}
	sb.WriteString("\nTap a button to switch preset.")

	type btn map[string]string
	var keyboard [][]btn
	row := []btn{}

	row = append(row, btn{
		"text":          func() string { if currentID == "" { return "✅ Inherit" }; return "Inherit" }(),
		"callback_data": "model:inherit",
	})

	for _, p := range items {
		label := p.Name
		if p.ID == currentID {
			label = "✅ " + label
		}
		row = append(row, btn{
			"text":          label,
			"callback_data": "model:" + p.ID,
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

func (uc *HandleModelCommand) getChannelPresetID(ctx context.Context, channelID string) (string, error) {
	existing, err := uc.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return "", err
	}
	ch, ok := existing[channelID]
	if !ok {
		return "", fmt.Errorf("channel %q not found", channelID)
	}
	return strings.TrimSpace(ch.PresetID), nil
}

func (uc *HandleModelCommand) updateChannelPresetID(ctx context.Context, channelID, presetID string) error {
	existing, err := uc.channelStore.Get(ctx, []string{channelID})
	if err != nil {
		return err
	}
	ch, ok := existing[channelID]
	if !ok {
		return fmt.Errorf("channel %q not found", channelID)
	}
	ch.PresetID = presetID
	ch.ModelID = ""
	_, err = uc.channelStore.Update(ctx, []types.Channel{ch})
	return err
}
