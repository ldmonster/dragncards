package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

type AddFunction struct{}

type SubFunction struct{}

type MulFunction struct{}

type DivFunction struct{}

func parseTwoInts(operation string, args []any) (int, int, error) {
	if err := requireArgCount(operation, args, 2); err != nil {
		return 0, 0, err
	}
	x, ok := asInt(args[0])
	if !ok {
		return 0, 0, errors.New(operation + " first argument must be int")
	}
	y, ok := asInt(args[1])
	if !ok {
		return 0, 0, errors.New(operation + " second argument must be int")
	}
	return x, y, nil
}

func (f *AddFunction) Name() string { return AddFunctionName }

func (f *AddFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := parseTwoInts("add", args)
	if err != nil {
		return nil, err
	}
	return x + y, nil
}

func (f *SubFunction) Name() string { return SubFunctionName }

func (f *SubFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := parseTwoInts("sub", args)
	if err != nil {
		return nil, err
	}
	return x - y, nil
}

func (f *MulFunction) Name() string { return MulFunctionName }

func (f *MulFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := parseTwoInts("mul", args)
	if err != nil {
		return nil, err
	}
	return x * y, nil
}

func (f *DivFunction) Name() string { return DivFunctionName }

func (f *DivFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := parseTwoInts("div", args)
	if err != nil {
		return nil, err
	}
	if y == 0 {
		return nil, errors.New("division by zero")
	}
	return x / y, nil
}
