package functions

import "github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"

// RegisterBuiltins registers all built-in DSL functions into the evaluator.
func RegisterBuiltins(r evaluate.Registrar) {
	r.RegisterFunction(&NoopFunction{})
	r.RegisterFunction(&AddFunction{})
	r.RegisterFunction(&SubFunction{})
	r.RegisterFunction(&MulFunction{})
	r.RegisterFunction(&DivFunction{})
}
