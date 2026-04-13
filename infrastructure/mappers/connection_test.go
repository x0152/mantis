package mappers

import (
	"encoding/json"
	"testing"
	"time"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func TestConnectionToRow(t *testing.T) {
	c := types.Connection{
		ID:          "c1",
		Type:        "ssh",
		Name:        "Server",
		Description: "Main server",
		ModelID:     "m1",
		PresetID:    "pr1",
		Config:      json.RawMessage(`{"host":"10.0.0.1"}`),
		Memories: []types.Memory{
			{ID: "mem1", Content: "something", CreatedAt: time.Now()},
		},
		ProfileIDs:    []string{"prof1", "prof2"},
		MemoryEnabled: true,
	}
	row := ConnectionToRow(c)
	if row.ID != "c1" {
		t.Fatalf("ID: %s", row.ID)
	}
	if row.Type != "ssh" {
		t.Fatalf("Type: %s", row.Type)
	}
	if row.PresetID != "pr1" {
		t.Fatalf("PresetID: %s", row.PresetID)
	}
	if !row.MemoryEnabled {
		t.Fatal("MemoryEnabled should be true")
	}

	var memories []types.Memory
	if err := json.Unmarshal(row.Memories, &memories); err != nil {
		t.Fatalf("Memories unmarshal: %v", err)
	}
	if len(memories) != 1 || memories[0].ID != "mem1" {
		t.Fatal("memories mismatch")
	}

	var profileIDs []string
	if err := json.Unmarshal(row.ProfileIDs, &profileIDs); err != nil {
		t.Fatalf("ProfileIDs unmarshal: %v", err)
	}
	if len(profileIDs) != 2 || profileIDs[0] != "prof1" {
		t.Fatal("profileIDs mismatch")
	}
}

func TestConnectionFromRow(t *testing.T) {
	memories, _ := json.Marshal([]types.Memory{
		{ID: "m1", Content: "fact"},
	})
	profileIDs, _ := json.Marshal([]string{"p1"})
	row := models.ConnectionRow{
		ID:            "c1",
		Type:          "ssh",
		Name:          "Test",
		Config:        json.RawMessage(`{"port":22}`),
		Memories:      memories,
		ProfileIDs:    profileIDs,
		MemoryEnabled: true,
	}
	c := ConnectionFromRow(row)
	if c.ID != "c1" || c.Type != "ssh" {
		t.Fatal("basic fields")
	}
	if len(c.Memories) != 1 || c.Memories[0].ID != "m1" {
		t.Fatal("memories")
	}
	if len(c.ProfileIDs) != 1 || c.ProfileIDs[0] != "p1" {
		t.Fatal("profileIDs")
	}
}

func TestConnectionFromRow_NilArrays(t *testing.T) {
	row := models.ConnectionRow{ID: "c1"}
	c := ConnectionFromRow(row)
	if c.Memories == nil {
		t.Fatal("Memories should be empty slice, not nil")
	}
	if c.ProfileIDs == nil {
		t.Fatal("ProfileIDs should be empty slice, not nil")
	}
}

func TestConnectionRoundTrip(t *testing.T) {
	original := types.Connection{
		ID:          "c-rt",
		Type:        "ssh",
		Name:        "Roundtrip Server",
		Description: "For testing",
		ModelID:     "model-1",
		PresetID:    "preset-1",
		Config:      json.RawMessage(`{"host":"example.com","port":22}`),
		Memories: []types.Memory{
			{ID: "m1", Content: "remembers this"},
		},
		ProfileIDs:    []string{"guard1"},
		MemoryEnabled: true,
	}
	restored := ConnectionFromRow(ConnectionToRow(original))
	if restored.ID != original.ID {
		t.Fatal("ID")
	}
	if restored.Type != original.Type {
		t.Fatal("Type")
	}
	if restored.Name != original.Name {
		t.Fatal("Name")
	}
	if restored.PresetID != original.PresetID {
		t.Fatal("PresetID")
	}
	if restored.ModelID != original.ModelID {
		t.Fatal("ModelID")
	}
	if !restored.MemoryEnabled {
		t.Fatal("MemoryEnabled")
	}
	if len(restored.Memories) != 1 || restored.Memories[0].Content != "remembers this" {
		t.Fatal("Memories")
	}
	if len(restored.ProfileIDs) != 1 || restored.ProfileIDs[0] != "guard1" {
		t.Fatal("ProfileIDs")
	}
	if string(restored.Config) != string(original.Config) {
		t.Fatalf("Config: %s vs %s", string(restored.Config), string(original.Config))
	}
}
