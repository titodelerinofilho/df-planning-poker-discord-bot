package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestPresenterBuildsEphemeralMessage(t *testing.T) {
	presenter := NewPresenter()

	data := presenter.EphemeralMessage("  hello @everyone  ")

	if data.Content != "hello @everyone" {
		t.Fatalf("content = %q, want trimmed content", data.Content)
	}

	if data.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("flags = %v, want ephemeral", data.Flags)
	}

	assertNoMentions(t, data.AllowedMentions)
}

func TestPresenterBuildsPublicMessage(t *testing.T) {
	presenter := NewPresenter()

	data := presenter.PublicMessage("hello")

	if data.Content != "hello" {
		t.Fatalf("content = %q, want hello", data.Content)
	}

	if data.Flags != 0 {
		t.Fatalf("flags = %v, want public message", data.Flags)
	}

	assertNoMentions(t, data.AllowedMentions)
}

func TestPresenterBuildsErrorMessage(t *testing.T) {
	presenter := NewPresenter()

	data := presenter.ErrorMessage(" Algo deu errado ")

	if data.Content != "Algo deu errado" {
		t.Fatalf("content = %q, want trimmed error", data.Content)
	}

	if data.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("flags = %v, want ephemeral", data.Flags)
	}

	assertNoMentions(t, data.AllowedMentions)

	if len(data.Embeds) != 1 {
		t.Fatalf("embeds = %d, want 1", len(data.Embeds))
	}

	embed := data.Embeds[0]

	if embed.Title != "Erro" {
		t.Fatalf("embed title = %q, want Erro", embed.Title)
	}

	if embed.Description != "Algo deu errado" {
		t.Fatalf("embed description = %q, want trimmed error", embed.Description)
	}

	if embed.Color != errorEmbedColor {
		t.Fatalf("embed color = %d, want %d", embed.Color, errorEmbedColor)
	}
}

func TestPresenterBuildsEmbed(t *testing.T) {
	presenter := NewPresenter()

	embed := presenter.Embed(Embed{
		Title:       " Status ",
		Description: " Tudo certo ",
		Fields: []EmbedField{
			{Name: " Latencia ", Value: " 42ms ", Inline: true},
			{Name: "", Value: "ignored"},
			{Name: "ignored", Value: ""},
		},
	})

	if embed.Type != discordgo.EmbedTypeRich {
		t.Fatalf("embed type = %q, want rich", embed.Type)
	}

	if embed.Title != "Status" {
		t.Fatalf("embed title = %q, want Status", embed.Title)
	}

	if embed.Description != "Tudo certo" {
		t.Fatalf("embed description = %q, want Tudo certo", embed.Description)
	}

	if embed.Color != defaultEmbedColor {
		t.Fatalf("embed color = %d, want default color", embed.Color)
	}

	if len(embed.Fields) != 1 {
		t.Fatalf("fields = %d, want 1", len(embed.Fields))
	}

	field := embed.Fields[0]

	if field.Name != "Latencia" {
		t.Fatalf("field name = %q, want Latencia", field.Name)
	}

	if field.Value != "42ms" {
		t.Fatalf("field value = %q, want 42ms", field.Value)
	}

	if !field.Inline {
		t.Fatal("field inline = false, want true")
	}
}

func TestPresenterBuildsEmbedWithCustomColor(t *testing.T) {
	presenter := NewPresenter()

	embed := presenter.Embed(Embed{
		Title: "Custom",
		Color: 0x123456,
	})

	if embed.Color != 0x123456 {
		t.Fatalf("embed color = %d, want custom color", embed.Color)
	}
}

func assertNoMentions(t *testing.T, mentions *discordgo.MessageAllowedMentions) {
	t.Helper()

	if mentions == nil {
		t.Fatal("allowed mentions is nil")
	}

	if len(mentions.Parse) != 0 {
		t.Fatalf("allowed mention parse = %v, want empty", mentions.Parse)
	}

	if len(mentions.Roles) != 0 {
		t.Fatalf("allowed mention roles = %v, want empty", mentions.Roles)
	}

	if len(mentions.Users) != 0 {
		t.Fatalf("allowed mention users = %v, want empty", mentions.Users)
	}

	if mentions.RepliedUser {
		t.Fatal("replied user mention = true, want false")
	}
}
