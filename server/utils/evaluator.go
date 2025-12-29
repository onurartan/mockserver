package server_utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// EvaluateCondition parses and executes boolean expressions against the request context.
// Supports logical operators (AND, OR) and grouping.
func EvaluateCondition(expr string, ctx EContext) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, errors.New("empty condition")
	}

	// Normalize logical operators
	expr = strings.ReplaceAll(expr, "&&", " AND ")
	expr = strings.ReplaceAll(expr, "||", " OR ")
	expr = strings.ReplaceAll(expr, " and ", " AND ")
	expr = strings.ReplaceAll(expr, " or ", " OR ")

	// Split OR
	orParts := splitLogical(expr, "OR")
	for _, orPart := range orParts {
		andOK := true
		for _, andPart := range splitLogical(orPart, "AND") {
			ok, err := evalSingleCondition(strings.TrimSpace(andPart), ctx)
			if err != nil {
				return false, fmt.Errorf("failed evaluating '%s': %w", andPart, err)
			}
			if !ok {
				andOK = false
				break
			}
		}
		if andOK {
			return true, nil
		}
	}
	return false, nil
}

func splitLogical(expr, op string) []string {
	return strings.Split(expr, " "+op+" ")
}

// evalSingleCondition parses a binary comparison (e.g., "a > b") or a type check.
func evalSingleCondition(cond string, ctx EContext) (bool, error) {
	ops := []string{"==", "!=", "<=", ">=", "<", ">"}

	var op string
	for _, o := range ops {
		if strings.Contains(cond, o) {
			op = o
			break
		}
	}
	if op == "" {
		return false, fmt.Errorf("invalid operator in '%s'", cond)
	}

	parts := strings.SplitN(cond, op, 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid condition format: '%s'", cond)
	}

	leftTrim := strings.TrimSpace(parts[0])
	rightTrim := strings.TrimSpace(parts[1])

	// Special Case: Type Checking function type(var)
	if strings.HasPrefix(leftTrim, "type(") && strings.HasSuffix(leftTrim, ")") {
		inner := strings.TrimSpace(leftTrim[5 : len(leftTrim)-1])
		val, err := evalResolveValue(inner, ctx)
		if err != nil {
			return false, fmt.Errorf("failed to resolve value for type(): %w", err)
		}
		expectedType := strings.Trim(rightTrim, "'\" ")
		ok, err := evalTypeCheck(val, expectedType, op)
		if err != nil {
			return false, err
		}
		return ok, nil
	}

	// Standard Value Comparison
	leftVal, err := evalResolveValue(leftTrim, ctx)
	if err != nil {
		return false, fmt.Errorf("left value error: %w", err)
	}

	rightVal, err := evalParseLiteral(rightTrim)
	if err != nil {
		return false, fmt.Errorf("right value error: %w", err)
	}

	return evalCompareValues(leftVal, rightVal, op)
}

func evalTypeCheck(value interface{}, expectedType string, operator string) (bool, error) {
	var actualType string

	switch v := value.(type) {
	case string:
		// Attempt Type Coercion: Check if string is actually a number
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			actualType = "number"
		} else {
			actualType = "string"
		}
	case float64, int:
		actualType = "number"
	case bool:
		actualType = "boolean"
	case map[string]interface{}:
		actualType = "dict"
	default:
		return false, fmt.Errorf("unsupported type detected: %T", value)
	}

	if operator != "==" && operator != "!=" {
		return false, fmt.Errorf("invalid operator for type() comparison: '%s'. Only '==' or '!=' allowed", operator)
	}

	switch operator {
	case "==":
		return actualType == expectedType, nil
	case "!=":
		return actualType != expectedType, nil
	}

	return false, fmt.Errorf("unknown error in type check")
}

func evalParseLiteral(val string) (interface{}, error) {
	val = strings.TrimSpace(val)

	// string
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") && len(val) >= 2 {
		return val[1 : len(val)-1], nil
	}

	// boolean
	if val == "true" {
		return true, nil
	}
	if val == "false" {
		return false, nil
	}

	// number
	if n, err := strconv.ParseFloat(val, 64); err == nil {
		return n, nil
	}

	return nil, fmt.Errorf("invalid literal value: '%s'", val)
}

// evalResolveValue extracts data from the EContext using dot notation (e.g., request.body.id).
// Supports scopes: body, query, headers, path.
func evalResolveValue(path string, ctx EContext) (interface{}, error) {
	if !strings.HasPrefix(path, "request.") {
		return nil, fmt.Errorf("invalid reference (must start with 'request.'): '%s'", path)
	}

	parts := strings.Split(path, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request reference: '%s'", path)
	}

	scope := parts[1]
	key := parts[2]

	var val interface{}
	var ok bool

	switch scope {
	case "body":
		for k, v := range ctx.Body {
			if strings.EqualFold(k, key) { // case-insensitive match
				val = v
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("body key '%s' not found", key)
		}
		return val, nil

	case "query":
		for k, v := range ctx.Query {
			if strings.EqualFold(k, key) {
				val = v
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("query key '%s' not found", key)
		}
		return val, nil

	case "headers":
		for k, v := range ctx.Headers {
			if strings.EqualFold(k, key) {
				val = v
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("header key '%s' not found", key)
		}
		return val, nil

	case "path":
		for k, v := range ctx.Path {
			if strings.EqualFold(k, key) {
				val = v
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("path key '%s' not found", key)
		}
		return val, nil

	default:
		return nil, fmt.Errorf("unknown request scope: '%s'", scope)
	}
}

// evalCompareValues performs the actual comparison logic with automatic type coercion.
func evalCompareValues(a interface{}, b interface{}, op string) (bool, error) {
	if a == nil || b == nil {
		return false, fmt.Errorf("cannot compare nil values: a=%v, b=%v", a, b)
	}

	// Helper: Coerces any numeric-like value (int, string-number) to float64
	convertToFloat := func(val interface{}) (float64, bool) {
		switch t := val.(type) {
		case float64:
			return t, true
		case int:
			return float64(t), true
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f, true
			}
		}
		return 0, false
	}

	switch av := a.(type) {
	// Numeric Comparison (includes int & float64)
	case float64, int:
		af, _ := convertToFloat(av)
		bf, bok := convertToFloat(b)
		if !bok {
			return false, fmt.Errorf("type mismatch: left numeric, right %T", b)
		}
		return compareFloats(af, bf, op)

	case string:

		if af, aok := convertToFloat(av); aok {
			if bf, bok := convertToFloat(b); bok {
				return compareFloats(af, bf, op)

			}
		}
		// string comparison
		bs, ok := b.(string)
		if !ok {
			return false, fmt.Errorf("type mismatch: left string, right %T", b)
		}
		switch op {
		case "==":
			return av == bs, nil
		case "!=":
			return av != bs, nil
		default:
			return false, fmt.Errorf("unsupported operator for string: %s", op)
		}

		// Boolean Comparison
	case bool:
		bb, ok := b.(bool)
		if !ok {
			return false, fmt.Errorf("type mismatch: left bool, right %T", b)
		}
		switch op {
		case "==":
			return av == bb, nil
		case "!=":
			return av != bb, nil
		default:
			return false, fmt.Errorf("unsupported operator for bool: %s", op)
		}
	default:
		return false, fmt.Errorf("unsupported comparison types: %T %T", a, b)
	}

	// return false, fmt.Errorf("unsupported comparison types: %T %T with operator '%s'", a, b, op) [OLD]
}

// Helper to reduce duplicated switch cases
func compareFloats(a, b float64, op string) (bool, error) {
	switch op {
	case ">":
		return a > b, nil
	case ">=":
		return a >= b, nil
	case "<":
		return a < b, nil
	case "<=":
		return a <= b, nil
	case "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	}
	return false, fmt.Errorf("unsupported comparison types: %T %T with operator '%s'", a, b, op)
}
