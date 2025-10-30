package player

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

type Player struct {
	client  disgolink.Client
	logger  *slog.Logger
	handler QueueEventHandler
}

func New(appID snowflake.ID, lavalinkHost, lavalinkPassword string) (*Player, error) {
	logger := slog.Default()

	queuePlugin := newQueuePlugin(logger)

	client := disgolink.New(appID,
		disgolink.WithPlugins(queuePlugin),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("Connecting to Lavalink...",
		slog.String("host", lavalinkHost),
	)

	node, err := client.AddNode(ctx, disgolink.NodeConfig{
		Name:     "main",
		Address:  lavalinkHost,
		Password: lavalinkPassword,
		Secure:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add lavalink node: %w", err)
	}

	logger.Info("Lavalink node added",
		slog.String("address", node.Config().Address),
	)

	player := &Player{
		client: client,
		logger: logger,
	}

	return player, nil
}

func (p *Player) GetPlayer(guildID snowflake.ID) disgolink.Player {
	return p.client.Player(guildID)
}

func (p *Player) BestNode() disgolink.Node {
	return p.client.BestNode()
}

func (p *Player) SetQueueEventHandler(handler QueueEventHandler) {
	p.handler = handler
}

func (p *Player) OnQueueEnd(guildID snowflake.ID) {
	if p.handler != nil {
		p.handler.OnQueueEnd(guildID)
	}
	p.logger.Info("Queue ended", slog.String("guild_id", guildID.String()))
}

var (
	_ disgolink.EventPlugins = (*queuePlugin)(nil)
	_ disgolink.Plugin       = (*queuePlugin)(nil)
)

type queuePlugin struct {
	eventPlugins []disgolink.EventPlugin
}

func newQueuePlugin(logger *slog.Logger) *queuePlugin {
	return &queuePlugin{
		eventPlugins: []disgolink.EventPlugin{
			&queueEndHandler{
				logger: logger,
			},
		},
	}
}

func (p *queuePlugin) EventPlugins() []disgolink.EventPlugin {
	return p.eventPlugins
}

func (p *queuePlugin) Name() string {
	return "lavaqueue"
}

func (p *queuePlugin) Version() string {
	return "0.0.0"
}

var _ disgolink.EventPlugin = (*queueEndHandler)(nil)

type queueEndHandler struct {
	logger *slog.Logger
}

func (h *queueEndHandler) Event() lavalink.EventType {
	return EventTypeQueueEnd
}

func (h *queueEndHandler) OnEventInvocation(player disgolink.Player, data []byte) {
	var e QueueEndEvent
	if err := json.Unmarshal(data, &e); err != nil {
		h.logger.Error("Failed to unmarshal QueueEndEvent", slog.Any("err", err))
		return
	}

	h.logger.Info("Queue end event received",
		slog.String("guild_id", e.GuildID.String()),
	)
}

const (
	EventTypeQueueEnd lavalink.EventType = "QueueEndEvent"
)

type QueueEndEvent struct {
	OpValue   lavalink.Op        `json:"op"`
	TypeValue lavalink.EventType `json:"type"`
	GuildID   snowflake.ID       `json:"guildId"`
}

func (e QueueEndEvent) Op() lavalink.Op {
	return e.OpValue
}

func (e QueueEndEvent) Type() lavalink.EventType {
	return e.TypeValue
}
