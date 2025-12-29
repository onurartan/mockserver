package server_utils

import "fmt"
import "errors"

import (
	config "mockserver/config"
)

var (
	StateErrNotFound = errors.New("state: item not found")
	StateErrConflict = errors.New("state: item already exists")
	StateErrBadInput = errors.New("state: invalid input")
)

func ApplyStateful(
	store *StateStore,
	cfg *config.StatefulConfig,
	ctx *EContext,
) error {

	if ctx.State == nil {
		ctx.State = &StateContext{}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	col := store.collections[cfg.Collection]
	if col == nil {
		col = []map[string]interface{}{}
	}

	idField := cfg.IDField
	if idField == "" {
		idField = "id"
	}

	switch cfg.Action {

	case "create":
		item := ctx.Body
		idVal, ok := item[idField]
		if !ok {
			return StateErrBadInput
		}

		// 🔥 CONFLICT CHECK
		for _, existing := range col {
			if fmt.Sprint(existing[idField]) == fmt.Sprint(idVal) {
				return StateErrConflict
			}
		}

		col = append(col, item)
		store.collections[cfg.Collection] = col

		ctx.State.Created = item
		ctx.State.List = col

	case "list":
		ctx.State.List = col

	case "get":
		id := ctx.Path[idField]
		for _, item := range col {
			if fmt.Sprint(item[idField]) == id {
				ctx.State.Item = item
				return nil
			}
		}
		return StateErrNotFound

	case "update":
		id := ctx.Path[idField]
		for i, item := range col {
			if fmt.Sprint(item[idField]) == id {
				for k, v := range ctx.Body {
					item[k] = v
				}
				col[i] = item
				store.collections[cfg.Collection] = col

				ctx.State.Updated = item
				return nil
			}
		}
		return StateErrNotFound

	case "delete":
		id := ctx.Path[idField]
		found := false
		newCol := make([]map[string]interface{}, 0, len(col))

		for _, item := range col {
			if fmt.Sprint(item[idField]) == id {
				found = true
				continue
			}
			newCol = append(newCol, item)
		}

		if !found {
			return StateErrNotFound
		}

		store.collections[cfg.Collection] = newCol
		ctx.State.List = newCol

	default:
		return fmt.Errorf("unknown stateful action: %s", cfg.Action)
	}

	return nil
}
