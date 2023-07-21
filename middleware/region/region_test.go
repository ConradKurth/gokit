package region_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConradKurth/gokit/middleware/region"
	"github.com/ConradKurth/gokit/middleware/userinfo"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegionMiddleware(t *testing.T) {
	tt := []struct {
		Name        string
		Expected    region.Region
		Header      map[string]string
		ErrExpected error
	}{

		{
			Name:     "Default to US when no lang tag or user set",
			Expected: region.USA,
		},
		{
			Name:     "US from the region header",
			Expected: region.USA,
			Header: map[string]string{
				"X-Region": "USA",
			},
		},
		{
			Name:     "Hong kong from the region header",
			Expected: region.HongKong,
			Header: map[string]string{
				"X-Region": "HongKong",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(userinfo.NewMiddleware())
			r.Use(region.NewMiddleware())

			r.Get("/v1/do", func(w http.ResponseWriter, r *http.Request) {
				l, err := region.GetRegion(r.Context())
				if tc.ErrExpected != nil {
					assert.EqualError(t, err, tc.ErrExpected.Error())
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tc.Expected, l)
			})

			req, err := http.NewRequest(http.MethodGet, "/v1/do", nil)
			require.NoError(t, err)
			for k, v := range tc.Header {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			expectedCode := http.StatusOK
			if tc.ErrExpected != nil {
				expectedCode = http.StatusBadRequest
			}
			assert.Equal(t, rec.Code, expectedCode)
		})
	}
}
