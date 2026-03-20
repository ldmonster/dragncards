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

	r, err := e.EvalFunction("add", nil, []any{3, 5})
	if err != nil {
		t.Fatalf("add builtin failed: %v", err)
	}
	if r != 8 {
		t.Fatalf("add result wrong: %v", r)
	}

	r, err = e.EvalFunction("sub", nil, []any{10, 7})
	if err != nil {
		t.Fatalf("sub builtin failed: %v", err)
	}
	if r != 3 {
		t.Fatalf("sub result wrong: %v", r)
	}

	r, err = e.EvalFunction("mul", nil, []any{4, 5})
	if err != nil {
		t.Fatalf("mul builtin failed: %v", err)
	}
	if r != 20 {
		t.Fatalf("mul result wrong: %v", r)
	}

	r, err = e.EvalFunction("div", nil, []any{20, 5})
	if err != nil {
		t.Fatalf("div builtin failed: %v", err)
	}
	if r != 4 {
		t.Fatalf("div result wrong: %v", r)
	}

	_, err = e.EvalFunction("div", nil, []any{10, 0})
	if err == nil {
		t.Fatal("expected div-by-zero error, got nil")
	}

	// Comparison builtin tests
	r, err = e.EvalFunction("eq", nil, []any{4, 4})
	if err != nil || r != true {
		t.Fatalf("eq builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("eq", nil, []any{"a", "a"})
	if err != nil || r != true {
		t.Fatalf("eq string builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("neq", nil, []any{4, 2})
	if err != nil || r != true {
		t.Fatalf("neq builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("gt", nil, []any{5, 3})
	if err != nil || r != true {
		t.Fatalf("gt builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("lt", nil, []any{2, 9})
	if err != nil || r != true {
		t.Fatalf("lt builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("gte", nil, []any{8, 8})
	if err != nil || r != true {
		t.Fatalf("gte builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("lte", nil, []any{7, 7})
	if err != nil || r != true {
		t.Fatalf("lte builtin failed: %v result=%v", err, r)
	}

	// Boolean operators
	r, err = e.EvalFunction("and", nil, []any{true, false})
	if err != nil || r != false {
		t.Fatalf("and builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("or", nil, []any{true, false})
	if err != nil || r != true {
		t.Fatalf("or builtin failed: %v result=%v", err, r)
	}

	r, err = e.EvalFunction("not", nil, []any{true})
	if err != nil || r != false {
		t.Fatalf("not builtin failed: %v result=%v", err, r)
	}
}
