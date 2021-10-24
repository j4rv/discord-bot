// +build linux

package main

import (
	"fmt"
	"os/exec"
)

func reboot() {
	cmd := exec.Command("reboot")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not reboot linux: %w", err)
	}
	return nil
}

func shutdown(t time.Duration) error {
	return fmt.Errorf("not implemented")
}

func abortShutdown() error {
	return fmt.Errorf("not implemented")
}
