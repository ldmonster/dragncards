package evaluate_test

import (
	"strings"
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
)

func TestEvaluateExpressionArithmeticAndLogic(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	cases := []struct {
		expr any
		want any
	}{
		{[]any{"add", 3, 5}, 8},
		{[]any{"sub", 10, 4}, 6},
		{[]any{"mul", 2, 3}, 6},
		{[]any{"div", 20, 5}, 4},
		{[]any{"eq", 5, 5}, true},
		{[]any{"neq", 5, 6}, true},
		{[]any{"gt", 5, 3}, true},
		{[]any{"lt", 3, 5}, true},
		{[]any{"gte", 5, 5}, true},
		{[]any{"lte", 4, 4}, true},
		{[]any{"and", true, true}, true},
		{[]any{"or", false, true}, true},
		{[]any{"not", false}, true},
	}

	// move_card semantics test via GameService applyGameAction
	// Prepare a game state with group/stack/card.
	ctx2 := evaluate.NewEvalContext(game.NewGameUI("room-2"))
	ctx2.Game.Cards["c1"] = &game.Card{ID: "c1", StackID: "s1"}
	ctx2.Game.Stacks["s1"] = game.NewStack("s1", "g1")
	ctx2.Game.Stacks["s1"].AddCard("c1")
	ctx2.Game.Groups["g1"] = game.NewGroup("g1", "g1", "")
	ctx2.Game.Groups["g1"].AddStack("s1")
	ctx2.Game.Stacks["s2"] = game.NewStack("s2", "g1")
	ctx2.Game.Groups["g1"].AddStack("s2")

	_, err := evaluate.EvaluateExpression(ctx2, nil, []any{"move_card", "c1", "g1", 1, 0})
	if err != nil {
		t.Fatalf("move_card failed: %v", err)
	}
	if ctx2.Game.Stacks["s1"].TopCardID() != "" {
		t.Fatalf("expected c1 removed from s1")
	}
	if ctx2.Game.Stacks["s2"].TopCardID() != "c1" {
		t.Fatalf("expected c1 added to s2")
	}

	for _, c := range cases {
		res, err := evaluate.EvaluateExpression(ctx, nil, c.expr)
		if err != nil {
			t.Fatalf("expr %v failed: %v", c.expr, err)
		}
		if res != c.want {
			t.Fatalf("expr %v got %v, want %v", c.expr, res, c.want)
		}
	}
}

func TestEvaluateExpressionStringOps(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{"join_string", "hello", "world"})
	if err != nil {
		t.Fatalf("join_string failed: %v", err)
	}
	if res != "helloworld" {
		t.Fatalf("join_string got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"in_string", "helloworld", "hello"})
	if err != nil {
		t.Fatalf("in_string failed: %v", err)
	}
	if res != true {
		t.Fatalf("in_string got %v", res)
	}
}

func TestEvaluateExpressionMapFilter(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{"map", []any{"list", 1, 2, 3}, []any{"add", 1}})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if got, ok := res.([]any); !ok || len(got) != 3 || got[0] != 2 || got[2] != 4 {
		t.Fatalf("map got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"filter", []any{"list", 1, 2, 3, 4}, []any{"gt", 2}})
	if err != nil {
		t.Fatalf("filter failed: %v", err)
	}
	if got, ok := res.([]any); !ok || len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("filter got %v", res)
	}

	obj := map[string]any{"a": map[string]any{"b": 42}, "arr": []any{10, 20, 30}}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"obj_get_by_path", obj, []any{"a", "b"}})
	if err != nil {
		t.Fatalf("obj_get_by_path failed: %v", err)
	}
	if res != 42 {
		t.Fatalf("obj_get_by_path got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"obj_get_val", obj, []any{"arr", "2"}})
	if err != nil {
		t.Fatalf("obj_get_val failed: %v", err)
	}
	if res != 30 {
		t.Fatalf("obj_get_val got %v", res)
	}

	// new collection op tests
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"set", map[string]any{"x": 1}, "x", 2})
	if err != nil || res.(map[string]any)["x"] != 2 {
		t.Fatalf("set failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"delete", map[string]any{"x": 1}, "x"})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, ok := res.(map[string]any)["x"]; ok {
		t.Fatalf("delete left x")
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"take", []any{"list", 1, 2, 3}, 2})
	if err != nil || len(res.([]any)) != 2 {
		t.Fatalf("take failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"drop", []any{"list", 1, 2, 3}, 2})
	if err != nil || len(res.([]any)) != 1 {
		t.Fatalf("drop failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"count", []any{"list", 1, 2, 3}})
	if err != nil || res != 3 {
		t.Fatalf("count failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"sort", []any{"list", 3, 1, 2}})
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if len(res.([]any)) != 3 || res.([]any)[0] != 1 {
		t.Fatalf("sort result wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"group_by", []any{"list", map[string]any{"a":1}, map[string]any{"a":2}, map[string]any{"a":1}}, "a"})
	if err != nil {
		t.Fatalf("group_by failed: %v", err)
	}
	if groups, ok := res.(map[string]any); !ok || len(groups) != 2 {
		t.Fatalf("group_by wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"rand_between", 1, 5})
	if err != nil {
		t.Fatalf("rand_between failed: %v", err)
	}
	if rn, ok := res.(int); !ok || rn < 1 || rn >= 5 {
		t.Fatalf("rand_between value wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"match", []any{map[string]any{"id": "x"}, map[string]any{"id": "y"}}, "id", "y"})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if m, ok := res.(map[string]any); !ok || m["id"] != "y" {
		t.Fatalf("match result wrong: %v", res)
	}

	// event transaction semantics: draw from stack and update game state
	stack := []any{"c1", "c2", "c3", "c4"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"draw", stack, 2})
	if err != nil {
		t.Fatalf("draw failed: %v", err)
	}
	drawMap, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("draw returned wrong type: %T", res)
	}
	if len(drawMap["drawn"].([]any)) != 2 || len(drawMap["remaining"].([]any)) != 2 {
		t.Fatalf("draw semantics wrong: %v", drawMap)
	}
	// apply to state via obj_set_by_path on stack map
	state := map[string]any{"stack": stack}
	_, err = evaluate.EvaluateExpression(ctx, nil, []any{"obj_set_by_path", state, []any{"stack"}, drawMap["remaining"]})
	if err != nil {
		t.Fatalf("obj_set_by_path stack update failed: %v", err)
	}
	if s, _ := state["stack"].([]any); len(s) != 2 {
		t.Fatalf("stack update wrong: %v", s)
	}

	// discard + shuffle on deck operations
	deck := []any{"c1", "c2", "c3", "c4"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"discard", deck, "c2"})
	if err != nil {
		t.Fatalf("discard failed: %v", err)
	}
	if len(res.([]any)) != 3 {
		t.Fatalf("discard result wrong: %v", res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"shuffle", res})
	if err != nil {
		t.Fatalf("shuffle failed: %v", err)
	}
	if len(res.([]any)) != 3 {
		t.Fatalf("shuffle result wrong: %v", res)
	}
}


func TestEvaluateExpressionDSLHelpers(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{"noop"})
	if err != nil {
		t.Fatalf("noop failed: %v", err)
	}
	if res != nil {
		t.Fatalf("noop expected nil got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"list", 1, 2, 3})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if list, ok := res.([]any); !ok || len(list) != 3 || list[0] != 1 || list[2] != 3 {
		t.Fatalf("list got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"length", []any{"list", 1, 2, 3}})
	if err != nil {
		t.Fatalf("length list failed: %v", err)
	}
	if res != 3 {
		t.Fatalf("length list got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"length", "abcd"})
	if err != nil {
		t.Fatalf("length string failed: %v", err)
	}
	if res != 4 {
		t.Fatalf("length string got %v", res)
	}

	c1 := &game.Card{ID: "c1", StackID: "s1"}
	c2 := &game.Card{ID: "c2", StackID: "s1"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"one_card", []any{"list", c1, c2}, []any{"eq", "$CARD.id", "c1"}})
	if err != nil {
		t.Fatalf("one_card failed: %v", err)
	}
	if card, ok := res.(*game.Card); !ok || card == nil || card.ID != "c1" {
		t.Fatalf("one_card got %v", res)
	}

	// for_each_key_val sets vars and executes body for each map entry
	ctx = evaluate.NewEvalContext(game.NewGameUI("room-1"))
	obj := map[string]any{"k1": 1, "k2": 2}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"for_each_key_val", "key", "val", obj, []any{"noop"}})
	if err != nil {
		t.Fatalf("for_each_key_val failed: %v", err)
	}
	if res != nil {
		t.Fatalf("for_each_key_val expected nil got %v", res)
	}

	ctx = evaluate.NewEvalContext(game.NewGameUI("room-1"))
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"var", "x", 5})
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
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"prev"})
	if err != nil {
		t.Fatalf("prev failed: %v", err)
	}
	if res != "hello" {
		t.Fatalf("prev got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"cond", []any{"eq", 1, 2}, "no", "yes"})
	if err != nil {
		t.Fatalf("cond failed: %v", err)
	}
	if res != "yes" {
		t.Fatalf("cond got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"while", []any{"eq", 0, 0}, []any{"noop"}})
	if err != nil {
		t.Fatalf("while failed: %v", err)
	}
	if res != nil {
		t.Fatalf("while expected nil got %v", res)
	}
}

func TestEvaluateExpressionErrorBehaviorAndExtendedOps(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	_, err := evaluate.EvaluateExpression(ctx, nil, []any{"map", []any{"list", 1, 2, 3}})
	if err == nil || !strings.Contains(err.Error(), "map requires 2 args") {
		t.Fatalf("expected requireArgCount error for map got: %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{"reduce", []any{"list", 1, 2, 3}, 0})
	if err == nil || !strings.Contains(err.Error(), "reduce requires 3 args") {
		t.Fatalf("expected requireArgCount error for reduce got: %v", err)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{"rand", 1, 2, 3})
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

	// obj_set_by_path and set
	obj := map[string]any{"a": map[string]any{"b": 1}}
	res, err := evaluate.EvaluateExpression(ctx, nil, []any{"obj_set_by_path", obj, []any{"a", "b"}, 42})
	if err != nil {
		t.Fatalf("obj_set_by_path failed: %v", err)
	}
	if o, ok := res.(map[string]any); !ok || o["a"].(map[string]any)["b"] != 42 {
		t.Fatalf("obj_set_by_path got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"set", map[string]any{"x": 1}, "x", 9})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if m, ok := res.(map[string]any); !ok || m["x"] != 9 {
		t.Fatalf("set got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"reduce", []any{"list", 1, 2, 3}, 0, []any{"add"}})
	if err != nil {
		t.Fatalf("reduce failed: %v", err)
	}
	if res != 6 {
		t.Fatalf("reduce got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"rand"})
	if err != nil {
		t.Fatalf("rand failed: %v", err)
	}
	if _, ok := res.(float64); !ok {
		t.Fatalf("rand expected float64 got %T", res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"rand", 5})
	if err != nil {
		t.Fatalf("rand(5) failed: %v", err)
	}
	if r, ok := res.(int); !ok || r < 0 || r >= 5 {
		t.Fatalf("rand(5) got %v", res)
	}

	ctx.Vars = map[string]any{"plugin_cards": map[string]map[string]any{"p1": {"c1": map[string]any{"id": "c1", "name": "PluginCard1"}}}}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{"plugin_card", "p1", "c1"})
	if err != nil {
		t.Fatalf("plugin_card failed: %v", err)
	}
	if card, ok := res.(map[string]any); !ok || card["id"] != "c1" {
		t.Fatalf("plugin_card got %v", res)
	}
}
