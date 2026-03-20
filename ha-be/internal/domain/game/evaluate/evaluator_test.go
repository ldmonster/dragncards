package evaluate

import (
	"testing"

	domaingame "github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

type addFunction struct{}

func (a *addFunction) Name() string { return "add" }
func (a *addFunction) Execute(ctx *EvalContext, args []any) (any, error) {
	if len(args) != 2 {
		return nil, nil
	}
	x, ok1 := args[0].(int)
	y, ok2 := args[1].(int)
	if !ok1 || !ok2 {
		return nil, nil
	}
	return x + y, nil
}

type versionVar struct{}

func (v *versionVar) Name() string { return "version" }
func (v *versionVar) Resolve(ctx *EvalContext) (any, error) {
	return "game-1", nil
}

func TestEvaluatorBasics(t *testing.T) {
	e := NewEvaluator()
	e.RegisterFunction(&addFunction{})
	e.RegisterVariable(&versionVar{})

	result, err := e.EvalFunction("add", nil, []any{1, 2})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if result != 3 {
		t.Fatalf("unexpected result: %v", result)
	}

	ctx := NewEvalContext(domaingame.NewGameUI("room-1"))
	varValue, err := e.GetVariable("version", ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if varValue != "game-1" {
		t.Fatalf("unexpected variable: %v", varValue)
	}
}
