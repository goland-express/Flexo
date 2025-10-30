package registry

import (
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Registry struct {
	commands []*Command
	data     Data
	prefix   string
	onError  func(err error, ctx *Context)
	onReady  func(event *events.Ready)
	mu       sync.RWMutex
}

type Options struct {
	Commands []*Command
	Data     Data
	Prefix   string
	OnError  func(err error, ctx *Context)
	OnReady  func(event *events.Ready)
}

func New(opts Options) *Registry {
	if opts.OnError == nil {
		opts.OnError = defaultErrorFunc
	}

	return &Registry{
		commands: opts.Commands,
		data:     opts.Data,
		prefix:   opts.Prefix,
		onError:  opts.OnError,
		onReady:  opts.OnReady,
	}
}

func defaultErrorFunc(err error, ctx *Context) {
	slog.Error("Error executing command", slog.Any("error", err))
	if ctx != nil {
		_ = ctx.Say("An error occurred while executing the command.")
	}
}

func (r *Registry) OnMessage(event *events.MessageCreate) {
	if event.Message.Author.Bot {
		return
	}

	content := event.Message.Content
	if !strings.HasPrefix(content, r.prefix) {
		return
	}

	commandName := strings.TrimPrefix(content, r.prefix)

	if idx := strings.Index(commandName, " "); idx != -1 {
		commandName = commandName[:idx]
	}

	r.execute(commandName, &Context{
		client:      event.Client(),
		messageData: event,
		data:        r.data,
		isSlash:     false,
	}, false)
}

func (r *Registry) OnSlashCommand(event *events.ApplicationCommandInteractionCreate) {
	commandName := event.Data.CommandName()

	r.execute(commandName, &Context{
		client:    event.Client(),
		slashData: event,
		data:      r.data,
		isSlash:   true,
	}, true)
}

func (r *Registry) execute(name string, ctx *Context, isSlash bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cmd := range r.commands {
		if isSlash && !cmd.SlashCommand {
			continue
		}
		if !isSlash && !cmd.PrefixCommand {
			continue
		}

		if cmd.Name != name && !slices.Contains(cmd.Aliases, name) {
			continue
		}

		if err := cmd.Execute(ctx); err != nil {
			r.onError(err, ctx)
		}
		return
	}
}

func (r *Registry) OnReady(event *events.Ready) {
	if err := r.RegisterSlash(event.Client()); err != nil {
		slog.Error("Failed to register slash commands", slog.Any("error", err))
		return
	}

	if r.onReady != nil {
		r.onReady(event)
	}
}

func (r *Registry) RegisterSlash(client bot.Client) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var slashCommands []discord.ApplicationCommandCreate
	for _, cmd := range r.commands {
		if !cmd.SlashCommand {
			continue
		}

		slashCommands = append(slashCommands, discord.SlashCommandCreate{
			Name:        cmd.Name,
			Description: cmd.Description,
			Options:     cmd.Options,
		})
	}

	if len(slashCommands) == 0 {
		return nil
	}

	_, err := client.Rest().SetGlobalCommands(client.ApplicationID(), slashCommands)
	if err != nil {
		return err
	}

	return nil
}

func (r *Registry) Commands() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commands
}

func (r *Registry) Add(cmd *Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, cmd)
}
