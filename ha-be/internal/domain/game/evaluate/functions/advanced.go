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

func (*ListFunction) Name() string { return "list" }
func (*ListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return args, nil
}

// LengthFunction returns size of string/list/map.
type LengthFunction struct{}

func (*LengthFunction) Name() string { return "length" }
func (*LengthFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("length", args, 1); err != nil {
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

func (*MapFunction) Name() string { return "map" }
func (*MapFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("map", args, 2); err != nil {
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

func (*FilterFunction) Name() string { return "filter" }
func (*FilterFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("filter", args, 2); err != nil {
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

// OneCardFunction: picks first matching card from cards list by condition.
type OneCardFunction struct{}

func (*OneCardFunction) Name() string { return "one_card" }
func (*OneCardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("one_card", args, 2); err != nil {
		return nil, err
	}
	cards, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("one_card cards must be list")
	}
	cond, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("one_card condition must be expression")
	}
	for _, c := range cards {
		cardObj, ok := c.(*game.Card)
		if !ok {
			continue
		}
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		out, err := ctx.Eval(cardObj, cond)
		if err != nil {
			return nil, err
		}
		if b, ok := toBool(out); ok && b {
			return cardObj, nil
		}
	}
	return nil, nil
}

// ForEachKeyValFunction: loops over map values with vars.
type ForEachKeyValFunction struct{}

func (*ForEachKeyValFunction) Name() string { return "for_each_key_val" }
func (*ForEachKeyValFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("for_each_key_val", args, 4); err != nil {
		return nil, err
	}
	keyVar, ok1 := args[0].(string)
	valVar, ok2 := args[1].(string)
	obj, ok3 := args[2].(map[string]any)
	body, ok4 := args[3].([]any)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, errors.New("for_each_key_val invalid args")
	}
	if ctx.Vars == nil {
		ctx.Vars = map[string]any{}
	}
	for k, v := range obj {
		ctx.Vars[keyVar] = k
		ctx.Vars[valVar] = v
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		_, err := ctx.Eval(nil, body)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// VarFunction stores and returns variable.
type VarFunction struct{}

func (*VarFunction) Name() string { return "var" }
func (*VarFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("var", args, 2); err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, errors.New("var name must be string")
	}
	if ctx.Eval == nil {
		return nil, errors.New("evaluation callback missing")
	}
	value, err := ctx.Eval(nil, args[1])
	if err != nil {
		return nil, err
	}
	if ctx.Vars == nil {
		ctx.Vars = map[string]any{}
	}
	ctx.Vars[name] = value
	return value, nil
}

// PrevFunction returns previous evaluation result.
type PrevFunction struct{}

func (*PrevFunction) Name() string { return "prev" }
func (*PrevFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return ctx.Prev, nil
}

// CondFunction: multibranch.
type CondFunction struct{}

func (*CondFunction) Name() string { return "cond" }
func (*CondFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	for i := 0; i+1 < len(args); i += 2 {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		pred, err := ctx.Eval(nil, args[i])
		if err != nil {
			return nil, err
		}
		if b, ok := toBool(pred); ok && b {
			return ctx.Eval(nil, args[i+1])
		}
	}
	if len(args)%2 == 1 {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		return ctx.Eval(nil, args[len(args)-1])
	}
	return nil, nil
}

// WhileFunction loops while condition true.
type WhileFunction struct{}

func (*WhileFunction) Name() string { return "while" }
func (*WhileFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("while", args, 2); err != nil {
		return nil, err
	}
	cond, ok1 := args[0].([]any)
	body, ok2 := args[1].([]any)
	if !ok1 || !ok2 {
		return nil, errors.New("while args must be expressions")
	}
	var last any
	for i := 0; i < 256; i++ {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		c, err := ctx.Eval(nil, cond)
		if err != nil {
			return nil, err
		}
		if b, ok := toBool(c); !ok || !b {
			break
		}
		last, err = ctx.Eval(nil, body)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

// MoveCardFunction moves card in game state.
type MoveCardFunction struct{}

func (*MoveCardFunction) Name() string { return "move_card" }
func (*MoveCardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount("move_card", args, 4); err != nil {
		return nil, err
	}
	cardID, ok := args[0].(string)
	if !ok {
		return nil, errors.New("move_card card id must be string")
	}
	destGroupID, ok := args[1].(string)
	if !ok {
		return nil, errors.New("move_card group id must be string")
	}
	stackIndex, ok := toInt(args[2])
	if !ok {
		return nil, errors.New("move_card stack index must be int")
	}
	cardIndex, ok := toInt(args[3])
	if !ok {
		return nil, errors.New("move_card card index must be int")
	}
	if ctx.Game == nil {
		return nil, errors.New("no game context")
	}
	c := ctx.Game.GetCard(cardID)
	if c == nil {
		return nil, errors.New("card not found")
	}
	if oldStack := ctx.Game.Stacks[c.StackID]; oldStack != nil {
		oldStack.RemoveCard(cardID)
	}
	group := ctx.Game.Groups[destGroupID]
	if group == nil {
		return nil, errors.New("dest group not found")
	}
	if stackIndex < 0 || stackIndex >= len(group.StackIDs) {
		return nil, errors.New("stack index out of bounds")
	}
	destStackID := group.StackIDs[stackIndex]
	destStack := ctx.Game.Stacks[destStackID]
	if destStack == nil {
		return nil, errors.New("dest stack not found")
	}
	if cardIndex < 0 || cardIndex > len(destStack.CardIDs) {
		cardIndex = len(destStack.CardIDs)
	}
	destStack.CardIDs = append(destStack.CardIDs[:cardIndex], append([]string{cardID}, destStack.CardIDs[cardIndex:]...)...)
	c.StackID = destStackID
	ctx.Game.PutCard(c)
	return true, nil
}
