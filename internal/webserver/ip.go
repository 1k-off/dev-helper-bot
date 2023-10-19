package webserver

import (
	"net"
)

func parseSystemIPs(allowedIPsString, deniedIPsString []string) (*IP, error) {
	for _, addr := range allowedIPsString {
		_, ipNetAllowed, _ := net.ParseCIDR(addr)
		sysIP.Allowed = append(sysIP.Allowed, ipNetAllowed)
	}
	for _, addr := range deniedIPsString {
		ipDenied, _, _ := net.ParseCIDR(addr)
		sysIP.Denied = append(sysIP.Denied, ipDenied)
	}

	return &sysIP, nil
}

func parseClientIP(clientIpString string) (net.IP, error) {
	ip := net.ParseIP(clientIpString)
	if len(ip) == 0 {
		return nil, ErrNotValidIp
	}
	return ip, nil
}

func CheckIfIpAllowed(allowedIPs, deniedIPs []string, ip string) error {
	sysIp, err := parseSystemIPs(allowedIPs, deniedIPs)
	if err != nil {
		return err
	}
	clientIp, err := parseClientIP(ip)
	if err != nil {
		return ErrIpParse
	}
	if !isPrivateIP(clientIp) {
		return ErrNotPrivateIp
	}

	isAllowedIP := false
	for _, network := range sysIp.Allowed {
		if network.Contains(clientIp) {
			if clientIp.To4()[3] != 0 {
				isAllowedIP = true
			} else {
				return ErrNetworkIP
			}
		}
	}
	if !isAllowedIP {
		return ErrNotOfficeIp
	}

	isDeniedIp := false
	for _, dIp := range sysIp.Denied {
		if clientIp.String() == dIp.String() {
			isDeniedIp = true
		}
	}
	if isDeniedIp {
		return ErrIpDenied
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
