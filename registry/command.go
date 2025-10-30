package registry

import "github.com/disgoorg/disgo/discord"

type ExecuteFunc func(ctx *Context) error

type Command struct {
	Name          string
	Description   string
	PrefixCommand bool
	SlashCommand  bool
	Aliases       []string
	Execute       ExecuteFunc
	Options       []discord.ApplicationCommandOption
}
