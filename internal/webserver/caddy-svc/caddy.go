package caddy_svc

import (
	"os/exec"
)

func Reload() error {
	exe := "caddy"
	cmd := exec.Command(exe, "validate", "--config", "/etc/caddy/Caddyfile")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	cmd = exec.Command(exe, "reload", "--config", "/etc/caddy/Caddyfile")
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
