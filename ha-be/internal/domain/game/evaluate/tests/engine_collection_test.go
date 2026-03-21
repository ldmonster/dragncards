package evaluate_test

import (
	"reflect"
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

func TestEvaluateExpressionMapFilter(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.MapFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, []any{functions.AddFunctionName, 1}})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if got, ok := res.([]any); !ok || len(got) != 3 || got[0] != 2 || got[2] != 4 {
		t.Fatalf("map got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.FilterFunctionName, []any{functions.ListFunctionName, 1, 2, 3, 4}, []any{functions.GtFunctionName, 2}})
	if err != nil {
		t.Fatalf("filter failed: %v", err)
	}
	if got, ok := res.([]any); !ok || len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("filter got %v", res)
	}

	obj := map[string]any{"a": map[string]any{"b": 42}, "arr": []any{10, 20, 30}}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ObjGetByPathFunctionName, obj, []any{"a", "b"}})
	if err != nil {
		t.Fatalf("obj_get_by_path failed: %v", err)
	}
	if res != 42 {
		t.Fatalf("obj_get_by_path got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ObjGetValFunctionName, obj, []any{"arr", "2"}})
	if err != nil {
		t.Fatalf("obj_get_val failed: %v", err)
	}
	if res != 30 {
		t.Fatalf("obj_get_val got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.SetFunctionName, map[string]any{"x": 1}, "x", 2})
	if err != nil || res.(map[string]any)["x"] != 2 {
		t.Fatalf("set failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.DeleteFunctionName, map[string]any{"x": 1}, "x"})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, ok := res.(map[string]any)["x"]; ok {
		t.Fatalf("delete left x")
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.TakeFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, 2})
	if err != nil || len(res.([]any)) != 2 {
		t.Fatalf("take failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.DropFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, 2})
	if err != nil || len(res.([]any)) != 1 {
		t.Fatalf("drop failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.CountFunctionName, []any{functions.ListFunctionName, 1, 2, 3}})
	if err != nil || res != 3 {
		t.Fatalf("count failed: %v %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.SortFunctionName, []any{functions.ListFunctionName, 3, 1, 2}})
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if len(res.([]any)) != 3 || res.([]any)[0] != 1 {
		t.Fatalf("sort result wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.GroupByFunctionName, []any{functions.ListFunctionName, map[string]any{"a": 1}, map[string]any{"a": 2}, map[string]any{"a": 1}}, "a"})
	if err != nil {
		t.Fatalf("group_by failed: %v", err)
	}
	if groups, ok := res.(map[string]any); !ok || len(groups) != 2 {
		t.Fatalf("group_by wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RandBetweenFunctionName, 1, 5})
	if err != nil {
		t.Fatalf("rand_between failed: %v", err)
	}
	if rn, ok := res.(int); !ok || rn < 1 || rn >= 5 {
		t.Fatalf("rand_between value wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.MatchFunctionName, []any{map[string]any{"id": "x"}, map[string]any{"id": "y"}}, "id", "y"})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if m, ok := res.(map[string]any); !ok || m["id"] != "y" {
		t.Fatalf("match result wrong: %v", res)
	}

	stack := []any{"c1", "c2", "c3", "c4"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.DrawFunctionName, stack, 2})
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

	state := map[string]any{"stack": stack}
	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ObjSetByPathFunctionName, state, []any{"stack"}, drawMap["remaining"]})
	if err != nil {
		t.Fatalf("obj_set_by_path stack update failed: %v", err)
	}
	if s, _ := state["stack"].([]any); len(s) != 2 {
		t.Fatalf("stack update wrong: %v", s)
	}

	deck := []any{"c1", "c2", "c3", "c4"}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.DiscardFunctionName, deck, "c2"})
	if err != nil {
		t.Fatalf("discard failed: %v", err)
	}
	if len(res.([]any)) != 3 {
		t.Fatalf("discard result wrong: %v", res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ShuffleFunctionName, res})
	if err != nil {
		t.Fatalf("shuffle failed: %v", err)
	}
	if len(res.([]any)) != 3 {
		t.Fatalf("shuffle result wrong: %v", res)
	}
}

func TestEvaluateExpressionAdditionalDSL(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.AppendFunctionName, []any{1, 2}, 3})
	if err != nil || !reflect.DeepEqual(res, []any{1, 2, 3}) {
		t.Fatalf("append failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.InListFunctionName, []any{1, 2, 3}, 2})
	if err != nil || res != true {
		t.Fatalf("in_list failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.IndexOfFunctionName, []any{"a", "b", "c"}, "b"})
	if err != nil || res != 1 {
		t.Fatalf("index_of failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.MaxFunctionName, []any{1, 5, 2}})
	if err != nil || res != 5 {
		t.Fatalf("max failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.MinFunctionName, []any{1, 5, 2}})
	if err != nil || res != 1 {
		t.Fatalf("min failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ToLowercaseFunctionName, "Hello"})
	if err != nil || res != "hello" {
		t.Fatalf("to_lowercase failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ToUppercaseFunctionName, "hello"})
	if err != nil || res != "HELLO" {
		t.Fatalf("to_uppercase failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ToIntFunctionName, "42"})
	if err != nil || res != 42 {
		t.Fatalf("to_int failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.TypeOfFunctionName, 123})
	if err != nil || res != "number" {
		t.Fatalf("type_of failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.SplitStringFunctionName, "a,b,c", ","})
	if err != nil || !reflect.DeepEqual(res, []any{"a", "b", "c"}) {
		t.Fatalf("split_string failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ListInsertAtFunctionName, []any{1, 2, 4}, 2, 3})
	if err != nil || !reflect.DeepEqual(res, []any{1, 2, 3, 4}) {
		t.Fatalf("list_insert_at failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ListDeleteAtFunctionName, []any{1, 2, 3, 4}, 1})
	if err != nil || !reflect.DeepEqual(res, []any{1, 3, 4}) {
		t.Fatalf("list_delete_at failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RemoveFromListByValueFunctionName, []any{"a", "b", "c", "b"}, "b"})
	if err != nil || !reflect.DeepEqual(res, []any{"a", "c", "b"}) {
		t.Fatalf("remove_from_list_by_value failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.FormatFunctionName, "Hello %s %d", "Go", 2026})
	if err != nil || res != "Hello Go 2026" {
		t.Fatalf("format failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RegexReplaceFunctionName, "one two one", "one", "1"})
	if err != nil || res != "1 two 1" {
		t.Fatalf("regex_replace failed: %v got %v", err, res)
	}

	cards := []any{
		map[string]any{"id": "c1", "sides": map[string]any{"A": map[string]any{"name": "Queen"}}},
		map[string]any{"id": "c2", "sides": map[string]any{"A": map[string]any{"name": "King"}}},
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.OneCardWithFaceKeyValFunctionName, cards, "name", "King"})
	if err != nil || res == nil {
		t.Fatalf("one_card_with_face_key_val failed: %v got %v", err, res)
	}
	if card, ok := res.(map[string]any); !ok || card["id"] != "c2" {
		t.Fatalf("one_card_with_face_key_val wrong card: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ListReplaceAtFunctionName, []any{1, 2, 3}, 1, 9})
	if err != nil || !reflect.DeepEqual(res, []any{1, 9, 3}) {
		t.Fatalf("list_replace_at failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ReplaceStringInListFunctionName, []any{"foo", "bar"}, "o", "0"})
	if err != nil || !reflect.DeepEqual(res, []any{"f00", "bar"}) {
		t.Fatalf("replace_string_in_list failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.EveryFunctionName, []any{functions.ListFunctionName, 2, 4, 6}, []any{functions.GtFunctionName, 0}})
	if err != nil || res != true {
		t.Fatalf("every failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.EveryFunctionName, []any{functions.ListFunctionName, 2, 3, 6}, []any{functions.GtFunctionName, 2}})
	if err != nil || res != false {
		t.Fatalf("every false check failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.SubstringFunctionName, "hello", 1, 4})
	if err != nil || res != "ell" {
		t.Fatalf("substring failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.RoundToIntFunctionName, 3.7})
	if err != nil || res != 4 {
		t.Fatalf("round_to_int failed: %v got %v", err, res)
	}
}

func TestEvaluateExpressionAliasAndErrorCases(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.IsInListFunctionName, []any{1, 2, 3}, 2})
	if err != nil || res != true {
		t.Fatalf("is_in_list failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ContainsInListFunctionName, []any{"a", "b"}, "b"})
	if err != nil || res != true {
		t.Fatalf("contains_in_list alias failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.IsInStringFunctionName, "hello", "ell"})
	if err != nil || res != true {
		t.Fatalf("is_in_string alias failed: %v got %v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.GetIndexFunctionName, []any{"a", "b"}, "b"})
	if err != nil || res != 1 {
		t.Fatalf("get_index alias failed: %v got %v", err, res)
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ModFunctionName, 1, 0})
	if err == nil {
		t.Fatalf("expected division by zero for mod 1 0")
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.PowFunctionName, "x", 2})
	if err == nil {
		t.Fatalf("expected error for pow non-int base")
	}

	_, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.FlattenFunctionName, 123})
	if err == nil {
		t.Fatalf("expected error for flatten non-list arg")
	}
}

func TestEvaluateExpressionExtraOps(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	cases := []struct {
		name string
		expr []any
		want any
	}{
		{"mod", []any{functions.ModFunctionName, 10, 3}, 1},
		{"pow", []any{functions.PowFunctionName, 2, 5}, 32},
		{"abs", []any{functions.AbsFunctionName, -7}, 7},
		{"floor", []any{functions.FloorFunctionName, 2.9}, 2},
		{"ceil", []any{functions.CeilFunctionName, 2.1}, 3},
		{"is_nil", []any{functions.IsNilFunctionName, nil}, true},
		{"is_number", []any{functions.IsNumberFunctionName, 42}, true},
		{"is_string", []any{functions.IsStringFunctionName, "x"}, true},
		{"is_list", []any{functions.IsListFunctionName, []any{1, 2}}, true},
		{"is_map", []any{functions.IsMapFunctionName, map[string]any{"a": 1}}, true},
		{"head", []any{functions.HeadFunctionName, []any{1, 2}}, 1},
		{"tail", []any{functions.TailFunctionName, []any{1, 2, 3}}, []any{2, 3}},
		{"first", []any{functions.FirstFunctionName, []any{9, 8}}, 9},
		{"last", []any{functions.LastFunctionName, []any{9, 8, 7}}, 7},
		{"keys", []any{functions.KeysFunctionName, map[string]any{"a": 1, "b": 2}}, []any{"a", "b"}},
		{"values", []any{functions.ValuesFunctionName, map[string]any{"a": 1, "b": 2}}, []any{1, 2}},
		{"merge", []any{functions.MergeFunctionName, map[string]any{"a": 1}, map[string]any{"b": 2}}, map[string]any{"a": 1, "b": 2}},
		{"flatten", []any{functions.FlattenFunctionName, []any{functions.ListFunctionName, []any{1, 2}, 3}}, []any{1, 2, 3}},
		{"union", []any{functions.UnionFunctionName, []any{1, 2}, []any{2, 3}}, []any{1, 2, 3}},
		{"difference", []any{functions.DifferenceFunctionName, []any{1, 2, 3}, []any{2}}, []any{1, 3}},
		{"unique", []any{functions.UniqueFunctionName, []any{1, 2, 2, 3}}, []any{1, 2, 3}},
	}

	for _, c := range cases {
		res, err := evaluate.EvaluateExpression(ctx, nil, c.expr)
		if err != nil {
			t.Fatalf("%s failed: %v", c.name, err)
		}
		if !reflect.DeepEqual(res, c.want) {
			t.Fatalf("%s got %v want %v", c.name, res, c.want)
		}
	}
}
