package auth

import (
	"net/http"

	"github.com/lucmichalski/cars-dataset/pkg/core/auth/claims"
	"github.com/lucmichalski/cars-dataset/pkg/core/session"
)

// Context context
type Context struct {
	*Auth
	Claims   *claims.Claims
	Provider Provider
	Request  *http.Request
	Writer   http.ResponseWriter
}

// Flashes get flash messages
func (context Context) Flashes() []session.Message {
	return context.Auth.SessionStorer.Flashes(context.Writer, context.Request)
}

// FormValue get form value with name
func (context Context) FormValue(name string) string {
	return context.Request.Form.Get(name)
}