package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

func asBool(val any) (bool, bool) {
	switch v := val.(type) {
	case bool:
		return v, true
	default:
		return false, false
	}
}

func asString(val any) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	default:
		return "", false
	}
}

func tryCompare[T comparable](args []any, conv func(any) (T, bool)) (T, T, bool) {
	if err := requireArgCount("tryCompare", args, 2); err != nil {
		var zeroT T
		return zeroT, zeroT, false
	}
	x, ok := conv(args[0])
	if !ok {
		var zeroT T
		return zeroT, zeroT, false
	}
	y, ok := conv(args[1])
	if !ok {
		var zeroT T
		return zeroT, zeroT, false
	}
	return x, y, true
}

func compareArgs[T comparable](name string, args []any, conv func(any) (T, bool), typeName string) (T, T, error) {
	var zeroT T
	if err := requireArgCount(name, args, 2); err != nil {
		return zeroT, zeroT, err
	}
	x, ok := conv(args[0])
	if !ok {
		return zeroT, zeroT, errors.New(name + " first argument must be " + typeName)
	}
	y, ok := conv(args[1])
	if !ok {
		return zeroT, zeroT, errors.New(name + " second argument must be " + typeName)
	}
	return x, y, nil
}

func compareIntArgs(name string, args []any) (int, int, error) {
	return compareArgs[int](name, args, asInt, "int")
}

func compareBoolArgs(name string, args []any) (bool, bool, error) {
	return compareArgs[bool](name, args, asBool, "bool")
}

func compareBoolArg1(name string, args []any) (bool, error) {
	if err := requireArgCount(name, args, 1); err != nil {
		return false, err
	}
	b, ok := asBool(args[0])
	if !ok {
		return false, errors.New(name + " argument must be bool")
	}
	return b, nil
}

func compareAnything(name string, args []any) (bool, error) {
	if err := requireArgCount(name, args, 2); err != nil {
		return false, err
	}

	if x, y, ok := tryCompare[int](args, asInt); ok {
		return x == y, nil
	}

	if x, y, ok := tryCompare[string](args, asString); ok {
		return x == y, nil
	}

	if x, y, ok := tryCompare[bool](args, asBool); ok {
		return x == y, nil
	}

	return false, errors.New(name + " arguments must be both ints, strings, or bools")
}

// EqFunction tests equality for ints or strings.
type EqFunction struct{}

func (*EqFunction) Name() string { return EqFunctionName }

func (f *EqFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	res, err := compareAnything("eq", args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NeqFunction tests non-equality for ints or strings.
type NeqFunction struct{}

func (*NeqFunction) Name() string { return NeqFunctionName }

func (f *NeqFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	res, err := compareAnything("neq", args)
	if err != nil {
		return nil, err
	}
	return !res, nil
}

// GtFunction tests greater-than for ints.
type GtFunction struct{}

func (*GtFunction) Name() string { return GtFunctionName }

func (f *GtFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := compareIntArgs("gt", args)
	if err != nil {
		return nil, err
	}
	return x > y, nil
}

// GteFunction tests greater-than-or-equal for ints.
type GteFunction struct{}

func (*GteFunction) Name() string { return GteFunctionName }

func (f *GteFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := compareIntArgs("gte", args)
	if err != nil {
		return nil, err
	}
	return x >= y, nil
}

// LtFunction tests less-than for ints.
type LtFunction struct{}

func (*LtFunction) Name() string { return LtFunctionName }

func (f *LtFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := compareIntArgs("lt", args)
	if err != nil {
		return nil, err
	}
	return x < y, nil
}

// LteFunction tests less-than-or-equal for ints.
type LteFunction struct{}

func (*LteFunction) Name() string { return LteFunctionName }

func (f *LteFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	x, y, err := compareIntArgs("lte", args)
	if err != nil {
		return nil, err
	}
	return x <= y, nil
}

// AndFunction performs boolean AND.
type AndFunction struct{}

// OrFunction performs boolean OR.
type OrFunction struct{}

// NotFunction performs boolean NOT.
type NotFunction struct{}

func (f *AndFunction) Name() string { return AndFunctionName }

func (f *AndFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	b1, b2, err := compareBoolArgs("and", args)
	if err != nil {
		return nil, err
	}
	return b1 && b2, nil
}

func (f *OrFunction) Name() string { return OrFunctionName }

func (f *OrFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	b1, b2, err := compareBoolArgs("or", args)
	if err != nil {
		return nil, err
	}
	return b1 || b2, nil
}

func (f *NotFunction) Name() string { return NotFunctionName }

func (f *NotFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	b, err := compareBoolArg1("not", args)
	if err != nil {
		return nil, err
	}
	return !b, nil
}
