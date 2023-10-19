package nginx

import (
	"os/exec"
)

func Reload() error {
	exe := "nginx"
	cmd := exec.Command(exe, "-t")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	cmd = exec.Command(exe, "-s", "reload")
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
