package discord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const maxComponentCustomIDLength = 100

var (
	ErrComponentIDInvalid     = errors.New("discord component id is invalid")
	ErrComponentRouteNotFound = errors.New("discord component route not found")
)

type componentHandler func(*discordgo.InteractionCreate, componentRoute) error

type componentRouter struct {
	handlers map[string]componentHandler
}

type componentRoute struct {
	Namespace string
	Action    string
	Resource  string
}

func newDefaultComponentRouter(*Bot) *componentRouter {
	return newComponentRouter()
}

func newComponentRouter() *componentRouter {
	return &componentRouter{
		handlers: make(map[string]componentHandler),
	}
}

func (router *componentRouter) Handle(route componentRoute, handler componentHandler) {
	router.handlers[route.key()] = handler
}

func (router *componentRouter) Route(event *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData) error {
	route, err := parseComponentRoute(data.CustomID)

	if err != nil {
		return err
	}

	handler, ok := router.handlers[route.key()]

	if !ok {
		return fmt.Errorf("%w: %s", ErrComponentRouteNotFound, route.String())
	}

	err = handler(event, route)

	if err != nil {
		return fmt.Errorf("handle component route %s: %w", route.String(), err)
	}

	return nil
}

func parseComponentRoute(customID string) (componentRoute, error) {
	customID = strings.TrimSpace(customID)

	if customID == "" {
		return componentRoute{}, fmt.Errorf("%w: custom id is required", ErrComponentIDInvalid)
	}

	if len(customID) > maxComponentCustomIDLength {
		return componentRoute{}, fmt.Errorf("%w: custom id is too long", ErrComponentIDInvalid)
	}

	parts := strings.Split(customID, ":")

	if len(parts) != 3 {
		return componentRoute{}, fmt.Errorf("%w: custom id must use namespace:action:resource", ErrComponentIDInvalid)
	}

	route := componentRoute{
		Namespace: strings.TrimSpace(parts[0]),
		Action:    strings.TrimSpace(parts[1]),
		Resource:  strings.TrimSpace(parts[2]),
	}

	if !validComponentPart(route.Namespace) {
		return componentRoute{}, fmt.Errorf("%w: namespace is invalid", ErrComponentIDInvalid)
	}

	if !validComponentPart(route.Action) {
		return componentRoute{}, fmt.Errorf("%w: action is invalid", ErrComponentIDInvalid)
	}

	if !validComponentPart(route.Resource) {
		return componentRoute{}, fmt.Errorf("%w: resource is invalid", ErrComponentIDInvalid)
	}

	return route, nil
}

func (route componentRoute) String() string {
	return strings.Join([]string{route.Namespace, route.Action, route.Resource}, ":")
}

func (route componentRoute) key() string {
	return strings.Join([]string{
		strings.TrimSpace(route.Namespace),
		strings.TrimSpace(route.Action),
	}, "\x00")
}

func validComponentPart(value string) bool {
	if value == "" {
		return false
	}

	for _, character := range value {
		if character >= 'a' && character <= 'z' {
			continue
		}

		if character >= 'A' && character <= 'Z' {
			continue
		}

		if character >= '0' && character <= '9' {
			continue
		}

		if character == '-' || character == '_' || character == '.' {
			continue
		}

		return false
	}

	return true
}
