package modules

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/goland-express/flexo/player"
	"github.com/goland-express/flexo/registry"
	"github.com/goland-express/flexo/types"
	"github.com/goland-express/flexo/utils"
)

var ErrInvalidBotData = errors.New("invalid bot data type")

type MusicModule struct{}

func (m *MusicModule) Name() string {
	return "Music"
}

func (m *MusicModule) Register(r *registry.Registry) {
	r.Add(&registry.Command{
		Name:          "play",
		Description:   "Play a song in the voice channel.",
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
		Description:   "Skip to the next song in the queue.",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"s", "next"},
		Execute:       m.executeSkip,
	})

	r.Add(&registry.Command{
		Name:          "queue",
		Description:   "Show the current music queue.",
		PrefixCommand: true,
		SlashCommand:  true,
		Aliases:       []string{"q"},
		Execute:       m.executeQueue,
	})
}

func (m *MusicModule) executePlay(ctx *registry.Context) error {
	guildID, err := getGuildID(ctx)
	if err != nil {
		return err
	}

	voiceState, ok := ctx.Client().Caches().VoiceState(guildID, ctx.Author().ID)
	if !ok || voiceState.ChannelID == nil {
		return &utils.UserError{Message: "You need to be in a voice channel to use this command."}
	}

	query := getQuery(ctx)
	if query == "" {
		return &utils.UserError{Message: "You need to specify a song. Ex: `!play <song>`"}
	}

	playerManager, err := getPlayerManager(ctx)
	if err != nil {
		return err
	}

	userData := map[string]any{
		"requesterId":   ctx.Author().ID.String(),
		"requesterName": ctx.Author().Username,
	}

	track, position, err := playerManager.Play(context.Background(), ctx.Client(), guildID, *voiceState.ChannelID, query, userData)
	if err != nil {
		return fmt.Errorf("failed to play song: %w", err)
	}

	embed := buildPlayEmbed(track, position, ctx.Author())
	if err := ctx.SendEmbed(embed); err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}

func (m *MusicModule) executeSkip(ctx *registry.Context) error {
	guildID, err := getGuildID(ctx)
	if err != nil {
		return err
	}

	playerManager, err := getPlayerManager(ctx)
	if err != nil {
		return err
	}

	track, err := playerManager.NextTrack(context.Background(), guildID)
	if err != nil {
		return fmt.Errorf("failed to skip song: %w", err)
	}

	embed := buildSkipEmbed(track)
	if err := ctx.SendEmbed(embed); err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}

func (m *MusicModule) executeQueue(ctx *registry.Context) error {
	guildID, err := getGuildID(ctx)
	if err != nil {
		return err
	}

	playerManager, err := getPlayerManager(ctx)
	if err != nil {
		return err
	}

	queue, err := playerManager.GetQueue(context.Background(), guildID)
	if err != nil {
		return fmt.Errorf("failed to get queue: %w", err)
	}

	player := playerManager.GetPlayer(guildID)
	nowPlayingTrack := player.Track()

	if nowPlayingTrack == nil && (queue == nil || len(queue.Tracks) == 0) {
		if err := ctx.Say("The queue is empty."); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		return nil
	}

	embed := buildQueueEmbed(ctx, nowPlayingTrack, queue, player.Position())
	if err := ctx.SendEmbed(embed); err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}

func buildPlayEmbed(track *lavalink.Track, position int, author discord.User) discord.Embed {
	builder := discord.NewEmbedBuilder().
		SetTitle(track.Info.Title).
		SetURL(*track.Info.URI).
		SetColor(0x2371AB).
		AddField("Duration", utils.FormatDuration(int(track.Info.Length)), true).
		SetAuthor(track.Info.Author, "", "").
		SetFooter("Requested by "+author.Username, author.EffectiveAvatarURL()).
		SetTimestamp(time.Now())

	if track.Info.ArtworkURL != nil {
		builder.SetThumbnail(*track.Info.ArtworkURL)
	}

	if position > 0 {
		builder.AddField("Queue Position", fmt.Sprint(position), true)
	}

	return builder.Build()
}

func buildSkipEmbed(track *lavalink.Track) discord.Embed {
	builder := discord.NewEmbedBuilder().SetColor(0x1DB954)

	if track == nil {
		builder.SetDescription("The queue has ended.")
	} else {
		builder.SetTitle("Song Skipped")
		builder.SetDescription(fmt.Sprintf("Now playing: **[%s](%s)**", track.Info.Title, *track.Info.URI))
		if track.Info.ArtworkURL != nil {
			builder.SetThumbnail(*track.Info.ArtworkURL)
		}
	}

	return builder.Build()
}

func buildQueueEmbed(ctx *registry.Context, nowPlaying *lavalink.Track, queue *player.Queue, position lavalink.Duration) discord.Embed {
	embed := discord.NewEmbedBuilder().
		SetColor(0x5865F2).
		SetTimestamp(time.Now()).
		SetFooter("Requested by "+ctx.Author().Username, ctx.Author().EffectiveAvatarURL())

	if nowPlaying != nil {
		if nowPlaying.Info.ArtworkURL != nil {
			embed.SetThumbnail(*nowPlaying.Info.ArtworkURL)
		}

		duration := utils.FormatDuration(int(nowPlaying.Info.Length))
		currentPosition := utils.FormatDuration(int(position))

		trackInfo := fmt.Sprintf("**[%s](%s)** - `%s` / `%s`",
			nowPlaying.Info.Title, *nowPlaying.Info.URI, currentPosition, duration)

		if reqID := getRequesterID(*nowPlaying); reqID != "" {
			trackInfo += fmt.Sprintf("\n- Requested by <@%s>", reqID)
		}

		embed.AddField("▶ Now Playing", trackInfo, false)
	}

	if queue != nil && len(queue.Tracks) > 0 {
		var (
			sb            strings.Builder
			totalDuration lavalink.Duration
		)

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

		if len(queue.Tracks) > 5 {
			sb.WriteString(fmt.Sprintf("\n*...and %d more song(s)*", len(queue.Tracks)-5))
		}

		embed.AddField(fmt.Sprintf("Up Next (%d)", len(queue.Tracks)), sb.String(), false)
		embed.AddField("Total Queue Duration", utils.FormatDuration(int(totalDuration)), true)
	}

	return embed.Build()
}

func getGuildID(ctx *registry.Context) (snowflake.ID, error) {
	if ctx.GuildID() == nil {
		return 0, &utils.UserError{Message: "This command can only be used in a server."}
	}

	return *ctx.GuildID(), nil
}

func getPlayerManager(ctx *registry.Context) (*player.Player, error) {
	botData, ok := ctx.Data().(*types.BotData)
	if !ok {
		return nil, ErrInvalidBotData
	}

	if botData.Player == nil {
		return nil, &utils.UserError{Message: "The music player is not available."}
	}

	return botData.Player, nil
}

func getQuery(ctx *registry.Context) string {
	if ctx.IsSlash() {
		query, _ := ctx.GetStringOption("query")
		return query
	}

	return strings.Join(ctx.Args(), " ")
}

func getRequesterID(track lavalink.Track) string {
	if len(track.UserData) == 0 {
		return ""
	}

	var dataMap map[string]any
	if err := json.Unmarshal(track.UserData, &dataMap); err == nil {
		if reqID, ok := dataMap["requesterId"].(string); ok {
			return reqID
		}
	}

	return ""
}
