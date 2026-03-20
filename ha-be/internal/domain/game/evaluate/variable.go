package evaluate

// GameVariable is the interface every DSL variable resolver must implement.
// Resolve receives a typed *EvalContext giving full access to game state.
type GameVariable interface {
	Name() string
	Resolve(ctx *EvalContext) (any, error)
}
