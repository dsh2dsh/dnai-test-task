package router

import (
	"net/http"
)

// A routesList defines app's HTTP endpoints.
type routesList []struct {
	// HTTP method like [http.MethodGet].
	Method string
	// URI pattern in format expected by [chi].
	Pattern string
	// Function which handles this endpoint.
	Handler handleFunc
}

// allAppRoutes contains list of HTTP endpoints under appIDPattern.
// Every item has type of routesList.
var allAppRoutes = routesList{
	{http.MethodGet, "/hello", hello},
}
