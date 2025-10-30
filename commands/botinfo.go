// TODO: Refactor: Move commands to separate files (e.g., info.go, music.go) for better organization.

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"

	"flexo/registry"
	"flexo/types"
)

func Register(r *registry.Registry) {
	r.Add(&registry.Command{
		Name:          "ping",
		Description:   "Responde com Pong!",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"status"},
		Execute: func(ctx *registry.Context) error {
			embed := discord.NewEmbedBuilder().
				SetTitle("Pong").
				SetDescription(fmt.Sprintf("Latência: **%dms**\nShard: **%d**",
					ctx.Client().Gateway().Latency().Milliseconds(),
					ctx.Client().Gateway().ShardID())).
				SetColor(0x00FF00).
				SetTimestamp(time.Now()).
				Build()
			return ctx.SendEmbed(embed)
		},
	})
	r.Add(&registry.Command{
		Name:          "play",
		Description:   "Toca uma música no canal de voz",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"p", "tocar"},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "query",
				Description: "Nome da música ou URL",
				Required:    true,
			},
		},
		Execute: func(ctx *registry.Context) error {
			guildID := ctx.GuildID()
			if guildID == nil {
				return nil
			}
			botData := ctx.Data().(*types.BotData)
			pm := botData.Player
			var query string
			if ctx.IsSlash() {
				var ok bool
				query, ok = ctx.GetStringOption("query")
				if !ok || query == "" {
					return ctx.Say("Não há nada a ser fazido sem uma query!")
				}
			} else {
				args := ctx.Args()
				query = strings.Join(args, " ")
				fmt.Printf("%s", query)
				if query == "" {
					return ctx.Say("Use: `!play <música>`")
				}
			}

			voiceState, ok := ctx.Client().Caches().VoiceState(*guildID, ctx.Author().ID)
			if !ok || voiceState.ChannelID == nil {
				return ctx.Say("Você precisa estar em um canal de voz")
			}

			channelID := *voiceState.ChannelID

			track, err := pm.Play(context.Background(), ctx.Client(), *guildID, channelID, query)
			if err != nil {
				return ctx.Say(fmt.Sprintf("Erro: %s", err.Error()))
			}
			queue, err := pm.GetQueue(context.Background(), *guildID)
			if err != nil {
				return ctx.Say(fmt.Sprintf("Erro ao obter fila: %s", err.Error()))
			}
			fmt.Printf("Current queue: %+v\n", queue)
			embed := discord.NewEmbedBuilder().
				SetTitle("Tocando").
				SetDescription(fmt.Sprintf("**[%s](%s)**\nPor: %s",
					track.Info.Title,
					*track.Info.URI,
					track.Info.Author)).
				SetThumbnail(*track.Info.ArtworkURL).
				SetColor(0x1DB954).
				Build()
			fmt.Printf("track: %+v\n", track)
			return ctx.SendEmbed(embed)
		},
	})
	r.Add(&registry.Command{
		Name:          "skip",
		Description:   "Pula para a próxima música da fila",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"s", "next"},
		Execute: func(ctx *registry.Context) error {
			guildID := ctx.GuildID()
			if guildID == nil {
				return nil
			}

			pm := ctx.Data().(*types.BotData).Player
			track, err := pm.NextTrack(context.Background(), *guildID)
			if err != nil {
				return ctx.Say(fmt.Sprintf("Erro: %s", err.Error()))
			}

			if track == nil {
				return pm.Stop(context.Background(), *guildID)
			}

			embed := discord.NewEmbedBuilder().
				SetTitle("Música Pulada").
				SetDescription(fmt.Sprintf("Tocando agora: **[%s](%s)**\nPor: %s",
					track.Info.Title,
					*track.Info.URI,
					track.Info.Author)).
				SetColor(0x1DB954).
				Build()

			return ctx.SendEmbed(embed)
		},
	})

}
