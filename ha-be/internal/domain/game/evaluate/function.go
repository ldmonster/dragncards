package evaluate

// GameFunction is the interface every DSL operation must implement.
// Execute receives a typed *EvalContext giving full access to game state.
type GameFunction interface {
	Name() string
	Execute(ctx *EvalContext, args []any) (any, error)
}

// Registrar is implemented by Evaluator; it allows separate packages (like
// the functions sub-package) to register their implementations without
// importing the concrete Evaluator type.
type Registrar interface {
	RegisterFunction(GameFunction)
	RegisterVariable(GameVariable)
}
