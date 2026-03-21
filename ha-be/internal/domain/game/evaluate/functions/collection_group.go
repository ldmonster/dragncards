package functions

import (
	"errors"
	"fmt"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

type GroupByFunction struct{}

func (*GroupByFunction) Name() string { return "group_by" }
func (*GroupByFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("group_by", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("group_by first arg must be list")
	}
	out := map[string]any{}

	// second arg may be key name string or expression.
	if keyName, ok := args[1].(string); ok {
		for _, item := range list {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			keyVal, ok := m[keyName]
			if !ok {
				keyVal = ""
			}
			key := fmt.Sprint(keyVal)
			arr, exists := out[key].([]any)
			if !exists {
				arr = []any{}
			}
			arr = append(arr, item)
			out[key] = arr
		}
		return out, nil
	}

	pred, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("group_by second arg must be string or expression")
	}
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		keyVal, err := ctx.Eval(nil, append([]any{pred[0], item}, pred[1:]...))
		if err != nil {
			return nil, err
		}
		key := fmt.Sprint(keyVal)
		arr, exists := out[key].([]any)
		if !exists {
			arr = []any{}
		}
		arr = append(arr, item)
		out[key] = arr
	}
	return out, nil
}

// RandBetweenFunction returns random integer between min and max (exclusive).
type RandBetweenFunction struct{}

func (*RandBetweenFunction) Name() string { return "rand_between" }
func (*RandBetweenFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("rand_between", args, 2); err != nil {
		return nil, err
	}
	min, ok := asInt(args[0])
	if !ok {
		return nil, errors.New("rand_between first arg must be int")
	}
	max, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("rand_between second arg must be int")
	}
	if min >= max {
		return nil, errors.New("rand_between min must be less than max")
	}
	return rng.Intn(max-min) + min, nil
}

// MatchFunction returns first map with key == value.
type MatchFunction struct{}

func (*MatchFunction) Name() string { return "match" }
func (*MatchFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("match", args, 3); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("match first arg must be list")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, errors.New("match second arg must be string")
	}
	value := args[2]
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if m[key] == value {
			return item, nil
		}
	}
	return nil, nil
}

// EnqueueFunction appends item to event queue list.
type EnqueueFunction struct{}

func (*EnqueueFunction) Name() string { return "enqueue" }
func (*EnqueueFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("enqueue", args, 2); err != nil {
		return nil, err
	}
	queue, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("enqueue first arg must be list")
	}
	queue = append(queue, args[1])
	return queue, nil
}

// DequeueFunction removes first item and returns {head, tail}.
type DequeueFunction struct{}

func (*DequeueFunction) Name() string { return "dequeue" }
func (*DequeueFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("dequeue", args, 1); err != nil {
		return nil, err
	}
	queue, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("dequeue arg must be list")
	}
	if len(queue) == 0 {
		return map[string]any{"head": nil, "tail": []any{}}, nil
	}
	return map[string]any{"head": queue[0], "tail": queue[1:]}, nil
}
