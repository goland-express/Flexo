package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"

	"github.com/goland-express/Flexo/config"
	"github.com/goland-express/Flexo/modules"
	"github.com/goland-express/Flexo/player"
	"github.com/goland-express/Flexo/registry"
	"github.com/goland-express/Flexo/types"
	"github.com/goland-express/Flexo/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	botData := &types.BotData{
		StartTime: time.Now(),
	}

	reg := registry.New(registry.Options{
		Data:   botData,
		Prefix: cfg.Prefix,
		OnError: func(err error, ctx *registry.Context) {
			if userErr, ok := err.(*utils.UserError); ok {
				_ = ctx.Reply(userErr.Message)
				return
			}
			slog.Error("An unexpected error occurred",
				slog.Any("error", err),
				slog.String("guild_id", ctx.GuildID().String()),
				slog.String("author_id", ctx.Author().ID.String()),
			)
			_ = ctx.Reply("An unexpected error occurred while running this command.")
		},
	})

	loadedModules := []modules.Module{
		&modules.MusicModule{},
	}

	for _, module := range loadedModules {
		module.Register(reg)
		slog.Info("Module loaded", slog.String("module", module.Name()))
	}

	client, err := disgo.New(cfg.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentGuildVoiceStates,
				gateway.IntentMessageContent,
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagVoiceStates),
		),
		bot.WithEventListenerFunc(reg.OnMessage),
		bot.WithEventListenerFunc(reg.OnSlashCommand),
		bot.WithEventListenerFunc(reg.OnReady),
		bot.WithEventListenerFunc(func(event *events.Ready) {
			slog.Info("Bot is ready",
				slog.String("username", event.User.Username),
				slog.String("user_id", event.User.ID.String()),
				slog.Int("guilds", len(event.Guilds)),
				slog.String("prefix", cfg.Prefix),
				slog.String("go_version", runtime.Version()),
				slog.String("disgo_version", disgo.Version),
			)

			go func() {
				pm, err := player.New(event.User.ID, cfg.LavalinkHost, cfg.LavalinkPassword)
				if err != nil {
					slog.Error("Failed to initialize player", slog.Any("error", err))
					slog.Warn("Bot will continue without music features")
					return
				}
				botData.Player = pm

				event.Client().EventManager().AddEventListeners(&events.ListenerAdapter{
					OnGuildVoiceStateUpdate: pm.OnVoiceStateUpdate,
					OnVoiceServerUpdate:     pm.OnVoiceServerUpdate,
				})

			}()
		}),
	)
	if err != nil {
		slog.Error("Failed to create Disgo client", slog.Any("err", err))
		return
	}

	defer client.Close(context.TODO())

	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("Failed to connect to gateway", slog.Any("err", err))
		return
	}

	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
