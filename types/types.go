package types

import (
	"time"

	"github.com/goland-express/Flexo/player"
)

type BotData struct {
	StartTime time.Time
	Player    *player.Player
}
