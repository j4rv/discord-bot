// +build linux

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

func reboot() error {
	cmd := exec.Command("sudo", "reboot")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not reboot linux: %w", err)
	}
	return nil
}

func shutdown(t time.Duration) error {
	cmd := exec.Command("sudo", "shutdown", "+"+strconv.Itoa(int(t.Minutes())))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not shutdown linux: %w", err)
	}
	return nil
}

func abortShutdown() error {
	cmd := exec.Command("sudo", "shutdown", "-c")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not abort linux shutdown: %w", err)
	}
	return nil
}
