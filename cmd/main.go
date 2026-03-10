package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"mantis/apps/chat"
	"mantis/apps/cron"
	"mantis/apps/logs"
	"mantis/apps/metadata"
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

	configStore := store.NewPostgres[string, types.Config, models.ConfigRow](
		db,
		func(c types.Config) string { return c.ID },
		mappers.ConfigToRow,
		mappers.ConfigFromRow,
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
	connectionStore := store.NewPostgres[string, types.Connection, models.ConnectionRow](
		db,
		func(c types.Connection) string { return c.ID },
		mappers.ConnectionToRow,
		mappers.ConnectionFromRow,
	)
	cronJobStore := store.NewPostgres[string, types.CronJob, models.CronJobRow](
		db,
		func(j types.CronJob) string { return j.ID },
		mappers.CronJobToRow,
		mappers.CronJobFromRow,
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
	mantisAgent := agents.NewMantisAgent(messageStore, modelStore, llmConnStore, connectionStore, cronJobStore, configStore, openaiAdapter, commandGuard, sessionLogger, asrAdapter, ocrAdapter, visionAdapter)

	buf := shared.NewBuffer()
	artifactMgr := artifactplugin.NewManager(artifactadapter.NewInMemorySessionStorage())
	memoryExtractor := memory.NewExtractor(openaiAdapter, configStore, connectionStore, modelStore, llmConnStore)

	metadataApp := metadata.NewApp(configStore, llmConnStore, modelStore, connectionStore, cronJobStore, guardProfileStore, channelStore)
	chatApp := chat.NewApp(sessionStore, messageStore, modelStore, channelStore, configStore, mantisAgent, buf, artifactMgr, memoryExtractor)
	logsApp := logs.NewApp(logStore)
	telegramApp := telegram.NewApp(channelStore, sessionStore, messageStore, modelStore, mantisAgent, buf, artifactMgr, asrAdapter, ttsAdapter, memoryExtractor)
	cronApp := cron.NewApp(configStore, channelStore, sessionStore, messageStore, modelStore, cronJobStore, mantisAgent, artifactMgr, memoryExtractor)

	go telegramApp.Start(context.Background())
	go cronApp.Start(context.Background())

	r := chi.NewMux()
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
