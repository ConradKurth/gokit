package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/logger"
	"github.com/ConradKurth/gokit/responses"
	"github.com/dgrijalva/jwt-go"
	jwtverifier "github.com/okta/okta-jwt-verifier-golang"
)

type contextKey string

func (c contextKey) String() string {
	return "auth-key-" + string(c)
}

const (
	adminKey        = contextKey("admin")
	adminHTTPHeader = "X-Admin-Key"
)

// NewMiddleware returns a new auth middleware.
func NewMiddleware(c *config.Config) func(http.Handler) http.Handler {
	return auth(c, newVerifier)
}

func auth(c *config.Config, createVerify createVerifyFunc) func(http.Handler) http.Handler {
	enabled := c.GetBoolDefault("auth.enabled", true)
	cids := c.GetStringSlice("auth.cids")

	verifiers := map[string]tokenVerify{}
	for _, id := range cids {
		toValidate := map[string]string{}
		toValidate["aud"] = c.GetString("auth.aud")
		toValidate["cid"] = id

		verifiers[id] = createVerify(c.GetString("auth.issuer"), toValidate)
	}

	adminSecrets := strings.Split(c.GetString("auth.admin.key"), ",")
	mapping := map[string]struct{}{}
	for _, a := range adminSecrets {
		mapping[a] = struct{}{}
	}
	parser := jwt.Parser{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			adminHeader := r.Header.Get(adminHTTPHeader)
			_, ok := mapping[adminHeader]
			if ok && adminHeader != "" {
				// let's inject that we are an admin
				ctx := SetAdmin(r.Context())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			splitToken := strings.Split(r.Header.Get("authorization"), "Bearer")
			if len(splitToken) != 2 {
				logger.GetLoggerReq(r).WarnCtx(r.Context(),
					"Token not of size two",
					logger.Any("size", len(splitToken)))
				responses.Empty(w, http.StatusUnauthorized)
				return
			}

			authToken := strings.TrimSpace(splitToken[1])

			claims := jwt.MapClaims{}
			_, _, err := parser.ParseUnverified(authToken, claims)
			if err != nil {
				logger.GetLoggerReq(r).WarnCtx(r.Context(),
					"Error parsing claims",
					logger.ErrField(err))
				responses.Empty(w, http.StatusUnauthorized)
				return
			}

			cid, ok := claims["cid"].(string)
			if !ok {
				logger.GetLoggerReq(r).WarnCtx(r.Context(),
					"no audience in the claims",
					logger.ErrField(err))
				responses.Empty(w, http.StatusUnauthorized)
				return
			}

			v, ok := verifiers[cid]
			if !ok {
				logger.GetLoggerReq(r).WarnCtx(r.Context(),
					"Did not have verifier for audience",
					logger.ErrField(err),
					logger.Any("cid", cid))
				responses.Empty(w, http.StatusUnauthorized)
				return
			}

			if _, err := v.VerifyAccessToken(authToken); err != nil {
				responses.Empty(w, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IsAdmin will return if the is admin token is set
func IsAdmin(ctx context.Context) bool {
	_, ok := ctx.Value(adminKey).(bool)
	return ok
}

// SetAdmin will set the context with the admin key. Be very careful using this
func SetAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, adminKey, true)
}

// NewAdminCheck will ensure that the is admin token is set. Useful for admin routes
func NewAdminCheck() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !IsAdmin(r.Context()) {
				responses.Empty(w, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type tokenVerify interface {
	VerifyAccessToken(jwt string) (*jwtverifier.Jwt, error)
}

type createVerifyFunc func(issuer string, claims map[string]string) tokenVerify

func newVerifier(issuer string, claims map[string]string) tokenVerify {
	jwtVerifierSetup := jwtverifier.JwtVerifier{
		Issuer:           issuer,
		ClaimsToValidate: claims,
	}

	return jwtVerifierSetup.New()
}
