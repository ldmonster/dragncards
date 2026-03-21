package evaluate_test

import (
	"strings"
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

func TestEvaluateExpressionDSLHelpers(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.NoopFunctionName})
	if err != nil {
		t.Fatalf("noop failed: %v", err)
	}
	if res != nil {
		t.Fatalf("noop expected nil got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ListFunctionName, 1, 2, 3})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if list, ok := res.([]any); !ok || len(list) != 3 || list[0] != 1 || list[2] != 3 {
		t.Fatalf("list got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.LengthFunctionName, []any{functions.ListFunctionName, 1, 2, 3}})
	if err != nil {
		t.Fatalf("length list failed: %v", err)
	}
	if res != 3 {
		t.Fatalf("length list got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.LengthFunctionName, "abcd"})
	if err != nil {
		t.Fatalf("length string failed: %v", err)
	}
	if res != 4 {
		t.Fatalf("length string got %v", res)
	}

	c1 := &game.Card{ID: "c1", StackID: "s1"}
	c2 := &game.Card{ID: "c2", StackID: "s1"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.OneCardFunctionName, []any{functions.ListFunctionName, c1, c2}, []any{functions.EqFunctionName, "$CARD.id", "c1"}})
	if err != nil {
		t.Fatalf("one_card failed: %v", err)
	}
	if card, ok := res.(*game.Card); !ok || card == nil || card.ID != "c1" {
		t.Fatalf("one_card got %v", res)
	}

	ctx = evaluate.NewEvalContext(game.NewGameUI("room-1"))
	obj := map[string]any{"k1": 1, "k2": 2}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ForEachKeyValFunctionName, "key", "val", obj, []any{functions.NoopFunctionName}})
	if err != nil {
		t.Fatalf("for_each_key_val failed: %v", err)
	}
	if res != nil {
		t.Fatalf("for_each_key_val expected nil got %v", res)
	}

	ctx = evaluate.NewEvalContext(game.NewGameUI("room-1"))
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.VarFunctionName, "x", 5})
	if err != nil {
		t.Fatalf("var failed: %v", err)
	}
	if res != 5 {
		t.Fatalf("var got %v", res)
	}
	if ctx.Vars["x"] != 5 {
		t.Fatalf("var store with %v", ctx.Vars)
	}

	ctx.Prev = "hello"
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.PrevFunctionName})
	if err != nil {
		t.Fatalf("prev failed: %v", err)
	}
	if res != "hello" {
		t.Fatalf("prev got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.CondFunctionName, []any{functions.EqFunctionName, 1, 2}, "no", "yes"})
	if err != nil {
		t.Fatalf("cond failed: %v", err)
	}
	if res != "yes" {
		t.Fatalf("cond got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.WhileFunctionName, []any{functions.EqFunctionName, 0, 0}, []any{functions.NoopFunctionName}})
	if err != nil {
		t.Fatalf("while failed: %v", err)
	}
	if res != nil {
		t.Fatalf("while expected nil got %v", res)
	}
}

func TestEvaluateExpressionErrorBehaviorAndExtendedOps(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	_, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.MapFunctionName, []any{functions.ListFunctionName, 1, 2, 3}})
	if err == nil || !strings.Contains(err.Error(), "map requires 2 args") {
		t.Fatalf("expected requireArgCount error for map got: %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ReduceFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, 0})
	if err == nil || !strings.Contains(err.Error(), "reduce requires 3 args") {
		t.Fatalf("expected requireArgCount error for reduce got: %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RandFunctionName, 1, 2, 3})
	if err == nil || !strings.Contains(err.Error(), "rand accepts 0, 1 or 2 args") {
		t.Fatalf("expected requireArgCount error for rand got: %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{"does_not_exist", 1})
	if err == nil || !strings.Contains(err.Error(), "function not found") {
		t.Fatalf("expected function not found error got %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{123, 4})
	if err == nil || !strings.Contains(err.Error(), "first element of expression must be a command string") {
		t.Fatalf("expected invalid syntax error got %v", err)
	}

	obj := map[string]any{"a": map[string]any{"b": 1}}
	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.ObjSetByPathFunctionName, obj, []any{"a", "b"}, 42})
	if err != nil {
		t.Fatalf("obj_set_by_path failed: %v", err)
	}
	if o, ok := res.(map[string]any); !ok || o["a"].(map[string]any)["b"] != 42 {
		t.Fatalf("obj_set_by_path got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.SetFunctionName, map[string]any{"x": 1}, "x", 9})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if m, ok := res.(map[string]any); !ok || m["x"] != 9 {
		t.Fatalf("set got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ReduceFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, 0, []any{functions.AddFunctionName}})
	if err != nil {
		t.Fatalf("reduce failed: %v", err)
	}
	if res != 6 {
		t.Fatalf("reduce got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RandFunctionName})
	if err != nil {
		t.Fatalf("rand failed: %v", err)
	}
	if _, ok := res.(float64); !ok {
		t.Fatalf("rand expected float64 got %T", res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RandFunctionName, 5})
	if err != nil {
		t.Fatalf("rand(5) failed: %v", err)
	}
	if r, ok := res.(int); !ok || r < 0 || r >= 5 {
		t.Fatalf("rand(5) got %v", res)
	}

	ctx.Vars = map[string]any{"plugin_cards": map[string]map[string]any{"p1": {"c1": map[string]any{"id": "c1", "name": "PluginCard1"}}}}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.PluginCardFunctionName, "p1", "c1"})
	if err != nil {
		t.Fatalf("plugin_card failed: %v", err)
	}
	if card, ok := res.(map[string]any); !ok || card["id"] != "c1" {
		t.Fatalf("plugin_card got %v", res)
	}
}
