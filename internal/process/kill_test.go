package process

import (
	"syscall"
	"testing"
)

func TestKillInvalidPID(t *testing.T) {
	err := Kill(0, false)
	if err == nil {
		t.Error("Kill(0) should return an error")
	}

	err = Kill(-1, false)
	if err == nil {
		t.Error("Kill(-1) should return an error")
	}
}

func TestKillWithSignalInvalidPID(t *testing.T) {
	err := KillWithSignal(0, syscall.SIGTERM)
	if err == nil {
		t.Error("KillWithSignal(0) should return an error")
	}
}

func TestKillNonExistentProcess(t *testing.T) {
	// PID 999999 is very unlikely to exist.
	err := Kill(999999, false)
	if err == nil {
		t.Error("Kill(999999) should return an error for non-existent process")
	}
}

func TestKillWithSignalNonExistentProcess(t *testing.T) {
	err := KillWithSignal(999999, syscall.SIGTERM)
	if err == nil {
		t.Error("KillWithSignal(999999) should return an error for non-existent process")
	}
}
