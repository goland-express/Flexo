package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"

	"github.com/goland-express/Flexo/registry"
	"github.com/goland-express/Flexo/types"
	"github.com/goland-express/Flexo/utils"
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

	r.Add(&registry.Command{
		Name:          "queue",
		Description:   "Displays the current song queue.",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"q"},
		Execute:       m.executeQueue,
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
		return &utils.UserError{Message: "The music player is not yet initialized. Please try again in a moment."}
	}

	userData := map[string]any{
		"requesterId":   ctx.Author().ID.String(),
		"requesterName": ctx.Author().Username,
	}

	track, position, err := pm.Play(context.Background(), ctx.Client(), *guildID, *voiceState.ChannelID, query, userData)
	if err != nil {
		return fmt.Errorf("failed to play track: %w", err)
	}

	author := ctx.Author()
	embed := discord.NewEmbedBuilder().
		SetTitle(track.Info.Title).
		SetURL(*track.Info.URI).
		SetColor(0x2371AB).
		AddField("Duration", utils.FormatDuration(int(track.Info.Length)), true).
		AddField("Position in Queue", fmt.Sprintf("%d", position), true).
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

func (m *MusicModule) executeQueue(ctx *registry.Context) error {
	guildID := ctx.GuildID()
	if guildID == nil {
		return &utils.UserError{Message: "This command can only be used in a server."}
	}

	botData := ctx.Data().(*types.BotData)
	pm := botData.Player
	if pm == nil {
		return &utils.UserError{Message: "The music player is not available."}
	}

	queue, err := pm.GetQueue(context.Background(), *guildID)
	if err != nil {
		return fmt.Errorf("failed to get queue: %w", err)
	}
	player := pm.GetPlayer(*guildID)
	nowPlayingTrack := player.Track()

	if nowPlayingTrack == nil && (queue == nil || len(queue.Tracks) == 0) {
		return ctx.Say("The queue is empty.")
	}

	totalTracks := 0
	if queue != nil {
		totalTracks = len(queue.Tracks)
	}

	embed := discord.NewEmbedBuilder().
		SetColor(0x5865F2)

	getRequesterID := func(track lavalink.Track) string {
		if len(track.UserData) > 0 {
			var dataMap map[string]any
			if err := json.Unmarshal(track.UserData, &dataMap); err == nil {
				if reqID, ok := dataMap["requesterId"].(string); ok {
					return reqID
				}
			}
		}
		return ""
	}

	if nowPlayingTrack != nil {
		if nowPlayingTrack.Info.ArtworkURL != nil {
			embed.SetThumbnail(*nowPlayingTrack.Info.ArtworkURL)
		}

		duration := utils.FormatDuration(int(nowPlayingTrack.Info.Length))
		position := utils.FormatDuration(int(player.Position()))

		trackInfo := fmt.Sprintf("**[%s](%s)**",
			nowPlayingTrack.Info.Title,
			*nowPlayingTrack.Info.URI)
		trackInfo += fmt.Sprintf("- `%s` / `%s`",
			position,
			duration)

		if reqID := getRequesterID(*nowPlayingTrack); reqID != "" {
			trackInfo += fmt.Sprintf("\n- Requested by <@%s>", reqID)
		}

		embed.AddField("▶ Now Playing", trackInfo, false)
	}

	if queue != nil && len(queue.Tracks) > 0 {
		var (
			sb            strings.Builder
			totalDuration lavalink.Duration
		)

		sb.WriteString("\n")
		for i, track := range queue.Tracks {
			totalDuration += track.Info.Length
			if i < 5 {
				duration := utils.FormatDuration(int(track.Info.Length))
				sb.WriteString(fmt.Sprintf("**%d.** [%s](%s) - `%s`", i+1, track.Info.Title, *track.Info.URI, duration))

				if reqID := getRequesterID(track); reqID != "" {
					sb.WriteString(fmt.Sprintf("\n- Requested by <@%s>", reqID))
				}
				sb.WriteString("\n\n")
			}
		}
		sb.WriteString("")

		if len(queue.Tracks) > 5 {
			sb.WriteString(fmt.Sprintf("\n*...and %d more track(s)*", len(queue.Tracks)-5))
		}

		embed.AddField(fmt.Sprintf("Up Next (%d)", totalTracks), sb.String(), false)
		embed.AddField("Total Queue Duration", utils.FormatDuration(int(totalDuration)), true)
	}

	embed.SetTimestamp(time.Now())
	embed.SetFooter(fmt.Sprintf("Requested by %s", ctx.Author().Username), ctx.Author().EffectiveAvatarURL())

	return ctx.SendEmbed(embed.Build())
}
