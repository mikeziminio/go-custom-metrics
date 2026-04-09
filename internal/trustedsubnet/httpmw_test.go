package trustedsubnet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMiddlewareHandler(t *testing.T) {
	testCases := []struct {
		name           string
		cidr           string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "success - IP in subnet",
			cidr:           "192.168.1.0/24",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - no network set",
			cidr:           "",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden - missing X-Real-IP header",
			cidr:           "192.168.1.0/24",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "forbidden - invalid IP",
			cidr:           "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "forbidden - IP not in subnet",
			cidr:           "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "success - IPv6 in subnet",
			cidr:           "2001:db8::/32",
			xRealIP:        "2001:db8::1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var network *net.IPNet
			if tc.cidr != "" {
				var err error
				network, err = ParseCIDR(tc.cidr)
				assert.NoError(t, err)
			}

			logger := zap.NewNop()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := MiddlewareHandler(network, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/", http.NoBody)
			if tc.xRealIP != "" {
				req.Header.Set("X-Real-IP", tc.xRealIP)
			}

			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}
