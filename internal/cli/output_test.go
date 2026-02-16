package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lu-zhengda/pstop/internal/process"
)

func TestFprintJSON_ProcessInfo(t *testing.T) {
	procs := []process.Info{
		{PID: 1, Name: "init", User: "root", CPU: 0.1, Mem: 0.2, State: "S", Command: "/sbin/init"},
		{PID: 42, Name: "node", User: "user", CPU: 25.5, Mem: 8.3, State: "R", Command: "node server.js"},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, procs); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got []process.Info
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	if got[0].PID != 1 {
		t.Errorf("got[0].PID = %d, want 1", got[0].PID)
	}
	if got[1].Name != "node" {
		t.Errorf("got[1].Name = %q, want %q", got[1].Name, "node")
	}
	if got[1].CPU != 25.5 {
		t.Errorf("got[1].CPU = %f, want 25.5", got[1].CPU)
	}
}

func TestFprintJSON_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	if err := fprintJSON(&buf, []process.Info{}); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got []process.Info
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if got == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("got %d items, want 0", len(got))
	}
}

func TestFprintJSON_KillResult(t *testing.T) {
	result := KillResult{OK: true, PID: 123, Signal: "SIGTERM"}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, result); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got KillResult
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !got.OK {
		t.Error("got OK = false, want true")
	}
	if got.PID != 123 {
		t.Errorf("got PID = %d, want 123", got.PID)
	}
	if got.Signal != "SIGTERM" {
		t.Errorf("got Signal = %q, want SIGTERM", got.Signal)
	}
}

func TestFprintJSON_Alert(t *testing.T) {
	alert := Alert{
		Timestamp: "2026-01-01T00:00:00Z",
		Threshold: "cpu",
		Value:     95.5,
		Limit:     80.0,
		Process: process.Info{
			PID:  456,
			Name: "heavy",
			CPU:  95.5,
			User: "user",
		},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, alert); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got Alert
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if got.Threshold != "cpu" {
		t.Errorf("Threshold = %q, want cpu", got.Threshold)
	}
	if got.Value != 95.5 {
		t.Errorf("Value = %f, want 95.5", got.Value)
	}
	if got.Limit != 80.0 {
		t.Errorf("Limit = %f, want 80.0", got.Limit)
	}
	if got.Process.PID != 456 {
		t.Errorf("Process.PID = %d, want 456", got.Process.PID)
	}
}

func TestFprintJSON_DetailedInfo(t *testing.T) {
	info := process.DetailedInfo{
		PID:       789,
		Name:      "nginx",
		User:      "www",
		CPU:       12.3,
		Mem:       4.5,
		OpenFiles: 100,
		Ports:     []int{80, 443},
		Children:  []int{790, 791},
		Connections: []process.Connection{
			{Protocol: "TCP", LocalAddr: "0.0.0.0:80", RemoteAddr: "*", State: "LISTEN"},
		},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, info); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got process.DetailedInfo
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if got.PID != 789 {
		t.Errorf("PID = %d, want 789", got.PID)
	}
	if got.Name != "nginx" {
		t.Errorf("Name = %q, want nginx", got.Name)
	}
	if len(got.Ports) != 2 {
		t.Fatalf("Ports length = %d, want 2", len(got.Ports))
	}
	if got.Ports[0] != 80 || got.Ports[1] != 443 {
		t.Errorf("Ports = %v, want [80, 443]", got.Ports)
	}
	if len(got.Connections) != 1 {
		t.Fatalf("Connections length = %d, want 1", len(got.Connections))
	}
	if got.Connections[0].State != "LISTEN" {
		t.Errorf("Connections[0].State = %q, want LISTEN", got.Connections[0].State)
	}
}

func TestFprintJSON_CrashReport(t *testing.T) {
	reports := []process.CrashReport{
		{
			Timestamp:  "2026-01-15 10:30:00",
			Process:    "Safari",
			PID:        1234,
			ExceptType: "EXC_BAD_ACCESS",
			Signal:     "SIGSEGV",
			Path:       "/tmp/Safari.ips",
			ReportType: "crash",
		},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, reports); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got []process.CrashReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d items, want 1", len(got))
	}
	if got[0].Process != "Safari" {
		t.Errorf("Process = %q, want Safari", got[0].Process)
	}
	if got[0].ReportType != "crash" {
		t.Errorf("ReportType = %q, want crash", got[0].ReportType)
	}
}

func TestFprintJSON_DevGroup(t *testing.T) {
	groups := []process.DevGroup{
		{
			Stack:    "Node.js",
			TotalCPU: 45.2,
			TotalMem: 12.3,
			Processes: []process.Info{
				{PID: 100, Name: "node", CPU: 25.0, Mem: 8.0},
				{PID: 101, Name: "npm", CPU: 20.2, Mem: 4.3},
			},
		},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, groups); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got []process.DevGroup
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d groups, want 1", len(got))
	}
	if got[0].Stack != "Node.js" {
		t.Errorf("Stack = %q, want Node.js", got[0].Stack)
	}
	if got[0].TotalCPU != 45.2 {
		t.Errorf("TotalCPU = %f, want 45.2", got[0].TotalCPU)
	}
	if len(got[0].Processes) != 2 {
		t.Errorf("Processes length = %d, want 2", len(got[0].Processes))
	}
}

func TestFprintJSON_Indentation(t *testing.T) {
	data := map[string]string{"key": "value"}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, data); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	want := "{\n  \"key\": \"value\"\n}\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

func TestFprintJSON_CrashDetail(t *testing.T) {
	detail := process.CrashDetail{
		CrashReport: process.CrashReport{
			Timestamp:  "2026-01-15 10:30:00",
			Process:    "Safari",
			PID:        1234,
			ExceptType: "EXC_BAD_ACCESS",
			Signal:     "SIGSEGV",
			Path:       "/tmp/Safari.ips",
			ReportType: "crash",
		},
		Version:     "17.2",
		OSVersion:   "macOS 15.0",
		CrashThread: 0,
		Backtrace:   []string{"frame0", "frame1"},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, detail); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got process.CrashDetail
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if got.Version != "17.2" {
		t.Errorf("Version = %q, want 17.2", got.Version)
	}
	if got.OSVersion != "macOS 15.0" {
		t.Errorf("OSVersion = %q, want macOS 15.0", got.OSVersion)
	}
	if len(got.Backtrace) != 2 {
		t.Fatalf("Backtrace length = %d, want 2", len(got.Backtrace))
	}
}

func TestFprintJSON_TreeEntry(t *testing.T) {
	entries := []process.FlatTreeEntry{
		{
			Process: process.Info{PID: 1, Name: "init", User: "root"},
			Depth:   0,
		},
		{
			Process: process.Info{PID: 2, Name: "child", User: "root"},
			Depth:   1,
		},
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, entries); err != nil {
		t.Fatalf("fprintJSON() error: %v", err)
	}

	var got []process.FlatTreeEntry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0].Depth != 0 {
		t.Errorf("got[0].Depth = %d, want 0", got[0].Depth)
	}
	if got[1].Depth != 1 {
		t.Errorf("got[1].Depth = %d, want 1", got[1].Depth)
	}
}
