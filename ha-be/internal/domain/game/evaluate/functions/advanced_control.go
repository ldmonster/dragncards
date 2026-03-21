package functions

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// OneCardFunction: picks first matching card from cards list by condition.
type OneCardFunction struct{}

func (*OneCardFunction) Name() string { return OneCardFunctionName }
func (*OneCardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(OneCardFunctionName, args, 2); err != nil {
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

func (*ForEachKeyValFunction) Name() string { return ForEachKeyValFunctionName }
func (*ForEachKeyValFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ForEachKeyValFunctionName, args, 4); err != nil {
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

func (*VarFunction) Name() string { return VarFunctionName }
func (*VarFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(VarFunctionName, args, 2); err != nil {
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

func (*PrevFunction) Name() string { return PrevFunctionName }
func (*PrevFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return ctx.Prev, nil
}

// CondFunction: multibranch.
type CondFunction struct{}

func (*CondFunction) Name() string { return CondFunctionName }
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

// EveryFunction returns true if all items in list satisfy predicate expression.
type EveryFunction struct{}

func (*EveryFunction) Name() string { return EveryFunctionName }
func (*EveryFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(EveryFunctionName, args, 2); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("every first arg must be list")
	}
	pred, ok := args[1].([]any)
	if !ok {
		return nil, errors.New("every second arg must be expression")
	}
	for _, item := range list {
		if ctx.Eval == nil {
			return nil, errors.New("evaluation callback missing")
		}
		cond, err := ctx.Eval(nil, append([]any{pred[0], item}, pred[1:]...))
		if err != nil {
			return nil, err
		}
		b, ok := toBool(cond)
		if !ok || !b {
			return false, nil
		}
	}
	return true, nil
}

// WhileFunction loops while condition true.
type WhileFunction struct{}

func (*WhileFunction) Name() string { return WhileFunctionName }
func (*WhileFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(WhileFunctionName, args, 2); err != nil {
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

func (*MoveCardFunction) Name() string { return MoveCardFunctionName }
func (*MoveCardFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(MoveCardFunctionName, args, 4); err != nil {
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
