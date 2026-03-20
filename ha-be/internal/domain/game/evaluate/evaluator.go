package evaluate

import "errors"

// Evaluator dispatches DSL function calls and variable resolutions.
type Evaluator struct {
	functions map[string]GameFunction
	variables map[string]GameVariable
}

func NewEvaluator() *Evaluator {
	return &Evaluator{functions: map[string]GameFunction{}, variables: map[string]GameVariable{}}
}

func (e *Evaluator) RegisterFunction(fn GameFunction) {
	if fn == nil || fn.Name() == "" {
		return
	}
	e.functions[fn.Name()] = fn
}

func (e *Evaluator) RegisterVariable(v GameVariable) {
	if v == nil || v.Name() == "" {
		return
	}
	e.variables[v.Name()] = v
}

func (e *Evaluator) EvalFunction(name string, ctx *EvalContext, args []any) (any, error) {
	fn, ok := e.functions[name]
	if !ok {
		return nil, errors.New("function not found")
	}
	return fn.Execute(ctx, args)
}

func (e *Evaluator) GetVariable(name string, ctx *EvalContext) (any, error) {
	v, ok := e.variables[name]
	if !ok {
		return nil, errors.New("variable not found")
	}
	return v.Resolve(ctx)
}
