package types

import (
	"time"

	"github.com/goland-express/flexo/player"
)

type BotData struct {
	StartTime time.Time
	Player    *player.Player
}
