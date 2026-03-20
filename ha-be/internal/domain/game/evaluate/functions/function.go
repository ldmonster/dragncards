package functions

import "github.com/ldmonster/dragncards/ha-be/internal/domain/game"

// EvalContext provides game state and evaluation helpers available to all DSL functions.
type EvalContext = game.EvalContext

// GameFunction is the interface for DSL functions.
type GameFunction interface {
	Name() string
	Execute(ctx *EvalContext, args []any) (any, error)
}

// GameVariable is the interface for variables exposed to DSL (resolve at runtime).
type GameVariable interface {
	Name() string
	Resolve(ctx *EvalContext) (any, error)
}

// Registrar is implemented by Evaluator to collect functions and variables.
type Registrar interface {
	RegisterFunction(GameFunction)
	RegisterVariable(GameVariable)
}
