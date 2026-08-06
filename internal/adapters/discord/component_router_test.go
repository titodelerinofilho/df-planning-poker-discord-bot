package discord

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestParseComponentRouteAcceptsOpaqueResource(t *testing.T) {
	route, err := parseComponentRoute("planning:vote:session_123-abc.def")

	if err != nil {
		t.Fatalf("parseComponentRoute() error = %v", err)
	}

	if route.Namespace != "planning" {
		t.Fatalf("namespace = %q, want planning", route.Namespace)
	}

	if route.Action != "vote" {
		t.Fatalf("action = %q, want vote", route.Action)
	}

	if route.Resource != "session_123-abc.def" {
		t.Fatalf("resource = %q, want opaque resource", route.Resource)
	}
}

func TestParseComponentRouteRejectsInvalidIDs(t *testing.T) {
	tests := []struct {
		name     string
		customID string
	}{
		{name: "empty", customID: ""},
		{name: "missing resource", customID: "planning:vote"},
		{name: "extra part", customID: "planning:vote:session:estimate"},
		{name: "blank namespace", customID: ":vote:session"},
		{name: "blank action", customID: "planning::session"},
		{name: "blank resource", customID: "planning:vote:"},
		{name: "invalid namespace", customID: "planning!:vote:session"},
		{name: "invalid action", customID: "planning:cast vote:session"},
		{name: "invalid resource", customID: "planning:vote:session/1"},
		{name: "unicode resource", customID: "planning:vote:sessao\u00e7"},
		{name: "too long", customID: strings.Repeat("a", maxComponentCustomIDLength+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseComponentRoute(tt.customID)

			if !errors.Is(err, ErrComponentIDInvalid) {
				t.Fatalf("parseComponentRoute() error = %v, want ErrComponentIDInvalid", err)
			}
		})
	}
}

func TestComponentRouterRoutesByNamespaceAndAction(t *testing.T) {
	router := newComponentRouter()
	var routed componentRoute
	calls := 0
	router.Handle(componentRoute{Namespace: "planning", Action: "join"}, func(_ *discordgo.InteractionCreate, route componentRoute) error {
		calls++
		routed = route

		return nil
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.MessageComponentInteractionData{
		CustomID: "planning:join:session-123",
	})

	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}

	if routed.Resource != "session-123" {
		t.Fatalf("resource = %q, want session-123", routed.Resource)
	}
}

func TestComponentRouterReturnsStandardRouteError(t *testing.T) {
	router := newComponentRouter()

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.MessageComponentInteractionData{
		CustomID: "planning:join:session-123",
	})

	if !errors.Is(err, ErrComponentRouteNotFound) {
		t.Fatalf("Route() error = %v, want ErrComponentRouteNotFound", err)
	}
}

func TestComponentRouterWrapsHandlerError(t *testing.T) {
	handlerErr := errors.New("handler failed")
	router := newComponentRouter()
	router.Handle(componentRoute{Namespace: "planning", Action: "join"}, func(*discordgo.InteractionCreate, componentRoute) error {
		return handlerErr
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.MessageComponentInteractionData{
		CustomID: "planning:join:session-123",
	})

	if !errors.Is(err, handlerErr) {
		t.Fatalf("Route() error = %v, want handler error", err)
	}
}
