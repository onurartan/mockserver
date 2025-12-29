package server_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperContext returns a rich context populated with various data types
// to simulate a real HTTP request environment.
func helperContext() EContext {
	return EContext{
		Body: map[string]interface{}{
			"age":      25,
			"price":    19.99,
			"role":     "admin",
			"active":   true,
			"quantity": "50", // String number to test type coercion
		},
		Query: map[string]string{
			"search": "laptop",
			"page":   "1",
			"limit":  "100",
		},
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
		Path: map[string]string{
			"id":       "101",
			"category": "electronics",
		},
	}
}

// TestEvaluateCondition_Basics verifies standard comparison operators (==, !=, >, <)
// against different data types (int, float, string, bool).
func TestEvaluateCondition_Basics(t *testing.T) {
	ctx := helperContext()

	tests := []struct {
		name      string
		expr      string
		want      bool
		expectErr bool
	}{
		// Numeric Comparisons
		{"Number Greater Than", "request.body.age > 18", true, false},
		{"Number Less Than", "request.body.price < 50", true, false},
		{"Number Equals", "request.body.age == 25", true, false},
		{"Number Not Equals", "request.body.age != 30", true, false},

		// String Comparisons
		{"String Equals", "request.body.role == 'admin'", true, false},
		{"String Not Equals", "request.body.role != 'user'", true, false},

		// Boolean Comparisons
		{"Bool True", "request.body.active == true", true, false},
		{"Bool False", "request.body.active == false", false, false},

		// Path & Query & Header Resolution
		{"Query Param Check", "request.query.search == 'laptop'", true, false},
		{"Path Param Check", "request.path.id == '101'", true, false},
		{"Header Check", "request.headers.Content-Type == 'application/json'", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.expr, ctx)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got, "Expression: %s", tt.expr)
			}
		})
	}
}

// TestEvaluateCondition_Logic verifies the parsing and execution of complex
// logical chains using AND/OR operators.
func TestEvaluateCondition_Logic(t *testing.T) {
	ctx := helperContext()

	tests := []struct {
		name string
		expr string
		want bool
	}{
		// AND Logic
		{"Simple AND (True)", "request.body.age > 18 AND request.body.role == 'admin'", true},
		{"Simple AND (False)", "request.body.age > 18 AND request.body.role == 'guest'", false},
		{"Symbol && Support", "request.body.age > 18 && request.body.active == true", true},

		// OR Logic
		{"Simple OR (Left True)", "request.body.role == 'admin' OR request.body.age < 10", true},
		{"Simple OR (Right True)", "request.body.role == 'guest' OR request.body.age > 10", true},
		{"Simple OR (Both False)", "request.body.role == 'guest' OR request.body.age < 10", false},
		{"Symbol || Support", "request.body.role == 'guest' || request.body.active == true", true},

		// Mixed Logic
		{"Mixed AND/OR", "request.body.age > 18 AND request.body.active == true OR request.query.page == '99'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestEvaluateCondition_TypeCoercion ensures that the system is smart enough
// to compare a string number ("50") with a real number (50).
func TestEvaluateCondition_TypeCoercion(t *testing.T) {
	ctx := helperContext()

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{"String Number vs Int", "request.body.quantity == 50", true},   // "50" == 50
		{"String Number vs Int GT", "request.body.quantity > 40", true}, // "50" > 40
		{"Query String vs Int", "request.query.limit == 100", true},     // "100" == 100
		{"Path String vs Int", "request.path.id > 100", true},           // "101" > 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestEvaluateCondition_TypeCheck verifies the special `type()` function logic.
func TestEvaluateCondition_TypeCheck(t *testing.T) {
	ctx := helperContext()

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{"Check Number Type", "type(request.body.age) == 'number'", true},
		{"Check String Type", "type(request.body.role) == 'string'", true},
		{"Check Boolean Type", "type(request.body.active) == 'boolean'", true},

		// "quantity" is string "50" in context, but looks like a number.
		// Your logic parses float, so it might return 'number'. Let's verify expected behavior.
		// Based on code: if ParseFloat succeeds, it returns "number".
		{"Check String-Number as Number", "type(request.body.quantity) == 'number'", true},

		{"Check Query Param as Number", "type(request.query.limit) == 'number'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestEvaluateCondition_EdgeCases verifies case-insensitive key lookups and error states.
func TestEvaluateCondition_EdgeCases(t *testing.T) {
	ctx := helperContext()

	// 1. Case Insensitive Lookup Test
	t.Run("Case Insensitive Key", func(t *testing.T) {
		// "Role" vs "role" in context
		got, err := EvaluateCondition("request.body.Role == 'admin'", ctx)
		require.NoError(t, err)
		assert.True(t, got)
	})

	// 2. Missing Key Error
	t.Run("Missing Key", func(t *testing.T) {
		_, err := EvaluateCondition("request.body.nonExistentKey == 'foo'", ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// 3. Invalid Operator
	t.Run("Invalid Operator", func(t *testing.T) {
		_, err := EvaluateCondition("request.body.age >> 18", ctx)
		require.Error(t, err)
	})

	// 4. Invalid Syntax
	t.Run("Invalid Syntax", func(t *testing.T) {
		_, err := EvaluateCondition("just a string", ctx)
		require.Error(t, err)
	})
}
