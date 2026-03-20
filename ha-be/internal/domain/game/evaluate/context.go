package evaluate

import "github.com/ldmonster/dragncards/ha-be/internal/domain/game"

// EvalContext is an alias for game.EvalContext.
type EvalContext = game.EvalContext

// NewEvalContext creates an EvalContext wrapping a running GameUI.
func NewEvalContext(gameUI *game.GameUI) *EvalContext {
	ctx := game.NewEvalContext(gameUI)
	ctx.Eval = func(card *game.Card, expr any) (any, error) {
		return EvaluateExpression(ctx, card, expr)
	}
	return ctx
}
