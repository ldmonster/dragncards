package functions

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// ConcatFunction joins two strings.
type ConcatFunction struct{}

func (*ConcatFunction) Name() string { return "concat" }
func (*ConcatFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("concat", args, 2); err != nil {
		return nil, err
	}
	a, err := requireStringArg("concat", args, 0)
	if err != nil {
		return nil, err
	}
	b, err := requireStringArg("concat", args, 1)
	if err != nil {
		return nil, err
	}
	return a + b, nil
}

// JoinStringFunction is an alias of concat.
type JoinStringFunction struct{}

func (*JoinStringFunction) Name() string { return "join_string" }
func (*JoinStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("join_string", args, 2); err != nil {
		return nil, err
	}
	a, err := requireStringArg("join_string", args, 0)
	if err != nil {
		return nil, err
	}
	b, err := requireStringArg("join_string", args, 1)
	if err != nil {
		return nil, err
	}
	return a + b, nil
}

// InStringFunction (case-sensitive contains).
type InStringFunction struct{}

func (*InStringFunction) Name() string { return "in_string" }
func (*InStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("in_string", args, 2); err != nil {
		return nil, err
	}
	L, err := requireStringArg("in_string", args, 0)
	if err != nil {
		return nil, err
	}
	R, err := requireStringArg("in_string", args, 1)
	if err != nil {
		return nil, err
	}
	return strings.Contains(L, R), nil
}

func resolveByPath(obj any, path []any) (any, error) {
	current := obj
	for _, segment := range path {
		s, ok := segment.(string)
		if !ok {
			return nil, errors.New("path segment must be string")
		}
		switch cur := current.(type) {
		case map[string]any:
			v, exists := cur[s]
			if !exists {
				return nil, nil
			}
			current = v
		case []any:
			idx, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.New("array index must be integer string")
			}
			if idx < 0 || idx >= len(cur) {
				return nil, nil
			}
			current = cur[idx]
		default:
			return nil, errors.New("unsupported object type for path")
		}
	}
	return current, nil
}

// ObjGetByPathFunction returns object field by path.
type ObjGetByPathFunction struct{}

func (*ObjGetByPathFunction) Name() string { return "obj_get_by_path" }
func (*ObjGetByPathFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("obj_get_by_path", args, 2); err != nil {
		return nil, err
	}
	obj := args[0]
	path, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("obj_get_by_path path argument must be list")
	}
	return resolveByPath(obj, path)
}

// ObjGetValFunction alias of obj_get_by_path.
type ObjGetValFunction struct{}

func (*ObjGetValFunction) Name() string { return "obj_get_val" }
func (*ObjGetValFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&ObjGetByPathFunction{}).Execute(ctx, args)
}

func setByPath(obj any, path []any, value any) (any, error) {
	if len(path) == 0 {
		return nil, errors.New("path cannot be empty")
	}
	current := obj
	for i, segment := range path {
		s, ok := segment.(string)
		if !ok {
			return nil, errors.New("path segment must be string")
		}
		if i == len(path)-1 {
			// final step, assign value in map or list.
			switch c := current.(type) {
			case map[string]any:
				c[s] = value
				return obj, nil
			case []any:
				idx, err := strconv.Atoi(s)
				if err != nil {
					return nil, errors.New("array index must be integer string")
				}
				if idx < 0 || idx >= len(c) {
					return nil, errors.New("array index out of bounds")
				}
				c[idx] = value
				return obj, nil
			default:
				return nil, errors.New("unsupported object type for set")
			}
		}
		// traverse into existing child, create if missing for map.
		switch c := current.(type) {
		case map[string]any:
			child, exists := c[s]
			if !exists {
				c[s] = map[string]any{}
				child = c[s]
			}
			current = child
		case []any:
			idx, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.New("array index must be integer string")
			}
			if idx < 0 || idx >= len(c) {
				return nil, errors.New("array index out of bounds")
			}
			current = c[idx]
		default:
			return nil, errors.New("unsupported object type for path")
		}
	}
	return nil, errors.New("unreachable")
}

// ObjSetByPathFunction sets object field by path (map or array indices).
type ObjSetByPathFunction struct{}

func (*ObjSetByPathFunction) Name() string { return "obj_set_by_path" }
func (*ObjSetByPathFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("obj_set_by_path", args, 3); err != nil {
		return nil, err
	}
	obj := args[0]
	path, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("obj_set_by_path path argument must be list")
	}
	return setByPath(obj, path, args[2])
}

// SetFunction adds or replaces key in map-style object.
type SetFunction struct{}

func (*SetFunction) Name() string { return "set" }
func (*SetFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("set", args, 3); err != nil {
		return nil, err
	}
	obj, ok := args[0].(map[string]any)
	if !ok {
		return nil, errors.New("set first arg must be map")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, errors.New("set second arg must be string")
	}
	obj[key] = args[2]
	return obj, nil
}

// ReduceFunction reduces list with accumulator and binary expression.
type ReduceFunction struct{}

func (*ReduceFunction) Name() string { return "reduce" }
func (*ReduceFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("reduce", args, 3); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("reduce first arg must be list")
	}
	acc := args[1]
	pred, ok := args[2].([]any)
	if !ok {
		return nil, errors.New("reduce third arg must be expression")
	}
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		if len(pred) == 0 {
			return nil, errors.New("reduce predicate expression is empty")
		}
		expr := append([]any{pred[0], acc, item}, pred[1:]...)
		res, err := ctx.Eval(nil, expr)
		if err != nil {
			return nil, err
		}
		acc = res
	}
	return acc, nil
}

// RandFunction produces random numeric values.
type RandFunction struct{}

func (*RandFunction) Name() string { return "rand" }
func (*RandFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	switch len(args) {
	case 0:
		return rand.Float64(), nil
	case 1:
		max, ok := asInt(args[0])
		if !ok || max <= 0 {
			return nil, errors.New("rand argument must be positive int")
		}
		return rand.Intn(max), nil
	case 2:
		min, ok1 := asInt(args[0])
		max, ok2 := asInt(args[1])
		if !ok1 || !ok2 {
			return nil, errors.New("rand arguments must be ints")
		}
		if min >= max {
			return nil, errors.New("rand min must be less than max")
		}
		return min + rand.Intn(max-min), nil
	default:
		return nil, errors.New("rand accepts 0, 1 or 2 args")
	}
}
