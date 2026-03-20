package evaluate

import "fmt"

// EvalError wraps errors from expression evaluation and allows WS path distinction.
type EvalError struct {
	Err error
}

func (e *EvalError) Error() string {
	if e == nil || e.Err == nil {
		return "evaluation error"
	}
	return fmt.Sprintf("evaluation error: %v", e.Err)
}

func (e *EvalError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewEvalError(err error) error {
	if err == nil {
		return nil
	}
	return &EvalError{Err: err}
}
