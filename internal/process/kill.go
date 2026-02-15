package process

import (
	"fmt"
	"syscall"
)

// Kill sends SIGTERM (or SIGKILL if force is true) to the given PID.
func Kill(pid int, force bool) error {
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	return KillWithSignal(pid, sig)
}

// KillWithSignal sends the specified signal to the given PID.
func KillWithSignal(pid int, sig syscall.Signal) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}
	if err := syscall.Kill(pid, sig); err != nil {
		return fmt.Errorf("failed to send signal %d to PID %d: %w", sig, pid, err)
	}
	return nil
}
