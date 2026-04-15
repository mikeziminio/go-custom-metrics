package trustedsubnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCIDR(t *testing.T) {
	testCases := []struct {
		name        string
		cidr        string
		shouldError bool
	}{
		{
			name:        "valid IPv4 CIDR",
			cidr:        "192.168.1.0/24",
			shouldError: false,
		},
		{
			name:        "valid IPv6 CIDR",
			cidr:        "2001:db8::/32",
			shouldError: false,
		},
		{
			name:        "invalid CIDR",
			cidr:        "invalid-cidr",
			shouldError: true,
		},
		{
			name:        "empty CIDR",
			cidr:        "",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			network, err := ParseCIDR(tc.cidr)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Nil(t, network)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, network)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()

	assert.NotEmpty(t, ip, "GetLocalIP should return a non-empty IP address")
	assert.NotNil(t, net.ParseIP(ip), "GetLocalIP should return a valid IP address")

	parsedIP := net.ParseIP(ip)
	assert.False(t, parsedIP.IsLoopback(), "GetLocalIP should not return loopback address")
}
