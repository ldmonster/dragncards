package functions

import "github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"

type NoopFunction struct{}

func (n *NoopFunction) Name() string {
	return "noop"
}

func (n *NoopFunction) Execute(ctx *evaluate.EvalContext, args []any) (any, error) {
	return nil, nil
}
