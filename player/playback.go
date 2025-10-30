package player

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func (p *Player) Play(ctx context.Context, client bot.Client, guildID, channelID snowflake.ID, query string) (*lavalink.Track, error) {
	if err := client.UpdateVoiceState(ctx, guildID, &channelID, false, false); err != nil {
		return nil, fmt.Errorf("failed to join voice channel: %w", err)
	}

	track, err := p.loadTrack(ctx, query)
	if err != nil {
		return nil, err
	}

	player := p.client.Player(guildID)
	currentTrack := player.Track()

	if currentTrack != nil {
		tracks := []QueueTrack{
			{Encoded: track.Encoded},
		}
		_, err := p.AddToQueue(ctx, guildID, tracks)
		if err != nil {
			p.logger.Warn("Failed to add to queue, playing directly", "error", err)
			if err := player.Update(ctx, lavalink.WithTrack(*track)); err != nil {
				return nil, fmt.Errorf("failed to play track: %w", err)
			}
		}
		return track, nil
	}

	if err := player.Update(ctx, lavalink.WithTrack(*track)); err != nil {
		return nil, fmt.Errorf("failed to play track: %w", err)
	}
	return track, nil
}

func (p *Player) PlayNow(ctx context.Context, client bot.Client, guildID, channelID snowflake.ID, query string) (*lavalink.Track, error) {
	if err := client.UpdateVoiceState(ctx, guildID, &channelID, false, false); err != nil {
		return nil, fmt.Errorf("failed to join voice channel: %w", err)
	}

	track, err := p.loadTrack(ctx, query)
	if err != nil {
		return nil, err
	}

	player := p.client.Player(guildID)
	if err := player.Update(ctx, lavalink.WithTrack(*track)); err != nil {
		return nil, fmt.Errorf("failed to play track: %w", err)
	}
	return track, nil
}

func (p *Player) Stop(ctx context.Context, guildID snowflake.ID) error {
	player := p.client.Player(guildID)
	return player.Update(ctx, lavalink.WithNullTrack())
}

func (p *Player) Pause(ctx context.Context, guildID snowflake.ID, paused bool) error {
	player := p.client.Player(guildID)
	return player.Update(ctx, lavalink.WithPaused(paused))
}

func (p *Player) Seek(ctx context.Context, guildID snowflake.ID, position int64) error {
	player := p.client.Player(guildID)
	return player.Update(ctx, lavalink.WithPosition(lavalink.Duration(position)))
}

func (p *Player) SetVolume(ctx context.Context, guildID snowflake.ID, volume int) error {
	player := p.client.Player(guildID)
	return player.Update(ctx, lavalink.WithVolume(volume))
}

func (p *Player) GetCurrentTrack(guildID snowflake.ID) *lavalink.Track {
	player := p.client.Player(guildID)
	return player.Track()
}

func (p *Player) IsPlaying(guildID snowflake.ID) bool {
	player := p.client.Player(guildID)
	return player.Track() != nil && !player.Paused()
}

func (p *Player) IsPaused(guildID snowflake.ID) bool {
	player := p.client.Player(guildID)
	return player.Paused()
}

func (p *Player) loadTrack(ctx context.Context, query string) (*lavalink.Track, error) {
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = "ytsearch:" + query
	}

	var toPlay *lavalink.Track
	var searchErr error

	p.client.BestNode().LoadTracksHandler(ctx, query, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			toPlay = &track
		},
		func(playlist lavalink.Playlist) {
			if len(playlist.Tracks) > 0 {
				toPlay = &playlist.Tracks[0]
			}
		},
		func(tracks []lavalink.Track) {
			if len(tracks) > 0 {
				toPlay = &tracks[0]
			}
		},
		func() {
			searchErr = fmt.Errorf("no results found")
		},
		func(err error) {
			searchErr = err
		},
	))

	if searchErr != nil {
		return nil, searchErr
	}

	if toPlay == nil {
		return nil, fmt.Errorf("no track found")
	}

	return toPlay, nil
}

func (p *Player) LoadPlaylist(ctx context.Context, url string) ([]lavalink.Track, error) {
	var tracks []lavalink.Track
	var searchErr error

	p.client.BestNode().LoadTracksHandler(ctx, url, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			tracks = []lavalink.Track{track}
		},
		func(playlist lavalink.Playlist) {
			tracks = playlist.Tracks
		},
		func(searchTracks []lavalink.Track) {
			tracks = searchTracks
		},
		func() {
			searchErr = fmt.Errorf("no results found")
		},
		func(err error) {
			searchErr = err
		},
	))

	if searchErr != nil {
		return nil, searchErr
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}
