package functions

import (
	"errors"
	"fmt"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

func toInt(v any) (int, bool) {
	switch i := v.(type) {
	case int:
		return i, true
	case int32:
		return int(i), true
	case int64:
		return int(i), true
	case float32:
		return int(i), true
	case float64:
		return int(i), true
	case bool:
		if i {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func toBool(v any) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	case int:
		return b != 0, true
	case int32:
		return b != 0, true
	case int64:
		return b != 0, true
	case float32:
		return b != 0, true
	case float64:
		return b != 0, true
	default:
		return false, false
	}
}

func requireArgCount(name string, args []any, expected int) error {
	if len(args) != expected {
		return errors.New(name + " requires " + fmt.Sprint(expected) + " args")
	}
	return nil
}

func requireStringArg(name string, args []any, index int) (string, error) {
	if index < 0 || index >= len(args) {
		return "", errors.New(name + " argument index out of bounds")
	}
	v, ok := args[index].(string)
	if !ok {
		return "", errors.New(name + " argument " + fmt.Sprint(index+1) + " must be string")
	}
	return v, nil
}

// ListFunction returns the arguments as a list.
type ListFunction struct{}

func (*ListFunction) Name() string { return ListFunctionName }
func (*ListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return args, nil
}

// LengthFunction returns size of string/list/map.
type LengthFunction struct{}

func (*LengthFunction) Name() string { return LengthFunctionName }
func (*LengthFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(LengthFunctionName, args, 1); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case string:
		return len(v), nil
	case []any:
		return len(v), nil
	case map[string]any:
		return len(v), nil
	default:
		return nil, errors.New("length unsupported type")
	}
}

// MapFunction applies a predicate to each item in a list.
type MapFunction struct{}

func (*MapFunction) Name() string { return MapFunctionName }
func (*MapFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(MapFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("map first arg must be list")
	}
	pred, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("map second arg must be expression")
	}
	out := make([]any, 0, len(list))
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		val, err := ctx.Eval(nil, append([]any{pred[0], item}, pred[1:]...))
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}

// FilterFunction filters list items by predicate result.
type FilterFunction struct{}

func (*FilterFunction) Name() string { return FilterFunctionName }
func (*FilterFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(FilterFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("filter first arg must be list")
	}
	pred, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("filter second arg must be expression")
	}
	out := make([]any, 0, len(list))
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		cond, err := ctx.Eval(nil, append([]any{pred[0], item}, pred[1:]...))
		if err != nil {
			return nil, err
		}
		if b, ok := toBool(cond); ok && b {
			out = append(out, item)
		}
	}
	return out, nil
}
