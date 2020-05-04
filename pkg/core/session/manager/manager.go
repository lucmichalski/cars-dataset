package manager

import (
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/lucmichalski/cars-dataset/pkg/core/middlewares"
	"github.com/lucmichalski/cars-dataset/pkg/core/session"
	"github.com/lucmichalski/cars-dataset/pkg/core/session/gorilla"
)

// SessionManager default session manager
var SessionManager session.ManagerInterface = gorilla.New("_session", sessions.NewCookieStore([]byte("secret")))

func init() {
	middlewares.Use(middlewares.Middleware{
		Name: "session",
		Handler: func(handler http.Handler) http.Handler {
			return SessionManager.Middleware(handler)
		},
	})
}
