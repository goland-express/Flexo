package main

import (
	"context"
	"flexo/commands"
	"flexo/player"
	"flexo/registry"
	"flexo/types"
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
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	token := os.Getenv("DISCORD_TOKEN")
	lavalinkHost := os.Getenv("LAVALINK_HOST")
	lavalinkPassword := os.Getenv("LAVALINK_PASSWORD")

	if token == "" {
		slog.Error("DISCORD_TOKEN is not set in environment variabls.")
		return
	}

	botData := &types.BotData{
		StartTime: time.Now(),
	}

	reg := registry.New(registry.Options{
		Data:   botData,
		Prefix: "!",
		OnError: func(err error, ctx *registry.Context) {
			slog.Error("Error executing command",
				slog.Any("error", err),
				slog.String("author", ctx.Author().Username),
			)
			_ = ctx.Say("https://media.discordapp.net/attachments/692443311318892585/1432174798192115763/UuYcrRF.png?ex=69001838&is=68fec6b8&hm=fe6504fca5a9d2d065217b4051a5ce669aa668602df75a413245487fc71ce19b&=&format=webp&quality=lossless&width=864&height=490")
		},
	})

	commands.Register(reg)

	client, err := disgo.New(token,
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
				slog.String("prefix", "!"),
				slog.Time("started_at", time.Now()),
				slog.String("go_version", runtime.Version()),
				slog.String("disgo_version", disgo.Version),
				slog.String("os", runtime.GOOS),
				slog.String("arch", runtime.GOARCH),
		)

			go func() {
				pm, err := player.New(event.User.ID, lavalinkHost, lavalinkPassword)
				if err != nil {
					slog.Error("Failed to initialize player",
						slog.Any("error", err),
					)
					slog.Warn("Bot will continue without music features")
					return
				}

				botData.Player = pm
			}()
		}),
	)
	if err != nil {
		slog.Error("Failed to create Disgo client", slog.Any("err", err))
		return
	}

	client.EventManager().AddEventListeners(&events.ListenerAdapter{
		OnGuildVoiceStateUpdate: func(event *events.GuildVoiceStateUpdate) {
			if botData.Player != nil {
				botData.Player.OnVoiceStateUpdate(event)
			}
		},
		OnVoiceServerUpdate: func(event *events.VoiceServerUpdate) {
			if botData.Player != nil {
				botData.Player.OnVoiceServerUpdate(event)
			}
		},
	})

	defer client.Close(context.TODO())

	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("Failed to connect to gateway", slog.Any("err", err))
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
