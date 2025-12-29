package server_utils

import "sync"

type StateStore struct {
	mu          sync.RWMutex
	collections map[string][]map[string]interface{}
}

func NewStateStore() *StateStore {
	return &StateStore{
		collections: make(map[string][]map[string]interface{}),
	}
}
