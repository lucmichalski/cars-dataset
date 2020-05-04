package auth

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"

	"github.com/lucmichalski/cars-dataset/pkg/core/auth/auth_identity"
	"github.com/lucmichalski/cars-dataset/pkg/core/auth/claims"
	"github.com/lucmichalski/cars-dataset/pkg/core/mailer"
	"github.com/lucmichalski/cars-dataset/pkg/core/mailer/logger"
	"github.com/lucmichalski/cars-dataset/pkg/core/redirect_back"
	"github.com/lucmichalski/cars-dataset/pkg/core/render"
	"github.com/lucmichalski/cars-dataset/pkg/core/session/manager"
)

// Auth auth struct
type Auth struct {
	*Config
	// Embed SessionStorer to match Authority's AuthInterface
	SessionStorerInterface
	providers []Provider
}

// Config auth config
type Config struct {
	// Default Database, which will be used in Auth when do CRUD, you can change a request's DB isntance by setting request Context's value, refer https://github.com/lucmichalski/cars-dataset/pkg/core/auth/blob/master/utils.go#L32
	DB *gorm.DB
	// AuthIdentityModel a model used to save auth info, like email/password, OAuth token, linked user's ID, https://github.com/lucmichalski/cars-dataset/pkg/core/auth/blob/master/auth_identity/auth_identity.go is the default implemention
	AuthIdentityModel interface{}
	// UserModel should be point of user struct's instance, it could be nil, then Auth will assume there is no user linked to auth info, and will return current auth info when get current user
	UserModel interface{}
	// Mount Auth into router with URLPrefix's value as prefix, default value is `/auth`.
	URLPrefix string
	// ViewPaths prepend views paths for auth
	ViewPaths []string

	// Auth is using [Render](https://github.com/lucmichalski/cars-dataset/pkg/core/render) to render pages, you could configure it with your project's Render if you have advanced usage like [BindataFS](https://github.com/qorpress/bindatafs)
	Render *render.Render
	// Auth is using [Mailer](https://github.com/lucmichalski/cars-dataset/pkg/core/mailer) to send email, by default, it will print email into console, you need to configure it to send real one
	Mailer *mailer.Mailer
	// UserStorer is an interface that defined how to get/save user, Auth provides a default one based on AuthIdentityModel, UserModel's definition
	UserStorer UserStorerInterface
	// SessionStorer is an interface that defined how to encode/validate/save/destroy session data and flash messages between requests, Auth provides a default method do the job, to use the default value, don't forgot to mount SessionManager's middleware into your router to save session data correctly. refer [session](https://github.com/lucmichalski/cars-dataset/pkg/core/session) for more details
	SessionStorer SessionStorerInterface
	// Redirector redirect user to a new page after registered, logged, confirmed...
	Redirector RedirectorInterface

	// LoginHandler defined behaviour when request `{Auth Prefix}/login`, default behaviour defined in http://godoc.org/github.com/lucmichalski/cars-dataset/pkg/core/auth#pkg-variables
	LoginHandler func(*Context, func(*Context) (*claims.Claims, error))
	// RegisterHandler defined behaviour when request `{Auth Prefix}/register`, default behaviour defined in http://godoc.org/github.com/lucmichalski/cars-dataset/pkg/core/auth#pkg-variables
	RegisterHandler func(*Context, func(*Context) (*claims.Claims, error))
	// LogoutHandler defined behaviour when request `{Auth Prefix}/logout`, default behaviour defined in http://godoc.org/github.com/lucmichalski/cars-dataset/pkg/core/auth#pkg-variables
	LogoutHandler func(*Context)
}

// New initialize Auth
func New(config *Config) *Auth {
	if config == nil {
		config = &Config{}
	}

	if config.URLPrefix == "" {
		config.URLPrefix = "/auth/"
	} else {
		config.URLPrefix = fmt.Sprintf("/%v/", strings.Trim(config.URLPrefix, "/"))
	}

	if config.AuthIdentityModel == nil {
		config.AuthIdentityModel = &auth_identity.AuthIdentity{}
	}

	if config.Render == nil {
		config.Render = render.New(nil)
	}

	if config.Mailer == nil {
		config.Mailer = mailer.New(&mailer.Config{
			Sender: logger.New(&logger.Config{}),
		})
	}

	if config.UserStorer == nil {
		config.UserStorer = &UserStorer{}
	}

	if config.SessionStorer == nil {
		config.SessionStorer = &SessionStorer{
			SessionName:    "_auth_session",
			SessionManager: manager.SessionManager,
			SigningMethod:  jwt.SigningMethodHS256,
		}
	}

	if config.Redirector == nil {
		config.Redirector = &Redirector{redirect_back.New(&redirect_back.Config{
			SessionManager:  manager.SessionManager,
			IgnoredPrefixes: []string{config.URLPrefix},
		})}
	}

	if config.LoginHandler == nil {
		config.LoginHandler = DefaultLoginHandler
	}

	if config.RegisterHandler == nil {
		config.RegisterHandler = DefaultRegisterHandler
	}

	if config.LogoutHandler == nil {
		config.LogoutHandler = DefaultLogoutHandler
	}

	for _, viewPath := range config.ViewPaths {
		config.Render.RegisterViewPath(viewPath)
	}

	config.Render.RegisterViewPath("github.com/lucmichalski/cars-dataset/pkg/core/auth/views")

	auth := &Auth{Config: config}

	auth.SessionStorerInterface = config.SessionStorer

	return auth
}
