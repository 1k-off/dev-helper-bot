package webserver

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strconv"
)

type IP struct {
	Allowed []*net.IPNet
	Denied  []net.IP
}

var (
	privateIPBlocks []*net.IPNet
	sysIP           IP
	Debug           bool
)

type Webserver interface {
	Create(c *entities.Domain) error
	Delete(domain string) error
}

func init() {
	debugModeString := os.Getenv("OOOPS_DEBUG")
	d, err := strconv.ParseBool(debugModeString)
	if err != nil {
		Debug = false
		log.Err(err).Msg("Error parsing OOOPS_DEBUG env variable")
	}
	Debug = d

	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

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
