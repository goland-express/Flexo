package types

import (
	"time"

	"flexo/player"
)

type BotData struct {
	StartTime time.Time
	Player    *player.Player
}
