package types

import (
	"flexo/player"
	"time"
)

type BotData struct {
	StartTime time.Time
	Player    *player.Player
}
