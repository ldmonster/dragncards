package evaluate

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

func EvaluateExpression(ctx *game.EvalContext, card *game.Card, expr any) (any, error) {
	if ctx != nil && ctx.Eval == nil {
		ctx.Eval = func(card *game.Card, expr any) (any, error) {
			return EvaluateExpression(ctx, card, expr)
		}
	}
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
		return evaluateArray(ctx, card, v)
	default:
		return v, nil
	}
}

func evaluateArray(ctx *game.EvalContext, card *game.Card, code []any) (any, error) {
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
			v, err := evaluateArray(ctx, card, barr)
			if err != nil {
				return nil, err
			}
			last = v
		}
		return last, nil
	}

	// Operation dispatch.
	cmdRaw := code[0]
	cmd, ok := cmdRaw.(string)
	if !ok {
		return nil, errors.New("first element of expression must be a command string")
	}
	cmd = strings.ToLower(cmd)

	// Convert operands, except for raw-expression DSL callbacks.
	args := make([]any, 0, len(code)-1)
	for idx, part := range code[1:] {
		isRaw := false
		switch cmd {
		case "map", "filter", "one_card", "obj_get_by_path", "obj_get_val", "obj_set_by_path":
			if idx == 1 {
				isRaw = true
			}
		case "reduce":
			if idx == 2 {
				isRaw = true
			}
		case "for_each_key_val":
			if idx == 3 {
				isRaw = true
			}
		case "var":
			if idx == 1 {
				isRaw = true
			}
		case "cond", "while":
			isRaw = true
		}

		if isRaw {
			args = append(args, part)
			continue
		}

		arg, err := EvaluateExpression(ctx, card, part)
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
}

func resolveGamePath(ctx *game.EvalContext, card *game.Card, path string) (any, error) {
	if path == "$GAME" {
		return ctx.Game, nil
	}
	if path == "$ACTIVE_CARD" {
		return card, nil
	}
	if strings.HasPrefix(path, "$GAME.") {
		parts := strings.Split(path[len("$GAME."):], ".")
		return getObjectByPath(ctx, card, ctx.Game, stringSliceToAny(parts))
	}
	if strings.HasPrefix(path, "$CARD.") {
		if card == nil {
			return nil, errors.New("no active card")
		}
		parts := strings.Split(path[len("$CARD."):], ".")
		return getObjectByPath(ctx, card, card, stringSliceToAny(parts))
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

func getObjectByPath(ctx *EvalContext, card *game.Card, obj any, path []any) (any, error) {
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
