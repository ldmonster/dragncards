package functions

import (
	"errors"
	"math"
	"reflect"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

func containsAny(list []any, value any) bool {
	for _, item := range list {
		if reflect.DeepEqual(item, value) {
			return true
		}
	}
	return false
}

type ModFunction struct{}

func (*ModFunction) Name() string { return ModFunctionName }
func (*ModFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ModFunctionName, args, 2); err != nil {
		return nil, err
	}
	a, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("mod first arg must be int")
	}
	b, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("mod second arg must be int")
	}
	if b == 0 {
		return nil, errors.New("mod by zero")
	}
	return a % b, nil
}

type PowFunction struct{}

func (*PowFunction) Name() string { return PowFunctionName }
func (*PowFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(PowFunctionName, args, 2); err != nil {
		return nil, err
	}
	b, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("pow base must be int")
	}
	e, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("pow exponent must be int")
	}
	return int(math.Pow(float64(b), float64(e))), nil
}

type AbsFunction struct{}

func (*AbsFunction) Name() string { return AbsFunctionName }
func (*AbsFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(AbsFunctionName, args, 1); err != nil {
		return nil, err
	}
	x, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("abs arg must be int")
	}
	if x < 0 {
		x = -x
	}
	return x, nil
}

type FloorFunction struct{}

func (*FloorFunction) Name() string { return FloorFunctionName }
func (*FloorFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(FloorFunctionName, args, 1); err != nil {
		return nil, err
	}
	v := args[0]
	switch x := v.(type) {
	case int, int32, int64:
		i, _ := asInt(x)
		return i, nil
	case float32:
		return int(math.Floor(float64(x))), nil
	case float64:
		return int(math.Floor(x)), nil
	default:
		return nil, errors.New("floor arg must be numeric")
	}
}

type CeilFunction struct{}

func (*CeilFunction) Name() string { return CeilFunctionName }
func (*CeilFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(CeilFunctionName, args, 1); err != nil {
		return nil, err
	}
	v := args[0]
	switch x := v.(type) {
	case int, int32, int64:
		i, _ := asInt(x)
		return i, nil
	case float32:
		return int(math.Ceil(float64(x))), nil
	case float64:
		return int(math.Ceil(x)), nil
	default:
		return nil, errors.New("ceil arg must be numeric")
	}
}

type IsNilFunction struct{}

func (*IsNilFunction) Name() string { return IsNilFunctionName }
func (*IsNilFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IsNilFunctionName, args, 1); err != nil {
		return nil, err
	}
	return args[0] == nil, nil
}

type IsNumberFunction struct{}

func (*IsNumberFunction) Name() string { return IsNumberFunctionName }
func (*IsNumberFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IsNumberFunctionName, args, 1); err != nil {
		return nil, err
	}
	_, ok := asInt(args[0])
	return ok, nil
}

type IsStringFunction struct{}

func (*IsStringFunction) Name() string { return IsStringFunctionName }
func (*IsStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IsStringFunctionName, args, 1); err != nil {
		return nil, err
	}
	_, ok := args[0].(string)
	return ok, nil
}

type IsListFunction struct{}

func (*IsListFunction) Name() string { return IsListFunctionName }
func (*IsListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IsListFunctionName, args, 1); err != nil {
		return nil, err
	}
	_, ok := args[0].([]any)
	return ok, nil
}

type IsMapFunction struct{}

func (*IsMapFunction) Name() string { return IsMapFunctionName }
func (*IsMapFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IsMapFunctionName, args, 1); err != nil {
		return nil, err
	}
	_, ok := args[0].(map[string]any)
	return ok, nil
}

type HeadFunction struct{}

func (*HeadFunction) Name() string { return HeadFunctionName }
func (*HeadFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(HeadFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("head arg must be list")
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

type TailFunction struct{}

func (*TailFunction) Name() string { return TailFunctionName }
func (*TailFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(TailFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("tail arg must be list")
	}
	if len(list) == 0 {
		return []any{}, nil
	}
	return list[1:], nil
}

type FirstFunction struct{}

func (*FirstFunction) Name() string { return FirstFunctionName }
func (*FirstFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&HeadFunction{}).Execute(ctx, args)
}

type LastFunction struct{}

func (*LastFunction) Name() string { return LastFunctionName }
func (*LastFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(LastFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("last arg must be list")
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[len(list)-1], nil
}

type KeysFunction struct{}

func (*KeysFunction) Name() string { return KeysFunctionName }
func (*KeysFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(KeysFunctionName, args, 1); err != nil {
		return nil, err
	}
	m, ok := args[0].(map[string]any)
	if !ok {
		return nil, errors.New("keys arg must be map")
	}
	out := make([]any, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out, nil
}

type ValuesFunction struct{}

func (*ValuesFunction) Name() string { return ValuesFunctionName }
func (*ValuesFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ValuesFunctionName, args, 1); err != nil {
		return nil, err
	}
	m, ok := args[0].(map[string]any)
	if !ok {
		return nil, errors.New("values arg must be map")
	}
	out := make([]any, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out, nil
}

type MergeFunction struct{}

func (*MergeFunction) Name() string { return MergeFunctionName }
func (*MergeFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(MergeFunctionName, args, 2); err != nil {
		return nil, err
	}
	m1, ok1 := args[0].(map[string]any)
	m2, ok2 := args[1].(map[string]any)
	if !ok1 || !ok2 {
		return nil, errors.New("merge args must be maps")
	}
	out := map[string]any{}
	for k, v := range m1 {
		out[k] = v
	}
	for k, v := range m2 {
		out[k] = v
	}
	return out, nil
}

type FlattenFunction struct{}

func (*FlattenFunction) Name() string { return FlattenFunctionName }
func (*FlattenFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(FlattenFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("flatten arg must be list")
	}
	out := []any{}
	for _, item := range list {
		if nested, ok := item.([]any); ok {
			out = append(out, nested...)
		} else {
			out = append(out, item)
		}
	}
	return out, nil
}

type UnionFunction struct{}

func (*UnionFunction) Name() string { return UnionFunctionName }
func (*UnionFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(UnionFunctionName, args, 2); err != nil {
		return nil, err
	}
	l1, ok1 := args[0].([]any)
	l2, ok2 := args[1].([]any)
	if !ok1 || !ok2 {
		return nil, errors.New("union args must be lists")
	}
	out := []any{}
	for _, item := range l1 {
		if !containsAny(out, item) {
			out = append(out, item)
		}
	}
	for _, item := range l2 {
		if !containsAny(out, item) {
			out = append(out, item)
		}
	}
	return out, nil
}

type DifferenceFunction struct{}

func (*DifferenceFunction) Name() string { return DifferenceFunctionName }
func (*DifferenceFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(DifferenceFunctionName, args, 2); err != nil {
		return nil, err
	}
	l1, ok1 := args[0].([]any)
	l2, ok2 := args[1].([]any)
	if !ok1 || !ok2 {
		return nil, errors.New("difference args must be lists")
	}
	out := []any{}
	for _, item := range l1 {
		if !containsAny(l2, item) {
			out = append(out, item)
		}
	}
	return out, nil
}

type UniqueFunction struct{}

func (*UniqueFunction) Name() string { return UniqueFunctionName }
func (*UniqueFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(UniqueFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("unique arg must be list")
	}
	out := []any{}
	for _, item := range list {
		if !containsAny(out, item) {
			out = append(out, item)
		}
	}
	return out, nil
}
