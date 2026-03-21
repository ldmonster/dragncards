package evaluate_test

import (
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

func TestEvaluateExpressionArithmeticAndLogic(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	cases := []struct {
		expr any
		want any
	}{
		{[]any{functions.AddFunctionName, 3, 5}, 8},
		{[]any{functions.SubFunctionName, 10, 4}, 6},
		{[]any{functions.MulFunctionName, 2, 3}, 6},
		{[]any{functions.DivFunctionName, 20, 5}, 4},
		{[]any{functions.EqFunctionName, 5, 5}, true},
		{[]any{functions.NeqFunctionName, 5, 6}, true},
		{[]any{functions.GtFunctionName, 5, 3}, true},
		{[]any{functions.LtFunctionName, 3, 5}, true},
		{[]any{functions.GteFunctionName, 5, 5}, true},
		{[]any{functions.LteFunctionName, 4, 4}, true},
		{[]any{functions.AndFunctionName, true, true}, true},
		{[]any{functions.OrFunctionName, false, true}, true},
		{[]any{functions.NotFunctionName, false}, true},
	}

	ctx2 := evaluate.NewEvalContext(game.NewGameUI("room-2"))
	ctx2.Game.Cards["c1"] = &game.Card{ID: "c1", StackID: "s1"}
	ctx2.Game.Stacks["s1"] = game.NewStack("s1", "g1")
	ctx2.Game.Stacks["s1"].AddCard("c1")
	ctx2.Game.Groups["g1"] = game.NewGroup("g1", "g1", "")
	ctx2.Game.Groups["g1"].AddStack("s1")
	ctx2.Game.Stacks["s2"] = game.NewStack("s2", "g1")
	ctx2.Game.Groups["g1"].AddStack("s2")

	_, err := evaluate.EvaluateExpression(ctx2, nil, []any{functions.MoveCardFunctionName, "c1", "g1", 1, 0})
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

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.JoinStringFunctionName, "hello", "world"})
	if err != nil {
		t.Fatalf("join_string failed: %v", err)
	}
	if res != "helloworld" {
		t.Fatalf("join_string got %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.InStringFunctionName, "helloworld", "hello"})
	if err != nil {
		t.Fatalf("in_string failed: %v", err)
	}
	if res != true {
		t.Fatalf("in_string got %v", res)
	}
}

func TestEvaluateExpressionContainsAndReverse(t *testing.T) {
	ctx := evaluate.NewEvalContext(game.NewGameUI("room-1"))

	res, err := evaluate.EvaluateExpression(ctx, nil, []any{functions.ContainsFunctionName, []any{functions.ListFunctionName, 1, 2, 3}, 2})
	if err != nil || res != true {
		t.Fatalf("contains failed: %v result=%v", err, res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.IsInListFunctionName, []any{functions.ListFunctionName, "a", "b", "c"}, "b"})
	if err != nil || res != true {
		t.Fatalf("is_in_list failed: %v result=%v", err, res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ContainsFunctionName, "hello", "ll"})
	if err != nil || res != true {
		t.Fatalf("contains string failed: %v result=%v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.IsInStringFunctionName, "helloworld", "world"})
	if err != nil || res != true {
		t.Fatalf("is_in_string failed: %v result=%v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.IsInListFunctionName, []any{functions.ListFunctionName, "a", "b", "c"}, "b"})
	if err != nil || res != true {
		t.Fatalf("is_in_list failed: %v result=%v", err, res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.GetIndexFunctionName, []any{functions.ListFunctionName, "x", "y", "z"}, "y"})
	if err != nil || res != 1 {
		t.Fatalf("get_index failed: %v result=%v", err, res)
	}
	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.AtIndexFunctionName, []any{functions.ListFunctionName, "x", "y", "z"}, 2})
	if err != nil || res != "z" {
		t.Fatalf("at_index failed: %v result=%v", err, res)
	}

	// get_stack_id / get_card_id
	ctx.Game.Groups["g1"] = game.NewGroup("g1", "g1", "")
	ctx.Game.Stacks["s1"] = game.NewStack("s1", "g1")
	ctx.Game.Stacks["s1"].AddCard("c1")
	ctx.Game.Stacks["s1"].AddCard("c2")
	ctx.Game.Groups["g1"].AddStack("s1")

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.GetStackIdFunctionName, "g1", 0})
	if err != nil || res != "s1" {
		t.Fatalf("get_stack_id failed: %v result=%v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.GetCardIdFunctionName, "g1", 0, 1})
	if err != nil || res != "c2" {
		t.Fatalf("get_card_id failed: %v result=%v", err, res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ReverseFunctionName, []any{functions.ListFunctionName, 1, 2, 3}})
	if err != nil {
		t.Fatalf("reverse list failed: %v", err)
	}
	if got, ok := res.([]any); !ok || len(got) != 3 || got[0] != 3 {
		t.Fatalf("reverse list wrong: %v", res)
	}

	res, err = evaluate.EvaluateExpression(ctx, nil, []any{functions.ReverseFunctionName, "abc"})
	if err != nil || res != "cba" {
		t.Fatalf("reverse string failed: %v result=%v", err, res)
	}
}
