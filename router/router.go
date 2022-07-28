package router

import (
	"dsh/px/app"

	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Every app's HTTP enpoint is under this pattern, like
//
//   /demoa/hello
//
// This endpoint process hello for app demoa.
const appIDPattern = "/{appID:[a-z]+[0-9]*}"

// New creates and returns [*chi.Mux] router for our global application app. It
// uses list of HTTP endpoints from [routes.go].
func New(app *app.Global) *chi.Mux {
	return NewWithRoutes(app, allAppRoutes)
}

// NewWithRoutes creates and returns [*chi.Mux] router for application app, adds
// middlewares and adds HTTP endpoints from subRoutes under [appIDPattern]
// route.
func NewWithRoutes(app *app.Global, subRoutes routesList) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route(appIDPattern, func(r chi.Router) {
		for _, v := range subRoutes {
			r.Method(v.Method, v.Pattern, appHandler{app, v.Handler})
		}
	})

	return r
}

// handleFunc defines function, which process HTTP endpoint. See [1] for an
// inspiration.
//
// [1]: http://blog.questionable.services/article/custom-handlers-avoiding-globals/
type handleFunc func(*app.Context, http.ResponseWriter, *http.Request)

// appHandler defines [http.Handler], which joins our global app and handleFn.
type appHandler struct {
	app      *app.Global
	handleFn handleFunc
}

// ServeHTTP process HTTP request. It extracts appID from URI, creates app
// context for this appID and calls our handleFn with that context.
func (self appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appID")
	ctx, err := self.app.NewContext(appID)
	if err != nil {
		panic(err)
	}
	self.handleFn(ctx, w, r)
}

func hello(app *app.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hello! AppID = %v", app.AppID())))
}
