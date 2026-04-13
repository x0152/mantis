package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"mantis/apps/chat"
	"mantis/apps/logs"
	"mantis/apps/metadata"
	plansapp "mantis/apps/plans"
	"mantis/apps/telegram"
	"mantis/core/agents"
	artifactplugin "mantis/core/plugins/artifact"
	"mantis/core/plugins/guard"
	"mantis/core/plugins/memory"
	"mantis/core/protocols"
	"mantis/core/types"
	artifactadapter "mantis/infrastructure/adapters/artifact"
	"mantis/infrastructure/adapters/asr"
	"mantis/infrastructure/adapters/llm"
	"mantis/infrastructure/adapters/ocr"
	"mantis/infrastructure/adapters/store"
	"mantis/infrastructure/adapters/tts"
	"mantis/infrastructure/mappers"
	"mantis/infrastructure/models"
	"mantis/shared"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mantis?sslmode=disable")
	port := env("PORT", "8080")

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	settingsStore := store.NewPostgres[string, types.Settings, models.SettingsRow](
		db,
		func(s types.Settings) string { return s.ID },
		mappers.SettingsToRow,
		mappers.SettingsFromRow,
	)
	llmConnStore := store.NewPostgres[string, types.LlmConnection, models.LlmConnectionRow](
		db,
		func(c types.LlmConnection) string { return c.ID },
		mappers.LlmConnectionToRow,
		mappers.LlmConnectionFromRow,
	)
	modelStore := store.NewPostgres[string, types.Model, models.ModelRow](
		db,
		func(m types.Model) string { return m.ID },
		mappers.ModelToRow,
		mappers.ModelFromRow,
	)
	presetStore := store.NewPostgres[string, types.Preset, models.PresetRow](
		db,
		func(p types.Preset) string { return p.ID },
		mappers.PresetToRow,
		mappers.PresetFromRow,
	)
	connectionStore := store.NewPostgres[string, types.Connection, models.ConnectionRow](
		db,
		func(c types.Connection) string { return c.ID },
		mappers.ConnectionToRow,
		mappers.ConnectionFromRow,
	)
	skillStore := store.NewPostgres[string, types.Skill, models.SkillRow](
		db,
		func(s types.Skill) string { return s.ID },
		mappers.SkillToRow,
		mappers.SkillFromRow,
	)
	planStore := store.NewPostgres[string, types.Plan, models.PlanRow](
		db,
		func(p types.Plan) string { return p.ID },
		mappers.PlanToRow,
		mappers.PlanFromRow,
	)
	planRunStore := store.NewPostgres[string, types.PlanRun, models.PlanRunRow](
		db,
		func(r types.PlanRun) string { return r.ID },
		mappers.PlanRunToRow,
		mappers.PlanRunFromRow,
	)
	sessionStore := store.NewPostgres[string, types.ChatSession, models.ChatSessionRow](
		db,
		func(s types.ChatSession) string { return s.ID },
		mappers.ChatSessionToRow,
		mappers.ChatSessionFromRow,
	)
	messageStore := store.NewPostgres[string, types.ChatMessage, models.ChatMessageRow](
		db,
		func(m types.ChatMessage) string { return m.ID },
		mappers.ChatMessageToRow,
		mappers.ChatMessageFromRow,
	)

	logStore := store.NewPostgres[string, types.SessionLog, models.SessionLogRow](
		db,
		func(s types.SessionLog) string { return s.ID },
		mappers.SessionLogToRow,
		mappers.SessionLogFromRow,
	)
	guardProfileStore := store.NewPostgres[string, types.GuardProfile, models.GuardProfileRow](
		db,
		func(p types.GuardProfile) string { return p.ID },
		mappers.GuardProfileToRow,
		mappers.GuardProfileFromRow,
	)
	channelStore := store.NewPostgres[string, types.Channel, models.ChannelRow](
		db,
		func(c types.Channel) string { return c.ID },
		mappers.ChannelToRow,
		mappers.ChannelFromRow,
	)

	openaiAdapter := llm.NewOpenAI()
	sessionLogger := shared.NewSessionLogger(logStore)
	commandGuard := guard.New(guardProfileStore)

	var asrAdapter protocols.ASR
	if u := env("ASR_API_URL", ""); u != "" {
		asrAdapter = asr.NewGigaAM(u, 5*time.Minute)
	}
	var ocrAdapter protocols.OCR
	if u := env("OCR_API_URL", ""); u != "" {
		ocrAdapter = ocr.NewAPI(u, 5*time.Minute)
	}
	var ttsAdapter protocols.TTS
	if u := env("TTS_API_URL", ""); u != "" {
		ttsAdapter = tts.NewCosyVoice(u, 5*time.Minute)
	}

	visionAdapter := llm.NewVision()
	mantisAgent := agents.NewMantisAgent(messageStore, modelStore, presetStore, llmConnStore, connectionStore, skillStore, planStore, channelStore, settingsStore, openaiAdapter, commandGuard, sessionLogger, asrAdapter, ocrAdapter, visionAdapter)

	buf := shared.NewBuffer()
	artifactMgr := artifactplugin.NewManager(artifactadapter.NewInMemorySessionStorage())
	memoryExtractor := memory.NewExtractor(openaiAdapter, settingsStore, connectionStore, modelStore, presetStore, llmConnStore)

	attachmentDir := env("ATTACHMENT_DIR", "/data/attachments")

	plansApp := plansapp.NewApp(settingsStore, sessionStore, messageStore, modelStore, presetStore, planStore, planRunStore, mantisAgent, artifactMgr, memoryExtractor, buf)
	mantisAgent.SetPlanRunner(plansApp.Runner())

	metadataApp := metadata.NewApp(settingsStore, llmConnStore, modelStore, presetStore, connectionStore, skillStore, planStore, planRunStore, plansApp.Runner(), guardProfileStore, channelStore)
	chatApp := chat.NewApp(sessionStore, messageStore, modelStore, presetStore, channelStore, settingsStore, mantisAgent, buf, artifactMgr, memoryExtractor)
	logsApp := logs.NewApp(logStore)
	telegramApp := telegram.NewApp(channelStore, sessionStore, messageStore, modelStore, presetStore, settingsStore, mantisAgent, buf, artifactMgr, asrAdapter, ttsAdapter, memoryExtractor)

	chatApp.SetAttachmentDir(attachmentDir)
	plansApp.SetAttachmentDir(attachmentDir)

	go telegramApp.Start(context.Background())
	go plansApp.Start(context.Background())

	r := chi.NewMux()

	r.Get("/api/artifacts/{sessionId}/{artifactId}", func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionId")
		artifactID := chi.URLParam(r, "artifactId")

		store := artifactMgr.ForSession(sessionID)
		if a, ok := store.Get(artifactID); ok {
			serveBinary(w, a.Bytes, a.MIME, a.Name)
			return
		}

		dataPath := filepath.Join(attachmentDir, artifactID)
		data, err := os.ReadFile(dataPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		var meta struct {
			MIME string `json:"mime"`
			Name string `json:"name"`
		}
		if raw, err := os.ReadFile(dataPath + ".json"); err == nil {
			_ = json.Unmarshal(raw, &meta)
		}
		serveBinary(w, data, meta.MIME, meta.Name)
	})

	api := humachi.New(r, huma.DefaultConfig("Mantis API", "1.0.0"))
	metadataApp.Register(api)
	chatApp.Register(api)
	logsApp.Register(api)

	log.Printf("listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func serveBinary(w http.ResponseWriter, data []byte, mime, name string) {
	if mime == "" {
		mime = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=1800")
	if name != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", name))
	}
	w.Write(data)
}
