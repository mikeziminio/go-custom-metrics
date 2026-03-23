package ip

import (
	"net"
	"net/http"
	"strings"
)

// extractIPAddress extracts the IP address from the request.
//
// It checks X-Forwarded-For, X-Real-IP headers, and falls back to RemoteAddr.
func ExtractIPAddress(r *http.Request) string {
	// First try to get IP from X-Forwarded-For header
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For might have multiple IPs, get the first one
		ips := []string{}
		for _, v := range []string{",", ";"} {
			ips = append(ips, strings.Split(ip, v)...)
		}
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
		// Validate IP format
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Then try X-Real-IP header
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Finally fall back to RemoteAddr
	if ip == "" {
		ip = r.RemoteAddr
		// Extract IP from "host:port" format
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		// Validate IP format
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// If we can't parse any IP, return empty string
	return ""
}
