package nginx

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"os"
	"os/exec"
)

type SystemIp struct {
	AllowedIPs []*net.IPNet
	DeniedIPs []net.IP
}

var (
	privateIPBlocks []*net.IPNet
	sysIP SystemIp
)

func init() {
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
		os.Mkdir(configBasePath, os.ModeDir)
	}
}

func parseSystemIPs(allowedIPsString, deniedIPsString []string) (*SystemIp, error) {
	for _, addr := range allowedIPsString {
		_, ipNetAllowed, _ := net.ParseCIDR(addr)
		sysIP.AllowedIPs = append(sysIP.AllowedIPs, ipNetAllowed)
	}
	for _, addr := range deniedIPsString {
		ipDenied, _, _ := net.ParseCIDR(addr)
		sysIP.DeniedIPs = append(sysIP.DeniedIPs, ipDenied)
	}

	return &sysIP, nil
}

func parseClientIP (clientIpString string) (net.IP, error) {
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
	if ! isPrivateIP(clientIp) {
		return errNotPrivateIp
	}

	isAllowedIP := false
	for _, network := range sysIp.AllowedIPs {
		if network.Contains(clientIp) {
			if clientIp.To4()[3] != 0 {
				isAllowedIP = true
			} else {
				return errNetworkIP
			}
		}
	}
	if ! isAllowedIP {
		return errNotOfficeIp
	}

	isDeniedIp := false
	for _, dIp := range sysIp.DeniedIPs {
		if clientIp.String() == dIp.String() {
			isDeniedIp = true
		}
	}
	if isDeniedIp {
		return errIpDenied
	}

	return nil
}

func Create(clientIp, domain string, basicauth bool) error {
	ba := "Restricted"
	if !basicauth {
		ba = "off"
	}

	configData := map[string]string {
		"ip": clientIp,
		"domain": domain,
		"basicauth": ba,
	}
	if _, err := os.Stat(configBasePath + "/" + domain); os.IsNotExist(err) {
		t, err := template.ParseFiles("data/nginx.conf.tpl")
		if err != nil {
			return err
		}
		f, err := os.Create(configBasePath + "/" + domain)
		if err != nil {
			return err
		}
		err = t.Execute(f, configData)
		if err != nil {
			return err
		}
		f.Close()
		//if err := reloadServer(); err != nil {
		//	return err
		//}
		log.Println(fmt.Sprintf("[INFO] Created config. ClientIP: %v, Domain: %s.", clientIp, domain))
		return nil
	} else {
		return errConfigExists
	}

}

func Delete(domain string) error {
	if err := os.Remove(configBasePath + "/" + domain); err != nil {
		return err
	}
	//if err := reloadServer(); err != nil {
	//	return err
	//}
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
