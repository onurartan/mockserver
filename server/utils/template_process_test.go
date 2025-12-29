package server_utils

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER: Mock Context Setup (Strict Types)
func getTemplateContext() EContext {
	return EContext{
		Body: map[string]interface{}{
			"username": "johndoe",
			"role":     "admin",
		},
		Query: map[string]string{
			"lang": "en",
		},
		Headers: map[string]string{
			"x-api-key": "secret-123",
		},
		State: &StateContext{
			List: []map[string]interface{}{
				{"id": 1, "status": "pending"},
				{"id": 2, "status": "shipped"},
			},
			Item: map[string]interface{}{
				"id": 99, "status": "delivered",
			},
			Created: map[string]interface{}{"success": true},
			Updated: map[string]interface{}{"modified": true},
		},
	}
}

// 1. FAKER & GENERATOR TESTS
func TestProcessTemplate_Faker(t *testing.T) {
	ctx := EContext{} 

	tests := []struct {
		name     string
		template string
		verify   func(t *testing.T, res interface{})
	}{
		{
			name:     "UUID Generation",
			template: "{{uuid}}",
			verify:   func(t *testing.T, res interface{}) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Len(t, s, 36)
			},
		},
		{
			name:     "Email Generation",
			template: "{{email}}",
			verify:   func(t *testing.T, res interface{}) {
				s, ok := res.(string)
				require.True(t, ok)
				assert.Contains(t, s, "@")
			},
		},
		{
			name:     "Number with Args",
			template: "{{number min=100 max=200}}",
			verify:   func(t *testing.T, res interface{}) {
				s, ok := res.(string)
				require.True(t, ok)
				val, _ := strconv.Atoi(s)
				assert.GreaterOrEqual(t, val, 100)
				assert.LessOrEqual(t, val, 200)
			},
		},
		{
			name:     "Date Future",
			template: "{{dateFuture days=5}}",
			verify:   func(t *testing.T, res interface{}) {
				s, ok := res.(string)
				require.True(t, ok)
				parsed, err := time.Parse("2006-01-02", s)
				require.NoError(t, err)
				assert.True(t, parsed.After(time.Now().AddDate(0, 0, -1)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ProcessTemplateJSON(tt.template, ctx)
			require.NoError(t, err)
			tt.verify(t, res)
		})
	}
}

// 2. CONTEXT INJECTION (Request Data)
func TestProcessTemplate_ContextInjection(t *testing.T) {
	ctx := getTemplateContext()

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"Inject Body", "Hello {{request.body.username}}", "Hello johndoe"},
		{"Inject Query", "Language: {{request.query.lang}}", "Language: en"},
		{"Inject Headers", "Key: {{request.headers.x-api-key}}", "Key: secret-123"},
		{"Partial Match", "User: {{request.body.username}} - Role: {{request.body.role}}", "User: johndoe - Role: admin"},
		{"Missing Key", "Missing: {{request.body.notfound}}", "Missing: {{request.body.notfound}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ProcessTemplateJSON(tt.template, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}


// 3. STATE SHORTCUTS (Raw Object Return)
func TestProcessTemplate_StateShortcuts(t *testing.T) {
	ctx := getTemplateContext()

	// Case 1: {{state.list}} -> Should return []map[string]interface{}
	resList, err := ProcessTemplateJSON("{{state.list}}", ctx)
	require.NoError(t, err)
	
	list, ok := resList.([]map[string]interface{})
	require.True(t, ok, "state.list should return a []map[string]interface{}, check your struct definition")
	assert.Len(t, list, 2)

	// Case 2: {{state.item}} -> Should return map[string]interface{}
	resItem, err := ProcessTemplateJSON("{{state.item}}", ctx)
	require.NoError(t, err)

	item, ok := resItem.(map[string]interface{})
	require.True(t, ok, "state.item should return a map")
	assert.Equal(t, 99, item["id"])
}

// 4. RECURSIVE PROCESSING
func TestProcessTemplate_Recursion(t *testing.T) {
	ctx := getTemplateContext()

	input := map[string]interface{}{
		"meta": map[string]interface{}{
			"user": "{{request.body.username}}",
			"timestamp": "{{date}}",
		},
		"data": []interface{}{
			map[string]interface{}{
				"id": "{{uuid}}",
				"type": "generated",
			},
			map[string]interface{}{
				"static": "value",
			},
		},
	}

	res, err := ProcessTemplateJSON(input, ctx)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)

	meta := resMap["meta"].(map[string]interface{})
	assert.Equal(t, "johndoe", meta["user"])

	data := resMap["data"].([]interface{})
	require.Len(t, data, 2)

	item1 := data[0].(map[string]interface{})
	assert.Len(t, item1["id"], 36)
}