package nginx

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/rs/zerolog/log"
	"html/template"
	"net"
	"os"
	"os/exec"
	"strconv"
)

type IP struct {
	Allowed []*net.IPNet
	Denied  []net.IP
}

var (
	privateIPBlocks []*net.IPNet
	sysIP           IP
	debug           bool
)

func init() {
	debugModeString := os.Getenv("OOOPS_DEBUG")
	d, err := strconv.ParseBool(debugModeString)
	if err != nil {
		debug = false
		log.Err(err).Msg("Error parsing OOOPS_DEBUG env variable")
	}
	debug = d
	if !debug {
		//check if nginx is installed
		if _, err := exec.LookPath("nginx"); err != nil {
			log.Fatal().Err(err).Msg("nginx not installed")
		}
	}

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

	if _, err := os.Stat(configBasePath); os.IsNotExist(err) {
		err := os.Mkdir(configBasePath, os.ModeDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating config directory")
			return
		}
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
		return nil, errNotValidIp
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
		return errIpParse
	}
	if !isPrivateIP(clientIp) {
		return errNotPrivateIp
	}

	isAllowedIP := false
	for _, network := range sysIp.Allowed {
		if network.Contains(clientIp) {
			if clientIp.To4()[3] != 0 {
				isAllowedIP = true
			} else {
				return errNetworkIP
			}
		}
	}
	if !isAllowedIP {
		return errNotOfficeIp
	}

	isDeniedIp := false
	for _, dIp := range sysIp.Denied {
		if clientIp.String() == dIp.String() {
			isDeniedIp = true
		}
	}
	if isDeniedIp {
		return errIpDenied
	}

	return nil
}

func Create(c *entities.Domain) error {
	ba := "Restricted"
	if !c.BasicAuth {
		ba = "off"
	}

	scheme := schemeHttp
	if c.FullSsl {
		scheme = schemeHttps
		if c.Port == "" {
			c.Port = "443"
		}
	} else {
		if c.Port == "" {
			c.Port = "80"
		}
	}

	configData := map[string]string{
		"ip":        c.IP,
		"domain":    c.FQDN,
		"basicauth": ba,
		"scheme":    scheme,
		"port":      c.Port,
	}
	if _, err := os.Stat(configBasePath + "/" + c.FQDN); os.IsNotExist(err) {
		t, err := template.ParseFiles("config/nginx.conf.tpl")
		if err != nil {
			return err
		}
		f, err := os.Create(configBasePath + "/" + c.FQDN)
		if err != nil {
			return err
		}
		err = t.Execute(f, configData)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		if !debug {
			if err := reloadServer(); err != nil {
				return err
			}
		}
		log.Info().Msg(fmt.Sprintf("[nginx] created config. ClientIP: %v, Domain: %s.", c.IP, c.FQDN))
		return nil
	} else {
		return errConfigExists
	}

}

func Delete(domain string) error {
	if err := os.Remove(configBasePath + "/" + domain); err != nil {
		return err
	}
	if !debug {
		if err := reloadServer(); err != nil {
			return err
		}
	}
	log.Info().Msg(fmt.Sprintf("[nginx] deleted config. Domain: %s.", domain))
	return nil
}

func reloadServer() error {
	nginx := "nginx"
	cmd := exec.Command(nginx, "-t")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	cmd = exec.Command(nginx, "-s", "reload")
	_, err = cmd.Output()
	if err != nil {
		return err
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

func TestConfig() error {
	nginx := "nginx"
	cmd := exec.Command(nginx, "-t")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
