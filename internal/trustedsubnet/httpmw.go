package trustedsubnet

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

// MiddlewareHandler creates an HTTP middleware that checks if the client IP
// is within the trusted subnet.
//
// The middleware reads the IP address from the X-Real-IP header.
// If network is nil, all requests are allowed.
// If the client IP is not in the trusted subnet, returns 403 Forbidden.
func MiddlewareHandler(network *net.IPNet, logger *zap.Logger) func(http.Handler) http.Handler {
	if network == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !network.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
