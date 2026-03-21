package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

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
