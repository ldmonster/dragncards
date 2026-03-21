package evaluate

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

const (
	GamePathLiteral       = "$GAME"
	ActiveCardPathLiteral = "$ACTIVE_CARD"
	GamePathPrefix        = "$GAME."
	CardPathPrefix        = "$CARD."
)

func EvaluateExpression(ctx *game.EvalContext, card *game.Card, expr any) (any, error) {
	if ctx != nil && ctx.Eval == nil {
		ctx.Eval = func(card *game.Card, expr any) (any, error) {
			return EvaluateExpression(ctx, card, expr)
		}
	}
	return evaluateExpression(ctx, card, expr, true)
}

func evaluateExpression(ctx *game.EvalContext, card *game.Card, expr any, isRoot bool) (any, error) {
	// Handle primitive values and variable references.
	switch v := expr.(type) {
	case nil:
		return nil, nil
	case bool, int, int32, int64, float32, float64:
		return v, nil
	case string:
		if strings.HasPrefix(v, "$") {
			return resolveGamePath(ctx, card, v)
		}
		return v, nil
	case []any:
		return evaluateArray(ctx, card, v, isRoot)
	default:
		return v, nil
	}
}

func evaluateArray(ctx *game.EvalContext, card *game.Card, code []any, isRoot bool) (any, error) {
	if len(code) == 0 {
		return nil, nil
	}

	// array as sequence of expressions, if first item is array.
	if _, ok := code[0].([]any); ok {
		var last any
		for _, block := range code {
			barr, ok := block.([]any)
			if !ok {
				return nil, errors.New("invalid nested code block")
			}
			v, err := evaluateArray(ctx, card, barr, false)
			if err != nil {
				return nil, err
			}
			last = v
		}
		return last, nil
	}

	// array literal (non-command) should be treated as a data list.
	switch first := code[0].(type) {
	case string:
		cmd := strings.ToLower(first)
		if !GetDefaultEvaluator().HasFunction(cmd) {
			if isRoot {
				return nil, errors.New("function not found")
			}
			// fallback to array literal for nested unknown-first-string arrays
			results := make([]any, len(code))
			for i, part := range code {
				v, err := evaluateExpression(ctx, card, part, false)
				if err != nil {
					return nil, err
				}
				results[i] = v
			}
			return results, nil
		}
		// command dispatch.
		// Convert operands, except for raw-expression DSL callbacks.
		args := make([]any, 0, len(code)-1)
		for idx, part := range code[1:] {
			isRaw := false
			switch cmd {
			case functions.MapFunctionName, functions.FilterFunctionName, functions.OneCardFunctionName, functions.ObjGetByPathFunctionName, functions.ObjGetValFunctionName, functions.ObjSetByPathFunctionName, functions.EveryFunctionName:
				if idx == 1 {
					isRaw = true
				}
			case functions.ReduceFunctionName:
				if idx == 2 {
					isRaw = true
				}
			case functions.ForEachKeyValFunctionName:
				if idx == 3 {
					isRaw = true
				}
			case functions.VarFunctionName:
				if idx == 1 {
					isRaw = true
				}
			case functions.CondFunctionName, functions.WhileFunctionName:
				isRaw = true
			}

			if isRaw {
				args = append(args, part)
				continue
			}

			arg, err := evaluateExpression(ctx, card, part, false)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}

		// Dispatch to registered function implementations.
		res, err := GetDefaultEvaluator().EvalFunction(cmd, ctx, args)
		if err != nil {
			return nil, NewEvalError(err)
		}
		ctx.Prev = res
		return res, nil
	case map[string]any:
		results := make([]any, len(code))
		for i, part := range code {
			v, err := evaluateExpression(ctx, card, part, false)
			if err != nil {
				return nil, err
			}
			results[i] = v
		}
		return results, nil
	default:
		if !isRoot {
			// treat as array literal, evaluating each element (non-command list)
			results := make([]any, len(code))
			for i, part := range code {
				v, err := evaluateExpression(ctx, card, part, false)
				if err != nil {
					return nil, err
				}
				results[i] = v
			}
			return results, nil
		}
		return nil, errors.New("first element of expression must be a command string")
	}
}

func resolveGamePath(ctx *game.EvalContext, card *game.Card, path string) (any, error) {
	if path == GamePathLiteral {
		return ctx.Game, nil
	}
	if path == ActiveCardPathLiteral {
		return card, nil
	}
	if strings.HasPrefix(path, GamePathPrefix) {
		parts := strings.Split(path[len(GamePathPrefix):], ".")
		return getObjectByPath(ctx.Game, stringSliceToAny(parts))
	}
	if strings.HasPrefix(path, CardPathPrefix) {
		if card == nil {
			return nil, errors.New("no active card")
		}
		parts := strings.Split(path[len("$CARD."):], ".")
		return getObjectByPath(card, stringSliceToAny(parts))
	}
	return path, nil
}

func stringSliceToAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

func getObjectByPath(obj any, path []any) (any, error) {
	current := obj
	for _, step := range path {
		stepStr, ok := step.(string)
		if !ok {
			return nil, errors.New("path element must be string")
		}

		if current == nil {
			return nil, nil
		}

		if lb := strings.TrimSpace(stepStr); strings.HasPrefix(lb, "[") && strings.HasSuffix(lb, "]") {
			idx, err := strconv.Atoi(lb[1 : len(lb)-1])
			if err != nil {
				return nil, err
			}
			switch c := current.(type) {
			case []any:
				if idx >= 0 && idx < len(c) {
					current = c[idx]
					continue
				}
				return nil, nil
			default:
				return nil, errors.New("indexing non-array")
			}
		}

		// navigate structs/maps
		if mm, ok := current.(map[string]any); ok {
			if child, found := mm[stepStr]; found {
				current = child
				continue
			}
			return nil, nil
		}

		switch c := current.(type) {
		case *game.GameUI:
			current = getGameField(c, stepStr)
		case map[string]*game.Card:
			current = c[stepStr]
		case map[string]*game.Stack:
			current = c[stepStr]
		case map[string]*game.Group:
			current = c[stepStr]
		case map[string]*game.PlayerInfo:
			current = c[stepStr]
		case *game.Card:
			current = getCardField(c, stepStr)
		case *game.Stack:
			current = getStackField(c, stepStr)
		case *game.Group:
			current = getGroupField(c, stepStr)
		default:
			// reflect fallback
			val := reflect.ValueOf(current)
			if val.Kind() == reflect.Struct {
				field := val.FieldByNameFunc(func(name string) bool {
					return strings.EqualFold(name, stepStr)
				})
				if field.IsValid() {
					current = field.Interface()
					continue
				}
			}
			return nil, nil
		}
	}
	return current, nil
}

func getGameField(g *game.GameUI, name string) any {
	switch strings.ToLower(name) {
	case "cards", "cardbyid":
		return g.Cards
	case "stacks", "stackbyid":
		return g.Stacks
	case "groups", "groupbyid":
		return g.Groups
	case "infos", "playerinfos":
		return g.Infos
	case "slug":
		return g.Slug
	case "players":
		return g.Players
	}
	return nil
}

func getCardField(c *game.Card, name string) any {
	switch strings.ToLower(name) {
	case "id":
		return c.ID
	case "stackid":
		return c.StackID
	case "sideup", "currentSide":
		return c.SideUp
	case "rotation":
		return c.Rotation
	case "tokens":
		return c.Tokens
	case "exhausted":
		return c.Exhausted
	case "hidden":
		return c.Hidden
	}
	return nil
}

func getStackField(s *game.Stack, name string) any {
	switch strings.ToLower(name) {
	case "id":
		return s.ID
	case "groupid":
		return s.GroupID
	case "cardids", "cards":
		return s.CardIDs
	case "order":
		return s.Order
	}
	return nil
}

func getGroupField(g *game.Group, name string) any {
	switch strings.ToLower(name) {
	case "id":
		return g.ID
	case "label":
		return g.Label
	case "ownerid":
		return g.OwnerID
	case "stackids", "stacks":
		return g.StackIDs
	}
	return nil
}
