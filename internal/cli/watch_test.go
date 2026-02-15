package cli

import (
	"testing"

	"github.com/lu-zhengda/pstop/internal/process"
)

func TestCheckThresholdsNone(t *testing.T) {
	// With thresholds at 0, checkThresholds should return nil (no alert).
	origCPU := watchCPU
	origMem := watchMem
	defer func() {
		watchCPU = origCPU
		watchMem = origMem
	}()

	watchCPU = 0
	watchMem = 0

	alert := checkThresholds()
	if alert != nil {
		t.Errorf("checkThresholds() with zero thresholds should return nil, got %+v", alert)
	}
}

func TestAlertStruct(t *testing.T) {
	alert := Alert{
		Timestamp: "2026-01-01T00:00:00Z",
		Threshold: "cpu",
		Value:     95.5,
		Limit:     80.0,
		Process: process.Info{
			PID:  123,
			Name: "test",
			CPU:  95.5,
		},
	}

	if alert.Threshold != "cpu" {
		t.Errorf("Alert.Threshold = %q, want cpu", alert.Threshold)
	}
	if alert.Value != 95.5 {
		t.Errorf("Alert.Value = %f, want 95.5", alert.Value)
	}
	if alert.Limit != 80.0 {
		t.Errorf("Alert.Limit = %f, want 80.0", alert.Limit)
	}
	if alert.Process.PID != 123 {
		t.Errorf("Alert.Process.PID = %d, want 123", alert.Process.PID)
	}
}
