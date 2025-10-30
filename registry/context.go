package registry

import (
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type Context struct {
	client      bot.Client
	messageData *events.MessageCreate
	slashData   *events.ApplicationCommandInteractionCreate
	data        Data
	isSlash     bool
}

func (c *Context) Client() bot.Client {
	return c.client
}

func (c *Context) Data() Data {
	return c.data
}

func (c *Context) IsSlash() bool {
	return c.isSlash
}

func (c *Context) ChannelID() snowflake.ID {
	if c.isSlash {
		return c.slashData.Channel().ID()
	}
	return c.messageData.ChannelID
}

func (c *Context) GuildID() *snowflake.ID {
	if c.isSlash {
		return c.slashData.GuildID()
	}
	if c.messageData.Message.GuildID != nil {
		return c.messageData.Message.GuildID
	}
	return nil
}

func (c *Context) Author() discord.User {
	if c.isSlash {
		return c.slashData.User()
	}
	return c.messageData.Message.Author
}

func (c *Context) Say(content string) error {
	builder := discord.NewMessageCreateBuilder().SetContent(content)

	if c.isSlash {
		return c.slashData.CreateMessage(builder.Build())
	}

	_, err := c.messageData.Client().Rest().CreateMessage(
		c.messageData.ChannelID,
		builder.Build(),
	)
	return err
}

func (c *Context) Reply(content string) error {
	builder := discord.NewMessageCreateBuilder().SetContent(content)

	if c.isSlash {
		return c.slashData.CreateMessage(builder.Build())
	}

	builder.SetMessageReference(&discord.MessageReference{
		MessageID: &c.messageData.Message.ID,
	})

	_, err := c.messageData.Client().Rest().CreateMessage(
		c.messageData.ChannelID,
		builder.Build(),
	)
	return err
}

func (c *Context) SendEmbed(embed discord.Embed) error {
	builder := discord.NewMessageCreateBuilder().SetEmbeds(embed)

	if c.isSlash {
		return c.slashData.CreateMessage(builder.Build())
	}

	_, err := c.messageData.Client().Rest().CreateMessage(
		c.messageData.ChannelID,
		builder.Build(),
	)
	return err
}

func (c *Context) Args() []string {
	if c.isSlash {
		return []string{}
	}
	content := c.messageData.Message.Content

	parts := strings.Fields(content)

	if len(parts) <= 1 {
		return []string{}
	}
	return parts[1:]
}

func (c *Context) GetStringOption(name string) (string, bool) {
	if !c.isSlash {
		return "", false
	}

	data := c.slashData.SlashCommandInteractionData()
	return data.OptString(name)
}

func (c *Context) GetIntOption(name string) (int64, bool) {
	if !c.isSlash {
		return 0, false
	}

	data := c.slashData.SlashCommandInteractionData()
	if opt, ok := data.OptInt(name); ok {
		return int64(opt), true
	}
	return 0, false
}

func (c *Context) GetBoolOption(name string) (bool, bool) {
	if !c.isSlash {
		return false, false
	}

	data := c.slashData.SlashCommandInteractionData()
	return data.OptBool(name)
}

func (c *Context) GetUserOption(name string) (discord.User, bool) {
	if !c.isSlash {
		return discord.User{}, false
	}

	data := c.slashData.SlashCommandInteractionData()
	return data.OptUser(name)
}
