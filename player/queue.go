package player

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

type Queue struct {
	Tracks []lavalink.Track `json:"tracks"`
}

type QueueUpdate struct {
	Tracks []QueueTrack `json:"tracks,omitempty"`
}

type QueueTrack struct {
	Encoded  string         `json:"encoded"`
	UserData map[string]any `json:"userData,omitempty"`
}

func (p *Player) GetQueue(ctx context.Context, guildID snowflake.ID) (*Queue, error) {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	var queue Queue
	if err = unmarshalBody(rs, &queue); err != nil {
		return nil, err
	}

	return &queue, nil
}

func (p *Player) AddToQueue(ctx context.Context, guildID snowflake.ID, tracks []QueueTrack) (*lavalink.Track, error) {
	node := p.BestNode()
	rqBody, err := marshalBody(tracks)
	if err != nil {
		return nil, err
	}

	rq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/tracks", node.SessionID(), guildID), rqBody)
	if err != nil {
		return nil, err
	}
	rq.Header.Add("Content-Type", "application/json")

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	if rs.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var track lavalink.Track
	if err = unmarshalBody(rs, &track); err != nil {
		return nil, err
	}
	return &track, nil
}

func (p *Player) NextTrack(ctx context.Context, guildID snowflake.ID) (*lavalink.Track, error) {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/next?count=1", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	if rs.StatusCode == http.StatusNoContent {
		if stopErr := p.Stop(ctx, guildID); stopErr != nil {
			return nil, fmt.Errorf("queue empty, but failed to stop player: %w", stopErr)
		}
		return nil, nil
	}

	var track lavalink.Track
	if err = unmarshalBody(rs, &track); err != nil {
		if stopErr := p.Stop(ctx, guildID); stopErr != nil {
			return nil, fmt.Errorf("failed to unmarshal track and failed to stop player: %w", stopErr)
		}
		return nil, nil
	}

	return &track, nil
}
func (p *Player) PreviousTrack(ctx context.Context, guildID snowflake.ID) (*lavalink.Track, error) {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/previous?count=1", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	var track lavalink.Track
	if err = unmarshalBody(rs, &track); err != nil {
		return nil, err
	}
	return &track, nil
}

func (p *Player) ShuffleQueue(ctx context.Context, guildID snowflake.ID) error {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/shuffle", node.SessionID(), guildID), nil)
	if err != nil {
		return err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	return unmarshalBody(rs, nil)
}

func (p *Player) ClearQueue(ctx context.Context, guildID snowflake.ID) error {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue", node.SessionID(), guildID), nil)
	if err != nil {
		return err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	return unmarshalBody(rs, nil)
}

func (p *Player) RemoveTrack(ctx context.Context, guildID snowflake.ID, trackID int) error {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/tracks/%d", node.SessionID(), guildID, trackID), nil)
	if err != nil {
		return err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	return unmarshalBody(rs, nil)
}

func (p *Player) GetHistory(ctx context.Context, guildID snowflake.ID) ([]lavalink.Track, error) {
	node := p.BestNode()
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("/v4/sessions/%s/players/%s/history", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, err
	}

	rs, err := node.Rest().Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	var history []lavalink.Track
	if err = unmarshalBody(rs, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func marshalBody(v any) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func unmarshalBody(rs *http.Response, v any) error {
	if rs.StatusCode < 200 || rs.StatusCode >= 300 {
		var lavalinkError lavalink.Error
		if err := json.NewDecoder(rs.Body).Decode(&lavalinkError); err != nil {
			return err
		}
		return lavalinkError
	}

	if rs.StatusCode == http.StatusNoContent {
		return nil
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(rs.Body).Decode(v)
}
