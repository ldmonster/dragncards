// Package evaluate_test contains integration tests that need to import both
// the evaluate package and its functions sub-package. Using an external test
// package (evaluate_test) avoids the import cycle that would arise if the
// internal evaluator_test.go imported evaluate/functions (which imports evaluate).
package evaluate_test

import (
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate/functions"
)

func TestEvaluatorBuiltins(t *testing.T) {
	e := evaluate.NewEvaluator()
	functions.RegisterBuiltins(e)

	r, err := e.EvalFunction(functions.AddFunctionName, nil, []any{3, 5})
	if err != nil {
		t.Fatalf("add builtin failed: %v", err)
	}
	if r != 8 {
		t.Fatalf("add result wrong: %v", r)
	}

	r, err = e.EvalFunction(functions.SubFunctionName, nil, []any{10, 7})
	if err != nil {
		t.Fatalf("sub builtin failed: %v", err)
	}
	if r != 3 {
		t.Fatalf("sub result wrong: %v", r)
	}

	r, err = e.EvalFunction(functions.MulFunctionName, nil, []any{4, 5})
	if err != nil {
		t.Fatalf("mul builtin failed: %v", err)
	}
	if r != 20 {
		t.Fatalf("mul result wrong: %v", r)
	}

	r, err = e.EvalFunction(functions.DivFunctionName, nil, []any{20, 5})
	if err != nil {
		t.Fatalf("div builtin failed: %v", err)
	}
	if r != 4 {
		t.Fatalf("div result wrong: %v", r)
	}

	_, err = e.EvalFunction(functions.DivFunctionName, nil, []any{10, 0})
	if err == nil {
		t.Fatal("expected div-by-zero error, got nil")
	}

	// Comparison builtin tests
	r, err = e.EvalFunction(functions.EqFunctionName, nil, []any{4, 4})
	if err != nil || r != true {
		t.Fatalf("eq builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.EqFunctionName, nil, []any{"a", "a"})
	if err != nil || r != true {
		t.Fatalf("eq string builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.NeqFunctionName, nil, []any{4, 2})
	if err != nil || r != true {
		t.Fatalf("neq builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.GtFunctionName, nil, []any{5, 3})
	if err != nil || r != true {
		t.Fatalf("gt builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.LtFunctionName, nil, []any{2, 9})
	if err != nil || r != true {
		t.Fatalf("lt builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.GteFunctionName, nil, []any{8, 8})
	if err != nil || r != true {
		t.Fatalf("gte builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.LteFunctionName, nil, []any{7, 7})
	if err != nil || r != true {
		t.Fatalf("lte builtin failed: %v result=%v", err, r)
	}

	// Boolean operators
	r, err = e.EvalFunction(functions.AndFunctionName, nil, []any{true, false})
	if err != nil || r != false {
		t.Fatalf("and builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.OrFunctionName, nil, []any{true, false})
	if err != nil || r != true {
		t.Fatalf("or builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.NotFunctionName, nil, []any{true})
	if err != nil || r != false {
		t.Fatalf("not builtin failed: %v result=%v", err, r)
	}

	// New DSL functions: contains and reverse
	r, err = e.EvalFunction(functions.ContainsFunctionName, nil, []any{[]any{1, 2, 3}, 2})
	if err != nil || r != true {
		t.Fatalf("contains list failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.ContainsFunctionName, nil, []any{"hello", "ll"})
	if err != nil || r != true {
		t.Fatalf("contains string failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction(functions.ReverseFunctionName, nil, []any{[]any{1, 2, 3}})
	if err != nil {
		t.Fatalf("reverse list failed: %v", err)
	}
	if reversed, ok := r.([]any); !ok || len(reversed) != 3 || reversed[0] != 3 {
		t.Fatalf("reverse list result wrong: %v", r)
	}

	r, err = e.EvalFunction(functions.ReverseFunctionName, nil, []any{"abc"})
	if err != nil || r != "cba" {
		t.Fatalf("reverse string failed: %v result=%v", err, r)
	}
}
