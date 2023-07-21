package routetest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/middleware/auth"
	"github.com/ConradKurth/gokit/middleware/userinfo"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// AdminHeader the header we use for auth
const AdminHeader = "X-Admin-Key"

// AdminKey is a testing key we will use
const AdminKey = "1234"

// AuthDisabled disabled auth for tests
const AuthDisabled = `{"auth": {"enabled": false}}`

// AdminAuth is a config with auth enabled and an admin key set
var AdminAuth = fmt.Sprintf(`{"auth": {"admin": {"key": "%v"}}}`, AdminKey)

// the userID in here is cow
const DummyJWT = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJjb3cifQ.mAdWYTzSCE69JKCtjeOHHP-acbM5Z29akKTh_vo-JM4`

const NoUserDummyJWT = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIifQ.UIZchxQD36xuhacrJF9HQ5SIUxH5HBiv9noESAacsxU`

func CheckUserId(path, method string, r Registerable) RouteTestCase {
	return RouteTestCase{
		Name:   "Ensure userID",
		JWT:    NoUserDummyJWT,
		Path:   path,
		Method: method,
		Router: func() (Registerable, []func(http.Handler) http.Handler) {
			return r, WithMiddlewares(AuthDisabledConfig())
		},
		Assert: func(*testing.T, Registerable) {},
		ResponseHandler: func(t *testing.T, resp *http.Response) {
			t.Helper()
			defer resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		},
	}
}

// CheckIsAdminBlocked will ensure that only admins can access this endpoint if the header is set
func CheckIsAdminBlocked(path, method string, r Registerable) RouteTestCase {
	return RouteTestCase{
		Name:   "Ensure only admin access",
		Path:   path,
		Method: method,
		Router: func() (Registerable, []func(http.Handler) http.Handler) {
			return r, WithMiddlewares(AuthAdminConfig())
		},
		Assert: func(t *testing.T, _ Registerable) {
			t.Helper()
		},
		ResponseHandler: func(t *testing.T, resp *http.Response) {
			t.Helper()
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		},
	}
}

// Registerable is the core requirement to register the routes
type Registerable interface {
	RegisterRoutes(c chi.Router, routeLevelMiddleware ...func(http.Handler) http.Handler)
}

// WithMiddlewares will add in all the core middleware we use
func WithMiddlewares(c *config.Config) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{auth.NewMiddleware(c)}
}

// AuthDisabledConfig will return an auth disabled config
func AuthDisabledConfig() *config.Config {
	buff := bytes.NewBuffer([]byte(AuthDisabled))
	return config.LoadConfig(config.WithReader(buff))
}

// AuthAdminConfig will return an admin auth config
func AuthAdminConfig() *config.Config {
	buff := bytes.NewBuffer([]byte(AdminAuth))
	return config.LoadConfig(config.WithReader(buff))
}

func initServer(r Registerable, m []func(http.Handler) http.Handler) *httptest.Server {

	mux := chi.NewMux()
	mux.Use(userinfo.NewMiddleware())
	r.RegisterRoutes(mux, m...)

	return httptest.NewServer(mux)
}

type RouteTestCase struct {
	Name            string
	JWT             string
	Path            string
	Header          map[string]string
	Method          string
	Body            []byte
	Router          func() (Registerable, []func(http.Handler) http.Handler)
	Assert          func(*testing.T, Registerable)
	ResponseHandler func(*testing.T, *http.Response)
}

func RouteRunner(t *testing.T, tt []RouteTestCase) {
	t.Helper()
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r, m := tc.Router()
			server := initServer(r, m)
			defer server.Close()

			buff := bytes.NewBuffer(tc.Body)

			req, err := http.NewRequest(tc.Method, server.URL+tc.Path, buff)
			req.Header.Set("Authorization", "Bearer "+tc.JWT)

			for k, v := range tc.Header {
				req.Header.Set(k, v)
			}
			assert.NoError(t, err)

			resp, err := server.Client().Do(req)
			assert.NoError(t, err)

			tc.ResponseHandler(t, resp)
			tc.Assert(t, r)
		})
	}
}
