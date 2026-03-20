package evaluate

import "github.com/ldmonster/dragncards/ha-be/internal/domain/game"

// EvalContext is passed to every DSL function and variable resolver.
// It gives typed, nil-safe access to the full game state.
type EvalContext struct {
	Game *game.GameUI
}

// NewEvalContext creates an EvalContext wrapping a running GameUI.
func NewEvalContext(gameUI *game.GameUI) *EvalContext {
	return &EvalContext{Game: gameUI}
}

// Card returns the Card with the given ID, or nil.
func (c *EvalContext) Card(id string) *game.Card {
	if c.Game == nil || c.Game.Cards == nil {
		return nil
	}
	return c.Game.Cards[id]
}

// Stack returns the Stack with the given ID, or nil.
func (c *EvalContext) Stack(id string) *game.Stack {
	if c.Game == nil || c.Game.Stacks == nil {
		return nil
	}
	return c.Game.Stacks[id]
}

// Group returns the Group with the given ID, or nil.
func (c *EvalContext) Group(id string) *game.Group {
	if c.Game == nil || c.Game.Groups == nil {
		return nil
	}
	return c.Game.Groups[id]
}

// PlayerInfo returns the PlayerInfo for the given player ID, or nil.
func (c *EvalContext) PlayerInfo(id string) *game.PlayerInfo {
	if c.Game == nil || c.Game.Infos == nil {
		return nil
	}
	return c.Game.Infos[id]
}
