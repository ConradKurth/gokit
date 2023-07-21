package userinfo

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ConradKurth/gokit/logger"
	jwt "github.com/dgrijalva/jwt-go"
)

type contextKey string

func (c contextKey) String() string {
	return "user-key-" + string(c)
}

const (
	groupsKey = contextKey("groups")
	scopeKey  = contextKey("scope")
	userIDKey = contextKey("userID")
)

// NewMiddleware returns a new user info middleware.
func NewMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			splitToken := strings.Split(r.Header.Get("authorization"), "Bearer")
			if len(splitToken) != 2 {
				next.ServeHTTP(w, r)
				return
			}

			claims := jwt.MapClaims{}
			authToken := strings.TrimSpace(splitToken[1])
			parser := jwt.Parser{}
			_, _, err := parser.ParseUnverified(authToken, claims)
			if err != nil {
				logger.GetLoggerReq(r).WarnCtx(r.Context(), "Error parsing claims", logger.ErrField(err))
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			if v, ok := claims["scp"]; ok {
				ctx = context.WithValue(ctx, scopeKey, v)
			}

			if v, ok := claims["uid"]; ok {
				ctx = context.WithValue(ctx, userIDKey, v)
			}

			if v, ok := claims["groups"]; ok {
				ctx = context.WithValue(ctx, groupsKey, v)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID finds the userId from the context. REQUIRES Scopes Middleware to have run.
func GetUserID(ctx context.Context) (string, error) {
	val, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", errors.New("no userID was set")
	}
	return val, nil
}

// GetScopes finds the scopes from the context. REQUIRES Scopes Middleware to have run.
func GetScopes(ctx context.Context) ([]string, error) {
	val, ok := ctx.Value(scopeKey).([]string)
	if !ok {
		return nil, errors.New("no scopes found")
	}
	return val, nil
}

// GetGroups finds the groups from the context. REQUIRES Scopes Middleware to have run.
func GetGroups(ctx context.Context) ([]string, error) {
	val, ok := ctx.Value(groupsKey).([]interface{})
	if !ok {
		return nil, errors.New("no groups found")
	}

	groups := make([]string, 0, len(val))
	for _, v := range val {
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("group is not a string")
		}
		groups = append(groups, s)
	}
	return groups, nil
}

// IsAdmin returns true if the groups has the admin group in it
func IsAdmin(ctx context.Context) bool {
	groups, err := GetGroups(ctx)
	if err != nil {
		return false
	}

	for _, g := range groups {
		if strings.EqualFold(g, "admins") {
			return true
		}
	}
	return false
}
