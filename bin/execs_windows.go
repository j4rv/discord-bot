// +build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

func reboot() error {
	cmd := exec.Command("shutdown", "-r", "-t", "60")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not reboot windows: %w", err)
	}
	return nil
}

func shutdown(t time.Duration) error {
	cmd := exec.Command("shutdown", "-r", "-t", strconv.Itoa(int(t.Seconds())))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not shutdown windows: %w", err)
	}
	return nil
}

func abortShutdown() error {
	cmd := exec.Command("shutdown", "-a")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not abort windows shutdown: %w", err)
	}
	return nil
}
