package ssl

import (
	"net/http"
)

var sslProxyHeaders = map[string]string{
	"X-Forwarded-Proto": "https",
}

// NewMiddleware returns a new ssl redirection middleware.
func NewMiddleware(isDevelopment bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ssl := false
			for k, v := range sslProxyHeaders {
				if r.Header.Get(k) == v {
					ssl = true
					break
				}
			}

			if ssl || isDevelopment {
				next.ServeHTTP(w, r)
				return
			}

			url := r.URL
			url.Scheme = "https"
			url.Host = r.Host

			http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)
		})
	}
}
