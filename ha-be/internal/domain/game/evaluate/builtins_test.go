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
}
