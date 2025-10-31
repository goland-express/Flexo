package player

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

var (
	ErrQueueEmpty      = errors.New("queue is empty")
	ErrFailedToStop    = errors.New("failed to stop player")
	ErrUnmarshalFailed = errors.New("failed to unmarshal response")
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
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	var queue Queue
	if err = unmarshalBody(response, &queue); err != nil {
		return nil, fmt.Errorf("unmarshal queue: %w", err)
	}

	return &queue, nil
}

func (p *Player) AddToQueue(ctx context.Context, guildID snowflake.ID, tracks []QueueTrack) (*lavalink.Track, error) {
	node := p.BestNode()
	requestBody, err := marshalBody(tracks)
	if err != nil {
		return nil, fmt.Errorf("marshal tracks: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/tracks", node.SessionID(), guildID), requestBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := node.Rest().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		return nil, ErrQueueEmpty
	}

	var track lavalink.Track
	if err = unmarshalBody(response, &track); err != nil {
		return nil, fmt.Errorf("unmarshal track: %w", err)
	}

	return &track, nil
}

func (p *Player) NextTrack(ctx context.Context, guildID snowflake.ID) (*lavalink.Track, error) {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/next?count=1", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		if stopErr := p.Stop(ctx, guildID); stopErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrQueueEmpty, errors.Join(ErrFailedToStop, stopErr))
		}
		return nil, ErrQueueEmpty
	}

	var track lavalink.Track
	if err = unmarshalBody(response, &track); err != nil {
		if stopErr := p.Stop(ctx, guildID); stopErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnmarshalFailed, errors.Join(ErrFailedToStop, stopErr))
		}
		return nil, fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
	}

	return &track, nil
}

func (p *Player) PreviousTrack(ctx context.Context, guildID snowflake.ID) (*lavalink.Track, error) {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/previous?count=1", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	var track lavalink.Track
	if err = unmarshalBody(response, &track); err != nil {
		return nil, fmt.Errorf("unmarshal track: %w", err)
	}

	return &track, nil
}

func (p *Player) ShuffleQueue(ctx context.Context, guildID snowflake.ID) error {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/shuffle", node.SessionID(), guildID), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	if err := unmarshalBody(response, nil); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (p *Player) ClearQueue(ctx context.Context, guildID snowflake.ID) error {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue", node.SessionID(), guildID), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	if err := unmarshalBody(response, nil); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (p *Player) RemoveTrack(ctx context.Context, guildID snowflake.ID, trackID int) error {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("/v4/sessions/%s/players/%s/queue/tracks/%d", node.SessionID(), guildID, trackID), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	if err := unmarshalBody(response, nil); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (p *Player) GetHistory(ctx context.Context, guildID snowflake.ID) ([]lavalink.Track, error) {
	node := p.BestNode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("/v4/sessions/%s/players/%s/history", node.SessionID(), guildID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	response, err := node.Rest().Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()

	var history []lavalink.Track
	if err = unmarshalBody(response, &history); err != nil {
		return nil, fmt.Errorf("unmarshal history: %w", err)
	}

	return history, nil
}

func marshalBody(value any) (io.Reader, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	return bytes.NewReader(data), nil
}

func unmarshalBody(response *http.Response, value any) error {
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var lavalinkError lavalink.Error
		if err := json.NewDecoder(response.Body).Decode(&lavalinkError); err != nil {
			return fmt.Errorf("decode lavalink error: %w", err)
		}
		return fmt.Errorf("lavalink error: %w", lavalinkError)
	}

	if response.StatusCode == http.StatusNoContent {
		return nil
	}

	if value == nil {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(value); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	return nil
}
