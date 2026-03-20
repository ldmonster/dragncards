package functions

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// FindInFunction searches for first element satisfying predicate expression.
// Usage: ["find_in", ["list", ...], ["eq", ...]]
type FindInFunction struct{}

func (*FindInFunction) Name() string { return "find_in" }
func (*FindInFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("find_in", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("find_in first arg must be list")
	}
	pred, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("find_in second arg must be expression")
	}
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		res, err := ctx.Eval(nil, append([]any{pred[0], item}, pred[1:]...))
		if err != nil {
			return nil, err
		}
		if b, ok := toBool(res); ok && b {
			return item, nil
		}
	}
	return nil, nil
}

// ShuffleFunction returns a shuffled copy of list.
type ShuffleFunction struct{}

func (*ShuffleFunction) Name() string { return "shuffle" }
func (*ShuffleFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("shuffle", args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("shuffle arg must be list")
	}
	out := make([]any, len(list))
	copy(out, list)
	for i := len(out) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// DrawFunction takes first N from list and returns {drawn, remaining}.
type DrawFunction struct{}

func (*DrawFunction) Name() string { return "draw" }
func (*DrawFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("draw", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("draw first arg must be list")
	}
	n, ok := asInt(args[1])
	if !ok || n < 0 {
		return nil, errors.New("draw second arg must be non-negative int")
	}
	if n > len(list) {
		n = len(list)
	}
	drawn := make([]any, n)
	copy(drawn, list[:n])
	remaining := make([]any, len(list)-n)
	copy(remaining, list[n:])
	return map[string]any{"drawn": drawn, "remaining": remaining}, nil
}

// DiscardFunction removes first occurrence of value from list and returns new list.
type DiscardFunction struct{}

func (*DiscardFunction) Name() string { return "discard" }
func (*DiscardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("discard", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("discard first arg must be list")
	}
	needle := args[1]
	out := make([]any, 0, len(list))
	removed := false
	for _, item := range list {
		if !removed && item == needle {
			removed = true
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// DeleteFunction removes key from map and returns it.
type DeleteFunction struct{}

func (*DeleteFunction) Name() string { return "delete" }
func (*DeleteFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("delete", args, 2); err != nil {
		return nil, err
	}
	obj, ok := args[0].(map[string]any)
	if !ok {
		return nil, errors.New("delete first arg must be map")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, errors.New("delete second arg must be string")
	}
	delete(obj, key)
	return obj, nil
}

// TakeFunction returns first n items of list.
type TakeFunction struct{}

func (*TakeFunction) Name() string { return "take" }
func (*TakeFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("take", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("take first arg must be list")
	}
	n, ok := asInt(args[1])
	if !ok || n < 0 {
		return nil, errors.New("take second arg must be non-negative int")
	}
	if n > len(list) {
		n = len(list)
	}
	return list[:n], nil
}

// DropFunction drops first n items of list.
type DropFunction struct{}

func (*DropFunction) Name() string { return "drop" }
func (*DropFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("drop", args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("drop first arg must be list")
	}
	n, ok := asInt(args[1])
	if !ok || n < 0 {
		return nil, errors.New("drop second arg must be non-negative int")
	}
	if n > len(list) {
		return []any{}, nil
	}
	return list[n:], nil
}

// CountFunction counts elements in list/map/string.
type CountFunction struct{}

func (*CountFunction) Name() string { return "count" }
func (*CountFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("count", args, 1); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case []any:
		return len(v), nil
	case map[string]any:
		return len(v), nil
	case string:
		return len(v), nil
	default:
		return nil, errors.New("count unsupported type")
	}
}

// SortFunction sorts list of ints or strings.
type SortFunction struct{}

func (*SortFunction) Name() string { return "sort" }
func (*SortFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("sort", args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("sort arg must be list")
	}
	ints := make([]int, 0, len(list))
	strs := make([]string, 0, len(list))
	for _, item := range list {
		switch v := item.(type) {
		case int, int32, int64, float32, float64:
			i, _ := asInt(v)
			ints = append(ints, i)
		case string:
			strs = append(strs, v)
		default:
			return nil, errors.New("sort list type must be all numbers or all strings")
		}
	}
	if len(ints) > 0 {
		for i := 1; i < len(ints); i++ {
			j := i
			for j > 0 && ints[j-1] > ints[j] {
				ints[j-1], ints[j] = ints[j], ints[j-1]
				j--
			}
		}
		out := make([]any, len(ints))
		for i, v := range ints {
			out[i] = v
		}
		return out, nil
	}
	for i := 1; i < len(strs); i++ {
		j := i
		for j > 0 && strs[j-1] > strs[j] {
			strs[j-1], strs[j] = strs[j], strs[j-1]
			j--
		}
	}
	out := make([]any, len(strs))
	for i, v := range strs {
		out[i] = v
	}
	return out, nil
}

// GroupByFunction groups list by key expression.
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
	return rand.Intn(max-min) + min, nil
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
