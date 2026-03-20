package evaluate

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

// Evaluator dispatches DSL function calls and variable resolutions.
type Evaluator struct {
	functions map[string]functions.GameFunction
	variables map[string]functions.GameVariable
}

func NewEvaluator() *Evaluator {
	return &Evaluator{functions: map[string]functions.GameFunction{}, variables: map[string]functions.GameVariable{}}
}

var (
	defaultEvaluator     *Evaluator
	defaultEvaluatorOnce sync.Once
)

func GetDefaultEvaluator() *Evaluator {
	defaultEvaluatorOnce.Do(func() {
		defaultEvaluator = NewEvaluator()
		functions.RegisterBuiltins(defaultEvaluator)
	})
	return defaultEvaluator
}

func (e *Evaluator) RegisterFunction(fn functions.GameFunction) {
	if fn == nil || fn.Name() == "" {
		return
	}
	e.functions[fn.Name()] = fn
}

func (e *Evaluator) RegisterVariable(v functions.GameVariable) {
	if v == nil || v.Name() == "" {
		return
	}
	e.variables[v.Name()] = v
}

func (e *Evaluator) EvalFunction(name string, ctx *game.EvalContext, args []any) (any, error) {
	fn, ok := e.functions[name]
	if !ok {
		return nil, errors.New("function not found")
	}
	return fn.Execute(ctx, args)
}

func (e *Evaluator) GetVariable(name string, ctx *game.EvalContext) (any, error) {
	v, ok := e.variables[name]
	if !ok {
		return nil, errors.New("variable not found")
	}
	return v.Resolve(ctx)
}
