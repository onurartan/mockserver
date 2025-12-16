package server_handlers

const (
	RouteTypeMock      = "mock"
	RouteTypeFetch     = "fetch"
	RouteTypeInternal  = "internal"
	RouteTypeUnmatched = "unmatched"
)

const (
	CtxRequestID      = "__req_id"
	CtxRouteType      = "__route_type" // "mock" | "fetch"
	CtxRoutePath      = "__route_path"
	CtxRouteName      = "__route_name"
	CtxUpstreamURL    = "__up_url"
	CtxUpstreamStatus = "__up_status"
	CtxUpstreamTimeMs = "__up_time_ms"
)
