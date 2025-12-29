package server_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mockserver/config"
)

func newTestStore() *StateStore {
	return NewStateStore()
}

// 1. CREATE ACTION TESTS
func TestApplyStateful_Create(t *testing.T) {
	store := newTestStore()
	cfg := &config.StatefulConfig{
		Collection: "users",
		Action:     "create",
		IDField:    "id",
	}

	ctx := &EContext{
		Body: map[string]interface{}{"id": 1, "name": "Ahmet"},
	}

	err := ApplyStateful(store, cfg, ctx)
	require.NoError(t, err)
	assert.NotNil(t, ctx.State.Created)
	assert.Equal(t, "Ahmet", ctx.State.Created["name"])
	assert.Len(t, store.collections["users"], 1)

	// Scenario 1: Re-creation with the same ID (Conflict Error)
	ctxConflict := &EContext{
		Body: map[string]interface{}{"id": 1, "name": "Mehmet"},
	}
	errConflict := ApplyStateful(store, cfg, ctxConflict)
	assert.Equal(t, StateErrConflict, errConflict, "Aynı ID varsa Conflict hatası dönmeli")

	// Scenario 2: Missing identity field (Bad Input)
	ctxBad := &EContext{
		Body: map[string]interface{}{"name": "No ID"},
	}
	errBad := ApplyStateful(store, cfg, ctxBad)
	assert.Equal(t, StateErrBadInput, errBad, "ID yoksa BadInput dönmeli")
}


// 2. GET & LIST ACTION TESTS
func TestApplyStateful_GetAndList(t *testing.T) {
	store := newTestStore()
	store.collections["products"] = []map[string]interface{}{
		{"code": "P1", "price": 100},
		{"code": "P2", "price": 200},
	}

	// Senaryo 1: Listing
	cfgList := &config.StatefulConfig{Collection: "products", Action: "list"}
	ctxList := &EContext{}
	err := ApplyStateful(store, cfgList, ctxList)
	require.NoError(t, err)
	assert.Len(t, ctxList.State.List, 2)

	// Senaryo 2: Get (Successful)
	cfgGet := &config.StatefulConfig{Collection: "products", Action: "get", IDField: "code"}
	ctxGet := &EContext{Path: map[string]string{"code": "P1"}}
	errGet := ApplyStateful(store, cfgGet, ctxGet)
	require.NoError(t, errGet)
	assert.Equal(t, 100, ctxGet.State.Item["price"])

	// Senaryo 3: Get (Not found)
	ctxNotFound := &EContext{Path: map[string]string{"code": "P99"}}
	errNotFound := ApplyStateful(store, cfgGet, ctxNotFound)
	assert.Equal(t, StateErrNotFound, errNotFound)
}

// 3. UPDATE ACTION TESTS
func TestApplyStateful_Update(t *testing.T) {
	store := newTestStore()
	store.collections["todos"] = []map[string]interface{}{
		{"id": 10, "title": "Old Title", "done": false},
	}

	cfg := &config.StatefulConfig{
		Collection: "todos",
		Action:     "update",
		IDField:    "id",
	}

	// Scenario 1: Successful Update
    // Note: Even if the ID string is “10”, it should match int 10 
	ctx := &EContext{
		Path: map[string]string{"id": "10"},
		Body: map[string]interface{}{"title": "New Title", "done": true},
	}

	err := ApplyStateful(store, cfg, ctx)
	require.NoError(t, err)
	assert.Equal(t, "New Title", ctx.State.Updated["title"])
	assert.Equal(t, true, ctx.State.Updated["done"])

	// Store'daki veriyi kontrol et
	updatedItem := store.collections["todos"][0]
	assert.Equal(t, "New Title", updatedItem["title"])

	// Scenario 2: Updating a non-existent ID
	ctxFail := &EContext{
		Path: map[string]string{"id": "999"},
		Body: map[string]interface{}{"title": "Ghost"},
	}
	errFail := ApplyStateful(store, cfg, ctxFail)
	assert.Equal(t, StateErrNotFound, errFail)
}


// 4. DELETE ACTION TESTS
func TestApplyStateful_Delete(t *testing.T) {
	store := newTestStore()
	store.collections["users"] = []map[string]interface{}{
		{"id": 1, "name": "Ali"},
		{"id": 2, "name": "Veli"},
	}

	cfg := &config.StatefulConfig{
		Collection: "users",
		Action:     "delete",
		IDField:    "id",
	}

	// Scenario 1: Successful Deletion (ID: 1)
	ctx := &EContext{Path: map[string]string{"id": "1"}}
	err := ApplyStateful(store, cfg, ctx)
	require.NoError(t, err)

	// There should be 1 person left on the list
	assert.Len(t, store.collections["users"], 1)
	assert.Equal(t, "Veli", store.collections["users"][0]["name"])
	assert.Len(t, ctx.State.List, 1) // Güncel liste context'e dönmeli

	// Scenario 2: Deleting a non-existent ID
	ctxFail := &EContext{Path: map[string]string{"id": "999"}}
	errFail := ApplyStateful(store, cfg, ctxFail)
	assert.Equal(t, StateErrNotFound, errFail)
}