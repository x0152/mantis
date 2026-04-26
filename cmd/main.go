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
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	authapp "mantis/apps/auth"
	"mantis/apps/chat"
	"mantis/apps/logs"
	"mantis/apps/metadata"
	plansapp "mantis/apps/plans"
	runtimeapp "mantis/apps/runtime"
	"mantis/apps/telegram"
	"mantis/core/agents"
	"mantis/core/auth"
	artifactplugin "mantis/core/plugins/artifact"
	"mantis/core/plugins/guard"
	"mantis/core/plugins/memory"
	"mantis/core/plugins/pipeline"
	"mantis/core/plugins/summarizer"
	"mantis/core/protocols"
	"mantis/core/types"
	artifactadapter "mantis/infrastructure/adapters/artifact"
	"mantis/infrastructure/adapters/asr"
	"mantis/infrastructure/adapters/llm"
	"mantis/infrastructure/adapters/ocr"
	dockerruntime "mantis/infrastructure/adapters/runtime/docker"
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
	userStore := store.NewPostgres[string, types.User, models.UserRow](
		db,
		func(u types.User) string { return u.ID },
		mappers.UserToRow,
		mappers.UserFromRow,
	)

	openaiAdapter := llm.NewOpenAI()
	gonkaAdapter := llm.NewGonka()
	llmAdapter := llm.NewRouter("openai", map[string]protocols.LLM{
		"openai": openaiAdapter,
		"gonka":  gonkaAdapter,
	})
	llmCatalogs := map[string]protocols.LLMCatalog{
		"openai": openaiAdapter,
		"gonka":  gonkaAdapter,
	}
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
	limits := shared.LoadLimits()
	log.Printf("limits: supervisor=%s/%d, server=%s/%d, plan_step=%s",
		shared.FormatDuration(limits.SupervisorTimeout), limits.SupervisorMaxIterations,
		shared.FormatDuration(limits.ServerTimeout), limits.ServerMaxIterations,
		shared.FormatDuration(limits.PlanStepTimeout))
	mantisAgent := agents.NewMantisAgent(messageStore, modelStore, presetStore, llmConnStore, connectionStore, skillStore, planStore, channelStore, settingsStore, sessionStore, llmAdapter, commandGuard, sessionLogger, asrAdapter, ocrAdapter, visionAdapter, limits)

	buf := shared.NewBuffer()
	artifactMgr := artifactplugin.NewManager(artifactadapter.NewInMemorySessionStorage())
	memoryExtractor := memory.NewExtractor(llmAdapter, settingsStore, connectionStore, modelStore, presetStore, llmConnStore)
	summ := summarizer.New(llmAdapter, sessionStore, messageStore, modelStore, presetStore, llmConnStore, buf)
	summ.SetMemoryFlusher(memoryExtractor)

	attachmentDir := env("ATTACHMENT_DIR", "/data/attachments")

	cancellations := pipeline.NewCancellations()

	plansApp := plansapp.NewApp(settingsStore, sessionStore, messageStore, modelStore, presetStore, planStore, planRunStore, mantisAgent, artifactMgr, memoryExtractor, summ, buf)
	mantisAgent.SetPlanRunner(plansApp.Runner())

	metadataApp := metadata.NewApp(settingsStore, llmConnStore, modelStore, presetStore, connectionStore, skillStore, planStore, planRunStore, plansApp.Runner(), guardProfileStore, channelStore, llmCatalogs)
	chatApp := chat.NewApp(sessionStore, messageStore, modelStore, presetStore, channelStore, settingsStore, mantisAgent, buf, artifactMgr, memoryExtractor, summ, cancellations, plansApp.Runner())
	logsApp := logs.NewApp(logStore)
	telegramApp := telegram.NewApp(channelStore, sessionStore, messageStore, modelStore, presetStore, settingsStore, mantisAgent, buf, artifactMgr, asrAdapter, ttsAdapter, memoryExtractor, summ, cancellations, plansApp.Runner())

	chatApp.SetAttachmentDir(attachmentDir)
	plansApp.SetAttachmentDir(attachmentDir)

	go telegramApp.Start(context.Background())
	go plansApp.Start(context.Background())

	authApp := authapp.NewApp(userStore)
	if token := env("AUTH_TOKEN", ""); token != "" {
		user, err := authApp.Bootstrap(context.Background(), env("AUTH_USER_NAME", "admin"), token)
		if err != nil {
			log.Fatalf("auth bootstrap failed: %v", err)
		}
		log.Printf("auth: ready, user %q (%s)", user.Name, user.ID)
	} else {
		log.Println("auth: AUTH_TOKEN not set, API will reject all requests until a user is created")
	}

	r := chi.NewMux()

	loginLimiter := auth.NewLoginRateLimiter(
		envInt("AUTH_RATE_LIMIT_MAX", 5),
		envDuration("AUTH_RATE_LIMIT_WINDOW", 15*time.Minute),
	)
	r.Use(loginLimiter.Middleware("/api/auth/login"))
	r.Use(auth.Middleware(userStore, isPublicPathFactory(env("RUNTIME_API_TOKEN", ""))))

	api := humachi.New(r, huma.DefaultConfig("Mantis API", "1.0.0"))

	authApp.Register(api)
	metadataApp.Register(api)
	chatApp.Register(api)
	logsApp.Register(api)

	if mode := env("RUNTIME_MODE", ""); mode == "docker" {
		rt := dockerruntime.New(env("DOCKER_SOCKET", ""), env("RUNTIME_NETWORK", ""))
		runtimeApp := runtimeapp.NewApp(rt, connectionStore, env("RUNTIME_API_TOKEN", ""))
		runtimeApp.Mount(r)
		mantisAgent.SetRuntime(rt)
		log.Printf("runtime: docker adapter ready (network=%s)", rt.Network())

		bootstrapper := runtimeapp.NewBootstrapper(rt, connectionStore)
		go func() {
			if err := bootstrapper.Run(context.Background()); err != nil {
				log.Printf("runtime bootstrap: %v", err)
			}
		}()
	}

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

	log.Printf("listening on :%s", port)
	log.Printf("docs: http://localhost:%s/docs", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func isPublicPathFactory(runtimeToken string) func(*http.Request) bool {
	return func(r *http.Request) bool {
		p := r.URL.Path
		switch p {
		case "/api/auth/login", "/api/auth/logout", "/docs", "/openapi.json", "/openapi.yaml":
			return true
		}
		if strings.HasPrefix(p, "/api/runtime/") {
			if runtimeToken == "" {
				return true
			}
			if r.Header.Get("X-Runtime-Token") != "" {
				return true
			}
		}
		return strings.HasPrefix(p, "/schemas/")
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
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
