package functions

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// ConcatFunction joins two strings.
type ConcatFunction struct{}

func (*ConcatFunction) Name() string { return ConcatFunctionName }
func (*ConcatFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ConcatFunctionName, args, 2); err != nil {
		return nil, err
	}
	a, err := requireStringArg("concat", args, 0)
	if err != nil {
		return nil, err
	}
	b, err := requireStringArg("concat", args, 1)
	if err != nil {
		return nil, err
	}
	return a + b, nil
}

// JoinStringFunction is an alias of concat.
type JoinStringFunction struct{}

func (*JoinStringFunction) Name() string { return JoinStringFunctionName }
func (*JoinStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(JoinStringFunctionName, args, 2); err != nil {
		return nil, err
	}
	a, err := requireStringArg("join_string", args, 0)
	if err != nil {
		return nil, err
	}
	b, err := requireStringArg("join_string", args, 1)
	if err != nil {
		return nil, err
	}
	return a + b, nil
}

// InStringFunction (case-sensitive contains).
type InStringFunction struct{}

func (*InStringFunction) Name() string { return InStringFunctionName }
func (*InStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(InStringFunctionName, args, 2); err != nil {
		return nil, err
	}
	L, err := requireStringArg("in_string", args, 0)
	if err != nil {
		return nil, err
	}
	R, err := requireStringArg("in_string", args, 1)
	if err != nil {
		return nil, err
	}
	return strings.Contains(L, R), nil
}

// IsInStringFunction alias for in_string (Elixir compatibility).
type IsInStringFunction struct{}

func (*IsInStringFunction) Name() string { return IsInStringFunctionName }
func (*IsInStringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	return (&InStringFunction{}).Execute(ctx, args)
}

// ToLowercaseFunction returns lower-case string.
type ToLowercaseFunction struct{}

func (*ToLowercaseFunction) Name() string { return ToLowercaseFunctionName }
func (*ToLowercaseFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ToLowercaseFunctionName, args, 1); err != nil {
		return nil, err
	}
	v, ok := args[0].(string)
	if !ok {
		return nil, errors.New("to_lowercase arg must be string")
	}
	return strings.ToLower(v), nil
}

// ToUppercaseFunction returns upper-case string.
type ToUppercaseFunction struct{}

func (*ToUppercaseFunction) Name() string { return ToUppercaseFunctionName }
func (*ToUppercaseFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ToUppercaseFunctionName, args, 1); err != nil {
		return nil, err
	}
	v, ok := args[0].(string)
	if !ok {
		return nil, errors.New("to_uppercase arg must be string")
	}
	return strings.ToUpper(v), nil
}

// ReplaceStringInListFunction replaces all occurrences of old in each string in list.
type ReplaceStringInListFunction struct{}

func (*ReplaceStringInListFunction) Name() string { return ReplaceStringInListFunctionName }
func (*ReplaceStringInListFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(ReplaceStringInListFunctionName, args, 3); err != nil {
		return nil, err
	}
	list, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("replace_string_in_list first arg must be list")
	}
	old, ok := args[1].(string)
	if !ok {
		return nil, errors.New("replace_string_in_list second arg must be string")
	}
	newVal, ok := args[2].(string)
	if !ok {
		return nil, errors.New("replace_string_in_list third arg must be string")
	}
	out := make([]any, 0, len(list))
	for _, item := range list {
		s, ok := item.(string)
		if !ok {
			return nil, errors.New("replace_string_in_list list elements must be strings")
		}
		out = append(out, strings.ReplaceAll(s, old, newVal))
	}
	return out, nil
}

// SubstringFunction returns a substring from [start:end] in string.
type SubstringFunction struct{}

func (*SubstringFunction) Name() string { return SubstringFunctionName }
func (*SubstringFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(SubstringFunctionName, args, 3); err != nil {
		return nil, err
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, errors.New("substring first arg must be string")
	}
	start, ok := asInt(args[1])
	if !ok {
		return nil, errors.New("substring second arg must be int")
	}
	end, ok := asInt(args[2])
	if !ok {
		return nil, errors.New("substring third arg must be int")
	}
	if start < 0 || end < start || end > len(s) {
		return nil, errors.New("substring index out of range")
	}
	return s[start:end], nil
}

// RoundToIntFunction rounds a numeric value to nearest int.
type RoundToIntFunction struct{}

func (*RoundToIntFunction) Name() string { return RoundToIntFunctionName }
func (*RoundToIntFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(RoundToIntFunctionName, args, 1); err != nil {
		return nil, err
	}
	switch v := args[0].(type) {
	case int, int32, int64:
		x, _ := asInt(v)
		return x, nil
	case float32:
		return int(math.Round(float64(v))), nil
	case float64:
		return int(math.Round(v)), nil
	default:
		return nil, errors.New("round_to_int arg must be numeric")
	}
}

// FormatFunction runs fmt.Sprintf over args.
type FormatFunction struct{}

func (*FormatFunction) Name() string { return FormatFunctionName }
func (*FormatFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if len(args) == 0 {
		return nil, errors.New("format requires at least 1 arg")
	}
	fmtStr, ok := args[0].(string)
	if !ok {
		return nil, errors.New("format first arg must be string")
	}
	return fmt.Sprintf(fmtStr, args[1:]...), nil
}

// RegexReplaceFunction replaces with regexp pattern.
type RegexReplaceFunction struct{}

func (*RegexReplaceFunction) Name() string { return RegexReplaceFunctionName }
func (*RegexReplaceFunction) Execute(ctx *game.EvalContext, args []any) (any, error) {
	if err := requireArgCount(RegexReplaceFunctionName, args, 3); err != nil {
		return nil, err
	}
	text, ok := args[0].(string)
	if !ok {
		return nil, errors.New("regex_replace first arg must be string")
	}
	pattern, ok := args[1].(string)
	if !ok {
		return nil, errors.New("regex_replace second arg must be string")
	}
	repl, ok := args[2].(string)
	if !ok {
		return nil, errors.New("regex_replace third arg must be string")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.ReplaceAllString(text, repl), nil
}
