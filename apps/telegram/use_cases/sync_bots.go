package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"mantis/core/protocols"
	"mantis/core/types"
	adapter "mantis/infrastructure/adapters/channel"
)

type botState struct {
	key    string
	cancel context.CancelFunc
}

type SyncBots struct {
	channelStore protocols.Store[string, types.Channel]
	makeHandler  func(channelID string) adapter.MessageHandler

	mu   sync.Mutex
	bots map[string]botState
}

func NewSyncBots(
	channelStore protocols.Store[string, types.Channel],
	makeHandler func(channelID string) adapter.MessageHandler,
) *SyncBots {
	return &SyncBots{
		channelStore: channelStore,
		makeHandler:  makeHandler,
		bots:         make(map[string]botState),
	}
}

func (uc *SyncBots) Execute(ctx context.Context) {
	if uc.channelStore == nil {
		return
	}
	channels, err := uc.channelStore.List(ctx, types.ListQuery{})
	if err != nil {
		log.Printf("telegram: list channels: %v", err)
		return
	}
	desired := make(map[string]types.Channel)
	for _, ch := range channels {
		if ch.Type == "telegram" && strings.TrimSpace(ch.Token) != "" {
			desired[ch.ID] = ch
		}
	}

	uc.mu.Lock()
	for id, st := range uc.bots {
		ch, ok := desired[id]
		if !ok || st.key != botKey(ch) {
			st.cancel()
			delete(uc.bots, id)
		}
	}

	for id, ch := range desired {
		if _, ok := uc.bots[id]; ok {
			continue
		}
		key := botKey(ch)
		bctx, cancel := context.WithCancel(ctx)
		uc.bots[id] = botState{key: key, cancel: cancel}
		bot := adapter.NewTelegram(ch.Token, ch.AllowedUserIDs, uc.makeHandler(ch.ID))
		go uc.runBot(bctx, id, key, bot)
	}
	uc.mu.Unlock()
}

func (uc *SyncBots) Stop() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	for id, st := range uc.bots {
		st.cancel()
		delete(uc.bots, id)
	}
}

func (uc *SyncBots) runBot(ctx context.Context, id, key string, bot protocols.Channel) {
	log.Printf("telegram: starting bot id=%s", id)
	if err := bot.Execute(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("telegram: bot stopped id=%s: %v", id, err)
	}

	uc.mu.Lock()
	if st, ok := uc.bots[id]; ok && st.key == key {
		delete(uc.bots, id)
	}
	uc.mu.Unlock()
}

func botKey(ch types.Channel) string {
	return ch.Token + "|" + fmt.Sprint(ch.AllowedUserIDs)
}
