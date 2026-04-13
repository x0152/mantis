package mappers

import (
	"testing"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func TestPresetToRow(t *testing.T) {
	temp := 0.7
	p := types.Preset{
		ID:              "pr1",
		Name:            "Default",
		ChatModelID:     "gpt-4",
		SummaryModelID:  "gpt-3.5",
		ImageModelID:    "dall-e",
		FallbackModelID: "claude",
		Temperature:     &temp,
		SystemPrompt:    "You are helpful.",
	}
	row := PresetToRow(p)
	if row.ID != "pr1" {
		t.Fatalf("ID: %s", row.ID)
	}
	if row.Name != "Default" {
		t.Fatalf("Name: %s", row.Name)
	}
	if row.ChatModelID != "gpt-4" {
		t.Fatalf("ChatModelID: %s", row.ChatModelID)
	}
	if row.SummaryModelID != "gpt-3.5" {
		t.Fatalf("SummaryModelID: %s", row.SummaryModelID)
	}
	if row.ImageModelID != "dall-e" {
		t.Fatalf("ImageModelID: %s", row.ImageModelID)
	}
	if row.FallbackModelID != "claude" {
		t.Fatalf("FallbackModelID: %s", row.FallbackModelID)
	}
	if row.Temperature == nil || *row.Temperature != 0.7 {
		t.Fatalf("Temperature: %v", row.Temperature)
	}
	if row.SystemPrompt != "You are helpful." {
		t.Fatalf("SystemPrompt: %s", row.SystemPrompt)
	}
}

func TestPresetFromRow(t *testing.T) {
	temp := 0.9
	row := models.PresetRow{
		ID:              "pr1",
		Name:            "Fast",
		ChatModelID:     "claude-3",
		FallbackModelID: "gpt-4",
		Temperature:     &temp,
		SystemPrompt:    "Be concise.",
	}
	p := PresetFromRow(row)
	if p.ID != "pr1" || p.Name != "Fast" {
		t.Fatal("basic field mismatch")
	}
	if p.ChatModelID != "claude-3" {
		t.Fatalf("ChatModelID: %s", p.ChatModelID)
	}
	if p.FallbackModelID != "gpt-4" {
		t.Fatalf("FallbackModelID: %s", p.FallbackModelID)
	}
	if p.Temperature == nil || *p.Temperature != 0.9 {
		t.Fatalf("Temperature: %v", p.Temperature)
	}
}

func TestPresetRoundTrip(t *testing.T) {
	temp := 0.5
	original := types.Preset{
		ID:              "pr-roundtrip",
		Name:            "Roundtrip",
		ChatModelID:     "gpt-4o",
		SummaryModelID:  "gpt-4o-mini",
		ImageModelID:    "dall-e-3",
		FallbackModelID: "claude-sonnet",
		Temperature:     &temp,
		SystemPrompt:    "System prompt here.",
	}
	restored := PresetFromRow(PresetToRow(original))
	if restored.ID != original.ID {
		t.Fatal("ID")
	}
	if restored.Name != original.Name {
		t.Fatal("Name")
	}
	if restored.ChatModelID != original.ChatModelID {
		t.Fatal("ChatModelID")
	}
	if restored.SummaryModelID != original.SummaryModelID {
		t.Fatal("SummaryModelID")
	}
	if restored.ImageModelID != original.ImageModelID {
		t.Fatal("ImageModelID")
	}
	if restored.FallbackModelID != original.FallbackModelID {
		t.Fatal("FallbackModelID")
	}
	if *restored.Temperature != *original.Temperature {
		t.Fatal("Temperature")
	}
	if restored.SystemPrompt != original.SystemPrompt {
		t.Fatal("SystemPrompt")
	}
}

func TestPresetRoundTrip_NilTemperature(t *testing.T) {
	original := types.Preset{ID: "pr1", Name: "No Temp"}
	restored := PresetFromRow(PresetToRow(original))
	if restored.Temperature != nil {
		t.Fatal("expected nil temperature")
	}
}
