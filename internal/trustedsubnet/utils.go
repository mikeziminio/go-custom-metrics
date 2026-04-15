package trustedsubnet

import "net"

// ParseCIDR parses a CIDR string and returns the IPNet.
// Returns an error if the CIDR is invalid.
func ParseCIDR(cidr string) (*net.IPNet, error) {
	_, network, err := net.ParseCIDR(cidr)
	return network, err
}

// GetLocalIP returns the local IP address
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		// Берем первый подходящий адрес
		// в целом не противоречит ТЗ, но для реального проекта,
		// естественно, требовалось бы уточнение.
		//
		// В целом это может быть и не из интерфейса IP,
		// а, например, натированный, тогда необходимо его
		// просто прописать в конфигах, а не пытаться вычислять.
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}

	return ""
}
