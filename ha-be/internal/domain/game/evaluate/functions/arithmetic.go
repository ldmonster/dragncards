package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
)

type AddFunction struct{}

func (f *AddFunction) Name() string { return "add" }

func (f *AddFunction) Execute(ctx *evaluate.EvalContext, args []any) (any, error) {
	if len(args) != 2 {
		return nil, errors.New("add requires 2 arguments")
	}
	x, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("add first argument must be int")
	}
	y, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("add second argument must be int")
	}
	return x + y, nil
}

type SubFunction struct{}

func (f *SubFunction) Name() string { return "sub" }

func (f *SubFunction) Execute(ctx *evaluate.EvalContext, args []any) (any, error) {
	if len(args) != 2 {
		return nil, errors.New("sub requires 2 arguments")
	}
	x, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("sub first argument must be int")
	}
	y, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("sub second argument must be int")
	}
	return x - y, nil
}

type MulFunction struct{}

func (f *MulFunction) Name() string { return "mul" }

func (f *MulFunction) Execute(ctx *evaluate.EvalContext, args []any) (any, error) {
	if len(args) != 2 {
		return nil, errors.New("mul requires 2 arguments")
	}
	x, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("mul first argument must be int")
	}
	y, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("mul second argument must be int")
	}
	return x * y, nil
}

type DivFunction struct{}

func (f *DivFunction) Name() string { return "div" }

func (f *DivFunction) Execute(ctx *evaluate.EvalContext, args []any) (any, error) {
	if len(args) != 2 {
		return nil, errors.New("div requires 2 arguments")
	}
	x, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("div first argument must be int")
	}
	y, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("div second argument must be int")
	}
	if y == 0 {
		return nil, errors.New("division by zero")
	}
	return x / y, nil
}
