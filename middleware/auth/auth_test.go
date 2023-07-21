package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConradKurth/gokit/config"
	"github.com/go-chi/chi/v5"
	jwtverifier "github.com/okta/okta-jwt-verifier-golang"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockVerify struct {
	mock.Mock
}

func (m *mockVerify) VerifyAccessToken(jwt string) (*jwtverifier.Jwt, error) {
	args := m.Called(jwt)
	return args.Get(0).(*jwtverifier.Jwt), args.Error(1)
}

func defaultMockVerify(_ *testing.T) func(issuer string, claims map[string]string) tokenVerify {
	return func(issuer string, claims map[string]string) tokenVerify {
		m := &mockVerify{}
		return m
	}
}

func TestAuthMiddleware(t *testing.T) {
	tt := []struct {
		Name       string
		Headers    map[string]string
		Config     *bytes.Buffer
		IsValid    bool
		Code       int
		AdminCheck bool
		Verifier   func(t *testing.T) func(issuer string, claims map[string]string) tokenVerify
	}{
		{
			Name:     "Auth is disabled",
			IsValid:  true,
			Code:     http.StatusOK,
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":false}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "Happy path",
			Headers: map[string]string{
				"Authorization": "Bearer " + "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4",
			},
			IsValid: true,
			Code:    http.StatusOK,
			Config:  bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: func(t *testing.T) func(issuer string, claims map[string]string) tokenVerify {
				t.Helper()
				return func(issuer string, claims map[string]string) tokenVerify {
					assert.Equal(t, "cowwww", issuer)
					assert.Equal(t, map[string]string{
						"aud": "foo",
						"cid": "1244",
					}, claims)
					m := &mockVerify{}
					m.On("VerifyAccessToken", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4").Once().Return(&jwtverifier.Jwt{}, nil)
					return m
				}
			},
		},
		{
			Name: "Happy path with multiple cids",
			Headers: map[string]string{
				"Authorization": "Bearer " + "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4",
			},
			IsValid: true,
			Code:    http.StatusOK,
			Config:  bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["6435","1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: func(t *testing.T) func(issuer string, claims map[string]string) tokenVerify {
				t.Helper()
				return func(issuer string, claims map[string]string) tokenVerify {
					assert.Equal(t, "cowwww", issuer)
					if claims["cid"] == "1244" {
						assert.Equal(t, map[string]string{
							"aud": "foo",
							"cid": "1244",
						}, claims)
					} else {
						assert.Equal(t, map[string]string{
							"aud": "foo",
							"cid": "6435",
						}, claims)
					}
					m := &mockVerify{}
					m.On("VerifyAccessToken", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4").Once().Return(&jwtverifier.Jwt{}, nil)
					return m
				}
			},
		},
		{
			Name:       "Admin token is present",
			IsValid:    true,
			AdminCheck: true,
			Code:       http.StatusOK,
			Headers: map[string]string{
				adminHTTPHeader: "secret",
			},
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"admin":{"key":"secret"}}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "No token present",
			Headers: map[string]string{
				"Authorization": "Bearer",
			},
			Code:     http.StatusUnauthorized,
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "Invalid token",
			Headers: map[string]string{
				"Authorization": "qweqweqwe",
			},
			Code:     http.StatusUnauthorized,
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "no cid present in the payload",
			Headers: map[string]string{
				"Authorization": "Bearer " + "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			IsValid:  false,
			Code:     http.StatusUnauthorized,
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "No matching cid",
			Headers: map[string]string{
				"Authorization": "Bearer " + "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4",
			},
			Code:     http.StatusUnauthorized,
			Config:   bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["5555"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: defaultMockVerify,
		},
		{
			Name: "Error verifying the token",
			Headers: map[string]string{
				"Authorization": "Bearer " + "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4",
			},
			Code:   http.StatusUnauthorized,
			Config: bytes.NewBuffer([]byte(`{"auth":{"enabled":true,"cids":["1244"],"aud":"foo","issuer":"cowwww"}}`)),
			Verifier: func(t *testing.T) func(issuer string, claims map[string]string) tokenVerify {
				t.Helper()
				return func(issuer string, claims map[string]string) tokenVerify {
					assert.Equal(t, "cowwww", issuer)
					assert.Equal(t, map[string]string{
						"aud": "foo",
						"cid": "1244",
					}, claims)
					m := &mockVerify{}
					m.On("VerifyAccessToken", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJjaWQiOiIxMjQ0In0.8KrCGzKATOuwE5H9_Uxn4b_Xnis1pn0vA6tLWf2mDd4").Once().Return(&jwtverifier.Jwt{}, errors.New("bad verify"))
					return m
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			conf := config.LoadConfig(config.WithReader(tc.Config))

			r := chi.NewRouter()
			r.Use(auth(conf, tc.Verifier(t)))

			executed := false

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if tc.AdminCheck {
					assert.True(t, IsAdmin(r.Context()))
				}
				executed = true
				w.WriteHeader(http.StatusOK)
			})

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)
			for k, v := range tc.Headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.IsValid, executed)
			assert.Equal(t, tc.Code, rec.Code)
		})
	}
}
