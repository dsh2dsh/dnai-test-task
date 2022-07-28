package router

import (
	"dsh/px/app"
	"time"

	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Let's test our HTTP routes.
func TestRouteAppID(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	g := new(app.Global)
	r := New(g)

	// Collect all defined routes under appIDPattern into a map with key combaned
	// from HTTP method and pattern, like "GET /hello". The value doesn't matter.
	allSubRoutes := make(map[string]bool)
	for _, route := range subRoutesOf(t, r, appIDPattern) {
		t.Logf("[%v] Pattern = %s Handlers = %v",
			appIDPattern, route.Pattern, route.Handlers)
		for method := range route.Handlers {
			allSubRoutes[method+" "+route.Pattern] = true
		}
	}

	// Num of registered endpoints should equal num of defined endpoints.
	assert.Len(allSubRoutes, len(allAppRoutes), "Extra routes detected")

	// Let's check all defined endpoints are in registered endpoints.
	for _, v := range allAppRoutes {
		k := v.Method + " " + v.Pattern
		require.Contains(allSubRoutes, k, "Route not found")
		delete(allSubRoutes, k)
	}

	// Let's check our appIDPattern hierarchy hasn't extra, unknown endpoints.
	assert.Emptyf(allSubRoutes, "Found extra sub routes %v", allSubRoutes)
}

// subRoutesOf extracts all sub routes under hierarchy of pattern route and
// returns them as slice. If pattern route isn't registered it generates Fatal.
func subRoutesOf(t *testing.T, r *chi.Mux, pattern string) []chi.Route {
	p := pattern + "/*"
	for _, route := range r.Routes() {
		t.Logf("Pattern = %s", route.Pattern)
		if route.Pattern == p {
			return route.SubRoutes.Routes()
		}
	}
	t.Fatalf("Route %v not found", pattern)
	return nil
}

// Let's test how our appHandler extracts appID from URI.
func TestExtractAppID(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rndURI := rndTestURI(t)
	var appID string

	routes := routesList{
		{
			http.MethodHead,
			rndURI,
			func(ctx *app.Context, w http.ResponseWriter, r *http.Request) {
				appID = ctx.AppID()
				w.Write([]byte(fmt.Sprintf("AppID = %v", appID)))
			},
		},
	}

	g := new(app.Global)
	r := NewWithRoutes(g, routes)

	ts := httptest.NewServer(r)
	defer ts.Close()

	wantAppID := "demoa"
	resp, err := http.DefaultClient.Head(ts.URL + "/" + wantAppID + rndURI)
	require.NoError(err)
	defer resp.Body.Close()

	assert.Equal(appID, wantAppID)
}

// rndTestURI returns random URI like "/test-RANDOMHEXSTRING".
func rndTestURI(t *testing.T) string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		t.Fatal(err)
	}
	return "/test-" + hex.EncodeToString(bytes)
}
