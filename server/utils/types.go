package server_utils


type StateContext struct {
	List    []map[string]interface{}
	Item    map[string]interface{}
	Created map[string]interface{}
	Updated map[string]interface{}
}

type EContext struct {
	Body    map[string]interface{}
	Query   map[string]string
	Headers map[string]string
	Path    map[string]string

	State *StateContext
}