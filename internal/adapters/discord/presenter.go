package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	defaultEmbedColor = 0x2F80ED
	errorEmbedColor   = 0xD92D20
)

type Presenter struct{}

type Embed struct {
	Title       string
	Description string
	Color       int
	Fields      []EmbedField
}

type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

func NewPresenter() Presenter {
	return Presenter{}
}

func (presenter Presenter) EphemeralMessage(content string) *discordgo.InteractionResponseData {
	data := presenter.PublicMessage(content)
	data.Flags = discordgo.MessageFlagsEphemeral

	return data
}

func (presenter Presenter) PublicMessage(content string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content:         strings.TrimSpace(content),
		AllowedMentions: noMentions(),
	}
}

func (presenter Presenter) ErrorMessage(content string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content:         strings.TrimSpace(content),
		Flags:           discordgo.MessageFlagsEphemeral,
		AllowedMentions: noMentions(),
		Embeds: []*discordgo.MessageEmbed{
			presenter.Embed(Embed{
				Title:       "Erro",
				Description: strings.TrimSpace(content),
				Color:       errorEmbedColor,
			}),
		},
	}
}

func (presenter Presenter) Embed(embed Embed) *discordgo.MessageEmbed {
	color := embed.Color

	if color == 0 {
		color = defaultEmbedColor
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(embed.Fields))

	for _, field := range embed.Fields {
		name := strings.TrimSpace(field.Name)
		value := strings.TrimSpace(field.Value)

		if name == "" || value == "" {
			continue
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   name,
			Value:  value,
			Inline: field.Inline,
		})
	}

	return &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       strings.TrimSpace(embed.Title),
		Description: strings.TrimSpace(embed.Description),
		Color:       color,
		Fields:      fields,
	}
}

func noMentions() *discordgo.MessageAllowedMentions {
	return &discordgo.MessageAllowedMentions{
		Parse:       []discordgo.AllowedMentionType{},
		Roles:       []string{},
		Users:       []string{},
		RepliedUser: false,
	}
}
