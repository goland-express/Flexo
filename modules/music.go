package modules

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"

	"flexo/registry"
	"flexo/types"
	"flexo/utils"
)

type MusicModule struct{}

func (m *MusicModule) Name() string {
	return "Music"
}

func (m *MusicModule) Register(r *registry.Registry) {
	r.Add(&registry.Command{
		Name:          "play",
		Description:   "Plays a song in the voice channel",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"p"},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "query", Description: "Song name or URL", Required: true},
		},
		Execute: m.executePlay,
	})

	r.Add(&registry.Command{
		Name:          "skip",
		Description:   "Skips to the next song in the queue",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"s", "next"},
		Execute:       m.executeSkip,
	})
}

func (m *MusicModule) executePlay(ctx *registry.Context) error {
	guildID := ctx.GuildID()
	if guildID == nil {
		return &utils.UserError{Message: "This command can only be used in a server."}
	}

	voiceState, ok := ctx.Client().Caches().VoiceState(*guildID, ctx.Author().ID)
	if !ok || voiceState.ChannelID == nil {
		return &utils.UserError{Message: "You must be in a voice channel to use this command."}
	}

	var query string
	if ctx.IsSlash() {
		query, _ = ctx.GetStringOption("query")
	} else {
		query = strings.Join(ctx.Args(), " ")
	}
	if query == "" {
		return &utils.UserError{Message: "You need to specify a song to play. Usage: `!play <song>`"}
	}

	botData := ctx.Data().(*types.BotData)
	pm := botData.Player
	if pm == nil {
		return &utils.UserError{Message: "The player is not yet initialized. Please try again in a moment."}
	}

	queue, _ := pm.GetQueue(context.Background(), *guildID)
	positionInQueue := 1
	if queue != nil {
		positionInQueue = len(queue.Tracks) + 1
	}

	track, err := pm.Play(context.Background(), ctx.Client(), *guildID, *voiceState.ChannelID, query)
	if err != nil {
		return fmt.Errorf("failed to play track: %w", err)
	}

	author := ctx.Author()
	embed := discord.NewEmbedBuilder().
		SetTitle(track.Info.Title).
		SetURL(*track.Info.URI).
		SetColor(0x2371AB).
		AddField("Duration", utils.FormatDuration(int(track.Info.Length)), true).
		AddField("Position in Queue", fmt.Sprintf("%d", positionInQueue), true).
		SetThumbnail(*track.Info.ArtworkURL).
		SetAuthor(track.Info.Author, "", "").
		SetFooter(fmt.Sprintf("Requested by %s", author.Username), author.EffectiveAvatarURL()).
		SetTimestamp(time.Now()).
		Build()

	return ctx.SendEmbed(embed)
}

func (m *MusicModule) executeSkip(ctx *registry.Context) error {
	guildID := ctx.GuildID()
	if guildID == nil {
		return &utils.UserError{Message: "This command can only be used in a server."}
	}

	botData := ctx.Data().(*types.BotData)
	pm := botData.Player
	if pm == nil {
		return &utils.UserError{Message: "The player is not available."}
	}

	track, err := pm.NextTrack(context.Background(), *guildID)
	if err != nil {
		return fmt.Errorf("failed to skip track: %w", err)
	}

	if track == nil {
		embed := discord.NewEmbedBuilder().
			SetDescription("The queue has ended.").
			SetColor(0x1DB954).
			Build()
		return ctx.SendEmbed(embed)
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Song Skipped").
		SetDescription(fmt.Sprintf("Now playing: **[%s](%s)**", track.Info.Title, *track.Info.URI)).
		SetColor(0x1DB954).
		SetThumbnail(*track.Info.ArtworkURL).
		Build()

	return ctx.SendEmbed(embed)
}
