package functions

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

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
		j := rng.Intn(i + 1)
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

// ContainsFunction returns true if a list contains an element or a string contains a substring.
type ContainsFunction struct{}

func (*ContainsFunction) Name() string { return ContainsFunctionName }
func (*ContainsFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ContainsFunctionName, args, 2); err != nil {
		return nil, err
	}
	switch list := args[0].(type) {
	case []any:
		needle := args[1]
		for _, item := range list {
			if reflect.DeepEqual(item, needle) {
				return true, nil
			}
		}
		return false, nil
	case string:
		substr, ok := args[1].(string)
		if !ok {
			return nil, errors.New("contains second arg must be string when first arg is string")
		}
		return strings.Contains(list, substr), nil
	default:
		return nil, errors.New("contains first arg must be list or string")
	}
}

// ContainsInListFunction is alias for contains (Elixir compatibility).
type ContainsInListFunction struct{}

func (*ContainsInListFunction) Name() string { return ContainsInListFunctionName }
func (*ContainsInListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&ContainsFunction{}).Execute(ctx, args)
}

// ReverseFunction reverses a list or string.
type ReverseFunction struct{}

func (*ReverseFunction) Name() string { return ReverseFunctionName }
func (*ReverseFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ReverseFunctionName, args, 1); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = v[len(v)-1-i]
		}
		return out, nil
	case string:
		runes := []rune(v)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	default:
		return nil, errors.New("reverse arg must be list or string")
	}
}

// AppendFunction appends value to a list.
type AppendFunction struct{}

func (*AppendFunction) Name() string { return AppendFunctionName }
func (*AppendFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(AppendFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("append first arg must be list")
	}
	return append(list, args[1]), nil
}

// InListFunction returns true if list contains value.
type InListFunction struct{}

func (*InListFunction) Name() string { return InListFunctionName }
func (*InListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(InListFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("in_list first arg must be list")
	}
	needle := args[1]
	for _, item := range list {
		if reflect.DeepEqual(item, needle) {
			return true, nil
		}
	}
	return false, nil
}

// IsInListFunction is alias of in_list (Elixir compatibility).
type IsInListFunction struct{}

func (*IsInListFunction) Name() string { return IsInListFunctionName }
func (*IsInListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&InListFunction{}).Execute(ctx, args)
}

// GetIndexFunction alias of index_of (Elixir compatibility).
type GetIndexFunction struct{}

func (*GetIndexFunction) Name() string { return GetIndexFunctionName }
func (*GetIndexFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&IndexOfFunction{}).Execute(ctx, args)
}

// AtIndexFunction returns element at index or nil.
type AtIndexFunction struct{}

func (*AtIndexFunction) Name() string { return AtIndexFunctionName }
func (*AtIndexFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(AtIndexFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("at_index first arg must be list")
	}
	idx, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("at_index second arg must be int")
	}
	if idx < 0 || idx >= len(list) {
		return nil, nil
	}
	return list[idx], nil
}

// GetStackIdFunction resolves stack ID from groupId and stack index.
type GetStackIdFunction struct{}

func (*GetStackIdFunction) Name() string { return GetStackIdFunctionName }
func (*GetStackIdFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(GetStackIdFunctionName, args, 2); err != nil {
		return nil, err
	}
	if ctx == nil || ctx.Game == nil {
		return nil, errors.New("context Game required")
	}
	groupID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("get_stack_id first arg must be string")
	}
	stackIndex, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("get_stack_id second arg must be int")
	}
	group := ctx.Game.Groups[groupID]
	if group == nil || stackIndex < 0 || stackIndex >= len(group.StackIDs) {
		return nil, nil
	}
	return group.StackIDs[stackIndex], nil
}

// GetCardIdFunction resolves card ID from groupId, stackIndex, cardIndex.
type GetCardIdFunction struct{}

func (*GetCardIdFunction) Name() string { return GetCardIdFunctionName }
func (*GetCardIdFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(GetCardIdFunctionName, args, 3); err != nil {
		return nil, err
	}
	if ctx == nil || ctx.Game == nil {
		return nil, errors.New("context Game required")
	}
	groupID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("get_card_id first arg must be string")
	}
	stackIndex, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("get_card_id second arg must be int")
	}
	cardIndex, ok := asInt(args[2])
	if !ok {
		return nil, errors.New("get_card_id third arg must be int")
	}
	group := ctx.Game.Groups[groupID]
	if group == nil || stackIndex < 0 || stackIndex >= len(group.StackIDs) {
		return nil, nil
	}
	stackID := group.StackIDs[stackIndex]
	stack := ctx.Game.Stacks[stackID]
	if stack == nil || cardIndex < 0 || cardIndex >= len(stack.CardIDs) {
		return nil, nil
	}
	return stack.CardIDs[cardIndex], nil
}

// IndexOfFunction returns index of value in list or -1.
type IndexOfFunction struct{}

func (*IndexOfFunction) Name() string { return IndexOfFunctionName }
func (*IndexOfFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(IndexOfFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("index_of first arg must be list")
	}
	needle := args[1]
	for i, item := range list {
		if reflect.DeepEqual(item, needle) {
			return i, nil
		}
	}
	return -1, nil
}

// MaxFunction returns maximum numeric value in list.
type MaxFunction struct{}

func (*MaxFunction) Name() string { return MaxFunctionName }
func (*MaxFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(MaxFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("max arg must be list")
	}
	if len(list) == 0 {
		return nil, errors.New("max requires non-empty list")
	}
	max, ok := toInt(list[0])
	if !ok {
		return nil, errors.New("max list elements must be numeric")
	}
	for _, item := range list[1:] {
		v, ok := toInt(item)
		if !ok {
			return nil, errors.New("max list elements must be numeric")
		}
		if v > max {
			max = v
		}
	}
	return max, nil
}

// MinFunction returns minimum numeric value in list.
type MinFunction struct{}

func (*MinFunction) Name() string { return MinFunctionName }
func (*MinFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(MinFunctionName, args, 1); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("min arg must be list")
	}
	if len(list) == 0 {
		return nil, errors.New("min requires non-empty list")
	}
	min, ok := toInt(list[0])
	if !ok {
		return nil, errors.New("min list elements must be numeric")
	}
	for _, item := range list[1:] {
		v, ok := toInt(item)
		if !ok {
			return nil, errors.New("min list elements must be numeric")
		}
		if v < min {
			min = v
		}
	}
	return min, nil
}

// ToIntFunction converts input to int if possible.
type ToIntFunction struct{}

func (*ToIntFunction) Name() string { return ToIntFunctionName }
func (*ToIntFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ToIntFunctionName, args, 1); err != nil {
		return nil, err
	}
	if i, ok := toInt(args[0]); ok {
		return i, nil
	}
	if s, ok := args[0].(string); ok {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("to_int string arg must be numeric")
		}
		return v, nil
	}
	return nil, errors.New("to_int arg must be numeric or numeric string")
}

// TypeOfFunction returns the type name of the argument.
type TypeOfFunction struct{}

func (*TypeOfFunction) Name() string { return TypeOfFunctionName }
func (*TypeOfFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(TypeOfFunctionName, args, 1); err != nil {
		return nil, err
	}
	switch args[0].(type) {
	case nil:
		return "nil", nil
	case bool:
		return "bool", nil
	case int, int32, int64, float32, float64:
		return "number", nil
	case string:
		return "string", nil
	case []any:
		return "list", nil
	case map[string]any:
		return "map", nil
	case *game.Card:
		return "card", nil
	default:
		return "unknown", nil
	}
}

// SplitStringFunction splits a string by separator.
type SplitStringFunction struct{}

func (*SplitStringFunction) Name() string { return SplitStringFunctionName }
func (*SplitStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(SplitStringFunctionName, args, 2); err != nil {
		return nil, err
	}
	input, ok := args[0].(string)
	if !ok {
		return nil, errors.New("split_string first arg must be string")
	}
	sep, ok := args[1].(string)
	if !ok {
		return nil, errors.New("split_string second arg must be string")
	}
	parts := strings.Split(input, sep)
	out := make([]any, len(parts))
	for i, p := range parts {
		out[i] = p
	}
	return out, nil
}

// ListInsertAtFunction inserts an element into a list at index.
type ListInsertAtFunction struct{}

func (*ListInsertAtFunction) Name() string { return ListInsertAtFunctionName }
func (*ListInsertAtFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ListInsertAtFunctionName, args, 3); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("list_insert_at first arg must be list")
	}
	idx, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("list_insert_at second arg must be int")
	}
	if idx < 0 || idx > len(list) {
		return nil, errors.New("list_insert_at index out of bounds")
	}
	out := make([]any, 0, len(list)+1)
	out = append(out, list[:idx]...)
	out = append(out, args[2])
	out = append(out, list[idx:]...)
	return out, nil
}

// ListDeleteAtFunction removes element at given index.
type ListDeleteAtFunction struct{}

func (*ListDeleteAtFunction) Name() string { return ListDeleteAtFunctionName }
func (*ListDeleteAtFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ListDeleteAtFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("list_delete_at first arg must be list")
	}
	idx, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("list_delete_at second arg must be int")
	}
	if idx < 0 || idx >= len(list) {
		return nil, errors.New("list_delete_at index out of bounds")
	}
	out := make([]any, 0, len(list)-1)
	out = append(out, list[:idx]...)
	out = append(out, list[idx+1:]...)
	return out, nil
}

// ListReplaceAtFunction replaces element at index.
type ListReplaceAtFunction struct{}

func (*ListReplaceAtFunction) Name() string { return ListReplaceAtFunctionName }
func (*ListReplaceAtFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ListReplaceAtFunctionName, args, 3); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("list_replace_at first arg must be list")
	}
	idx, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("list_replace_at second arg must be int")
	}
	if idx < 0 || idx >= len(list) {
		return nil, errors.New("list_replace_at index out of bounds")
	}
	out := make([]any, len(list))
	copy(out, list)
	out[idx] = args[2]
	return out, nil
}

// RemoveFromListByValueFunction removes first matching element by value.
type RemoveFromListByValueFunction struct{}

func (*RemoveFromListByValueFunction) Name() string { return RemoveFromListByValueFunctionName }
func (*RemoveFromListByValueFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(RemoveFromListByValueFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("remove_from_list_by_value first arg must be list")
	}
	needle := args[1]
	out := make([]any, 0, len(list))
	removed := false
	for _, item := range list {
		if !removed && reflect.DeepEqual(item, needle) {
			removed = true
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// OneCardWithFaceKeyValFunction returns first card in list whose sides[side][key] == value.
type OneCardWithFaceKeyValFunction struct{}

func (*OneCardWithFaceKeyValFunction) Name() string { return OneCardWithFaceKeyValFunctionName }
func (*OneCardWithFaceKeyValFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, errors.New("one_card_with_face_key_val requires 3 or 4 args")
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("one_card_with_face_key_val first arg must be list")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, errors.New("one_card_with_face_key_val second arg must be string")
	}
	val := args[2]
	side := "A"
	if len(args) == 4 {
		s, ok := args[3].(string)
		if !ok {
			return nil, errors.New("one_card_with_face_key_val fourth arg must be string")
		}
		side = s
	}

	for _, item := range list {
		var candidateVal any
		if cardMap, ok := item.(map[string]any); ok {
			sides, ok := cardMap["sides"].(map[string]any)
			if !ok {
				continue
			}
			sideMap, ok := sides[side].(map[string]any)
			if !ok {
				continue
			}
			candidateVal = sideMap[key]
		} else if cardObj, ok := item.(*game.Card); ok {
			switch strings.ToLower(key) {
			case "id":
				candidateVal = cardObj.ID
			case "stackid":
				candidateVal = cardObj.StackID
			default:
				continue
			}
		} else {
			continue
		}
		if reflect.DeepEqual(candidateVal, val) {
			return item, nil
		}
	}
	return nil, nil
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
