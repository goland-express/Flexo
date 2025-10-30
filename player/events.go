package player

import (
	"context"

	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type QueueEventHandler interface {
	OnTrackEnd(guildID snowflake.ID, track lavalink.Track)
	OnQueueEnd(guildID snowflake.ID)
}

func (p *Player) OnVoiceStateUpdate(e *events.GuildVoiceStateUpdate) {
	if e.VoiceState.UserID != e.Client().ApplicationID() {
		return
	}
	p.client.OnVoiceStateUpdate(
		context.Background(),
		e.VoiceState.GuildID,
		e.VoiceState.ChannelID,
		e.VoiceState.SessionID,
	)
}

func (p *Player) OnVoiceServerUpdate(e *events.VoiceServerUpdate) {
	if e.Endpoint == nil {
		return
	}
	p.client.OnVoiceServerUpdate(
		context.Background(),
		e.GuildID,
		e.Token,
		*e.Endpoint,
	)
}

func (p *Player) OnTrackEnd(guildID snowflake.ID, track lavalink.Track, endReason string) {
	if p.handler != nil {
		p.handler.OnTrackEnd(guildID, track)
	}

	p.logger.Info("Track ended",
		"guild_id", guildID.String(),
		"track", track.Info.Title,
		"reason", endReason,
	)
}
