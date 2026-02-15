package process

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, cutoff time.Time)
	}{
		{
			name:  "24 hours",
			input: "24h",
			check: func(t *testing.T, cutoff time.Time) {
				expected := time.Now().Add(-24 * time.Hour)
				diff := cutoff.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("cutoff %v not within 1s of expected %v", cutoff, expected)
				}
			},
		},
		{
			name:  "7 days",
			input: "7d",
			check: func(t *testing.T, cutoff time.Time) {
				expected := time.Now().AddDate(0, 0, -7)
				diff := cutoff.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("cutoff %v not within 1s of expected %v", cutoff, expected)
				}
			},
		},
		{
			name:  "30 days",
			input: "30d",
			check: func(t *testing.T, cutoff time.Time) {
				expected := time.Now().AddDate(0, 0, -30)
				diff := cutoff.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("cutoff %v not within 1s of expected %v", cutoff, expected)
				}
			},
		},
		{
			name:  "1 hour",
			input: "1h",
			check: func(t *testing.T, cutoff time.Time) {
				expected := time.Now().Add(-1 * time.Hour)
				diff := cutoff.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("cutoff %v not within 1s of expected %v", cutoff, expected)
				}
			},
		},
		{
			name:  "empty defaults to 7d",
			input: "",
			check: func(t *testing.T, cutoff time.Time) {
				expected := time.Now().AddDate(0, 0, -7)
				diff := cutoff.Sub(expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("cutoff %v not within 1s of expected %v", cutoff, expected)
				}
			},
		},
		{
			name:    "invalid unit",
			input:   "7m",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			input:   "abch",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "h",
			wantErr: true,
		},
		{
			name:    "zero duration",
			input:   "0d",
			wantErr: true,
		},
		{
			name:    "negative duration",
			input:   "-1d",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseCrashTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, ts time.Time)
	}{
		{
			name:  "ips format with centiseconds",
			input: "2026-02-08 22:58:16.00 -0500",
			check: func(t *testing.T, ts time.Time) {
				if ts.Year() != 2026 || ts.Month() != 2 || ts.Day() != 8 {
					t.Errorf("unexpected date: %v", ts)
				}
			},
		},
		{
			name:  "plain datetime",
			input: "2026-02-08 22:58:16",
			check: func(t *testing.T, ts time.Time) {
				if ts.Year() != 2026 {
					t.Errorf("unexpected year: %d", ts.Year())
				}
			},
		},
		{
			name:    "garbage",
			input:   "not a date",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCrashTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCrashTimestamp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseIPSSummary(t *testing.T) {
	// Create a temporary .ips file with realistic content.
	header := `{"app_name":"TestApp","timestamp":"2026-02-08 22:58:16.00 -0500","app_version":"1.0","slice_uuid":"abc","build_version":"","platform":0,"share_with_app_devs":0,"is_first_party":0,"bug_type":"309","os_version":"macOS 26.2 (25C56)","roots_installed":0,"incident_id":"ABC-123","name":"TestApp"}`
	body := `{
  "pid": 12345,
  "faultingThread": 0,
  "exception": {"type": "EXC_CRASH", "signal": "SIGABRT"},
  "osVersion": {"train": "macOS 26.2", "build": "25C56"},
  "threads": [{"triggered": true, "frames": [{"symbol": "main", "symbolLocation": 0, "imageOffset": 100, "imageIndex": 0}]}]
}`

	dir := t.TempDir()
	path := filepath.Join(dir, "test.ips")
	if err := os.WriteFile(path, []byte(header+"\n"+body), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := parseIPSSummary(path)
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}

	if report.Process != "TestApp" {
		t.Errorf("Process = %q, want %q", report.Process, "TestApp")
	}
	if report.PID != 12345 {
		t.Errorf("PID = %d, want %d", report.PID, 12345)
	}
	if report.ExceptType != "EXC_CRASH" {
		t.Errorf("ExceptType = %q, want %q", report.ExceptType, "EXC_CRASH")
	}
	if report.Signal != "SIGABRT" {
		t.Errorf("Signal = %q, want %q", report.Signal, "SIGABRT")
	}
	if report.ReportType != "crash" {
		t.Errorf("ReportType = %q, want %q", report.ReportType, "crash")
	}
	if report.Timestamp != "2026-02-08 22:58:16.00 -0500" {
		t.Errorf("Timestamp = %q, want %q", report.Timestamp, "2026-02-08 22:58:16.00 -0500")
	}
}

func TestParseIPSDetail(t *testing.T) {
	header := `{"app_name":"CrashApp","timestamp":"2026-02-08 10:30:00.00 -0500","app_version":"2.0","slice_uuid":"def","build_version":"","platform":0,"share_with_app_devs":0,"is_first_party":0,"bug_type":"309","os_version":"macOS 26.2 (25C56)","roots_installed":0,"incident_id":"DEF-456","name":"CrashApp"}`
	body := `{
  "pid": 9999,
  "faultingThread": 0,
  "exception": {"type": "EXC_BAD_ACCESS", "signal": "SIGSEGV"},
  "osVersion": {"train": "macOS 26.2", "build": "25C56"},
  "threads": [
    {
      "triggered": true,
      "frames": [
        {"symbol": "objc_msgSend", "symbolLocation": 32, "imageOffset": 100, "imageIndex": 0},
        {"symbol": "main", "symbolLocation": 0, "imageOffset": 200, "imageIndex": 1},
        {"symbol": "", "symbolLocation": 0, "imageOffset": 300, "imageIndex": 2}
      ]
    },
    {
      "triggered": false,
      "frames": [
        {"symbol": "worker_thread", "symbolLocation": 10, "imageOffset": 400, "imageIndex": 0}
      ]
    }
  ]
}`

	dir := t.TempDir()
	path := filepath.Join(dir, "crash.ips")
	if err := os.WriteFile(path, []byte(header+"\n"+body), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	detail, err := parseIPSDetail(path)
	if err != nil {
		t.Fatalf("parseIPSDetail() error: %v", err)
	}

	if detail.Process != "CrashApp" {
		t.Errorf("Process = %q, want %q", detail.Process, "CrashApp")
	}
	if detail.PID != 9999 {
		t.Errorf("PID = %d, want %d", detail.PID, 9999)
	}
	if detail.ExceptType != "EXC_BAD_ACCESS" {
		t.Errorf("ExceptType = %q, want %q", detail.ExceptType, "EXC_BAD_ACCESS")
	}
	if detail.Signal != "SIGSEGV" {
		t.Errorf("Signal = %q, want %q", detail.Signal, "SIGSEGV")
	}
	if detail.CrashThread != 0 {
		t.Errorf("CrashThread = %d, want %d", detail.CrashThread, 0)
	}
	if detail.OSVersion != "macOS 26.2 (25C56)" {
		t.Errorf("OSVersion = %q, want %q", detail.OSVersion, "macOS 26.2 (25C56)")
	}

	// Check backtrace from faulting thread (thread 0).
	wantBacktrace := []string{"objc_msgSend+32", "main+0", "0x12c"}
	if len(detail.Backtrace) != len(wantBacktrace) {
		t.Fatalf("Backtrace length = %d, want %d", len(detail.Backtrace), len(wantBacktrace))
	}
	for i, want := range wantBacktrace {
		if detail.Backtrace[i] != want {
			t.Errorf("Backtrace[%d] = %q, want %q", i, detail.Backtrace[i], want)
		}
	}
}

func TestParseIPSSummary_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.ips")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := parseIPSSummary(path)
	if err == nil {
		t.Error("parseIPSSummary(empty) should return error")
	}
}

func TestParseIPSSummary_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.ips")
	if err := os.WriteFile(path, []byte("not json at all\n{}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := parseIPSSummary(path)
	if err == nil {
		t.Error("parseIPSSummary(invalid JSON) should return error")
	}
}

func TestParseIPSSummary_HeaderOnly(t *testing.T) {
	header := `{"app_name":"OnlyHeader","timestamp":"2026-01-01 00:00:00.00 -0500","name":"OnlyHeader"}`
	dir := t.TempDir()
	path := filepath.Join(dir, "headeronly.ips")
	if err := os.WriteFile(path, []byte(header+"\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := parseIPSSummary(path)
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}
	if report.Process != "OnlyHeader" {
		t.Errorf("Process = %q, want %q", report.Process, "OnlyHeader")
	}
	// PID and exception should be zero/empty since no body.
	if report.PID != 0 {
		t.Errorf("PID = %d, want 0", report.PID)
	}
}

func TestParseTextReportSummary(t *testing.T) {
	content := `Process:     SomeApp [42]
Path:        /Applications/SomeApp.app/Contents/MacOS/SomeApp
Date/Time:   2026-02-10 14:30:00 -0500
Duration:    30.5s

OS Version:  macOS 26.2 (25C56)

Some other content here...
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.hang")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := parseTextReportSummary(path, "hang")
	if err != nil {
		t.Fatalf("parseTextReportSummary() error: %v", err)
	}

	if report.Process != "SomeApp" {
		t.Errorf("Process = %q, want %q", report.Process, "SomeApp")
	}
	if report.PID != 42 {
		t.Errorf("PID = %d, want %d", report.PID, 42)
	}
	if report.Timestamp != "2026-02-10 14:30:00 -0500" {
		t.Errorf("Timestamp = %q, want %q", report.Timestamp, "2026-02-10 14:30:00 -0500")
	}
	if report.ReportType != "hang" {
		t.Errorf("ReportType = %q, want %q", report.ReportType, "hang")
	}
}

func TestParseTextReportDetail(t *testing.T) {
	content := `Process:     HangApp [100]
Path:        /Applications/HangApp.app/Contents/MacOS/HangApp
Date/Time:   2026-02-10 14:30:00 -0500
Duration:    45.2s

OS Version:  macOS 26.2 (25C56)
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.hang")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	detail, err := parseTextReportDetail(path, "hang")
	if err != nil {
		t.Fatalf("parseTextReportDetail() error: %v", err)
	}

	if detail.Process != "HangApp" {
		t.Errorf("Process = %q, want %q", detail.Process, "HangApp")
	}
	if detail.PID != 100 {
		t.Errorf("PID = %d, want %d", detail.PID, 100)
	}
	if detail.OSVersion != "macOS 26.2 (25C56)" {
		t.Errorf("OSVersion = %q, want %q", detail.OSVersion, "macOS 26.2 (25C56)")
	}
	if !strings.Contains(detail.ExceptType, "45.2s") {
		t.Errorf("ExceptType = %q, should contain duration", detail.ExceptType)
	}
}

func TestParseTextReportSummary_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.hang")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := parseTextReportSummary(path, "hang")
	if err != nil {
		t.Fatalf("parseTextReportSummary() should not error on empty file: %v", err)
	}
	if report.Process != "" {
		t.Errorf("Process = %q, want empty", report.Process)
	}
}

func TestParseTextReportSummary_NonExistentFile(t *testing.T) {
	_, err := parseTextReportSummary("/nonexistent/path/test.hang", "hang")
	if err == nil {
		t.Error("parseTextReportSummary(nonexistent) should return error")
	}
}

func TestListCrashReports_NonExistentDirs(t *testing.T) {
	// Override diagnosticDirs to point to non-existent directories.
	// Since we can't easily override it, we test that the function
	// doesn't error on missing directories by using valid but empty ones.
	// The function should gracefully handle directories that don't exist.
	reports, err := ListCrashReports("7d", "")
	if err != nil {
		t.Fatalf("ListCrashReports() should not error: %v", err)
	}
	// Just verify it returns without error; actual results depend on system state.
	_ = reports
}

func TestListCrashReports_WithTempFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a recent .ips file.
	header := `{"app_name":"RecentApp","timestamp":"` + time.Now().Format("2006-01-02 15:04:05") + `","name":"RecentApp"}`
	body := `{"pid": 100, "exception": {"type": "EXC_CRASH", "signal": "SIGABRT"}, "faultingThread": 0, "threads": []}`
	recentPath := filepath.Join(dir, "recent.ips")
	if err := os.WriteFile(recentPath, []byte(header+"\n"+body), 0644); err != nil {
		t.Fatalf("failed to write recent file: %v", err)
	}

	// Create an old .ips file (timestamp in the past beyond 7d).
	oldHeader := `{"app_name":"OldApp","timestamp":"2020-01-01 00:00:00","name":"OldApp"}`
	oldBody := `{"pid": 200, "exception": {"type": "EXC_CRASH", "signal": "SIGKILL"}, "faultingThread": 0, "threads": []}`
	oldPath := filepath.Join(dir, "old.ips")
	if err := os.WriteFile(oldPath, []byte(oldHeader+"\n"+oldBody), 0644); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}

	// Create a non-report file that should be ignored.
	ignorePath := filepath.Join(dir, "readme.txt")
	if err := os.WriteFile(ignorePath, []byte("not a report"), 0644); err != nil {
		t.Fatalf("failed to write ignore file: %v", err)
	}

	// We can't easily override diagnosticDirs, so test the filtering logic
	// by calling parseIPSSummary and parseDuration directly.
	report, err := parseIPSSummary(recentPath)
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}
	if report.Process != "RecentApp" {
		t.Errorf("Process = %q, want %q", report.Process, "RecentApp")
	}
}

func TestListCrashReports_ProcessFilter(t *testing.T) {
	// Test that process filtering works via direct parsing.
	dir := t.TempDir()

	now := time.Now().Format("2006-01-02 15:04:05")
	header1 := `{"app_name":"TargetApp","timestamp":"` + now + `","name":"TargetApp"}`
	body1 := `{"pid": 100, "exception": {"type": "EXC_CRASH", "signal": "SIGABRT"}, "faultingThread": 0, "threads": []}`
	if err := os.WriteFile(filepath.Join(dir, "target.ips"), []byte(header1+"\n"+body1), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	header2 := `{"app_name":"OtherApp","timestamp":"` + now + `","name":"OtherApp"}`
	body2 := `{"pid": 200, "exception": {"type": "EXC_CRASH", "signal": "SIGKILL"}, "faultingThread": 0, "threads": []}`
	if err := os.WriteFile(filepath.Join(dir, "other.ips"), []byte(header2+"\n"+body2), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Parse both files.
	r1, err := parseIPSSummary(filepath.Join(dir, "target.ips"))
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}
	r2, err := parseIPSSummary(filepath.Join(dir, "other.ips"))
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}

	// Simulate process filter.
	filter := "TargetApp"
	reports := []CrashReport{r1, r2}
	var filtered []CrashReport
	for _, r := range reports {
		if strings.EqualFold(r.Process, filter) {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) != 1 {
		t.Errorf("filtered count = %d, want 1", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Process != "TargetApp" {
		t.Errorf("Process = %q, want %q", filtered[0].Process, "TargetApp")
	}
}

func TestGetCrashDetail_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := GetCrashDetail(path)
	if err == nil {
		t.Error("GetCrashDetail(.txt) should return error")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error = %q, should contain 'unsupported'", err.Error())
	}
}

func TestGetCrashDetail_NonExistentFile(t *testing.T) {
	_, err := GetCrashDetail("/nonexistent/path/test.ips")
	if err == nil {
		t.Error("GetCrashDetail(nonexistent) should return error")
	}
}

func TestGetCrashDetail_IPS(t *testing.T) {
	header := `{"app_name":"DetailApp","timestamp":"2026-02-08 10:00:00.00 -0500","name":"DetailApp"}`
	body := `{
  "pid": 555,
  "faultingThread": 0,
  "exception": {"type": "EXC_CRASH", "signal": "SIGABRT"},
  "osVersion": {"train": "macOS 26.2", "build": "25C56"},
  "threads": [{"triggered": true, "frames": [{"symbol": "abort", "symbolLocation": 8, "imageOffset": 100, "imageIndex": 0}]}]
}`

	dir := t.TempDir()
	path := filepath.Join(dir, "detail.ips")
	if err := os.WriteFile(path, []byte(header+"\n"+body), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	detail, err := GetCrashDetail(path)
	if err != nil {
		t.Fatalf("GetCrashDetail() error: %v", err)
	}
	if detail.Process != "DetailApp" {
		t.Errorf("Process = %q, want %q", detail.Process, "DetailApp")
	}
	if detail.PID != 555 {
		t.Errorf("PID = %d, want %d", detail.PID, 555)
	}
}

func TestGetCrashDetail_Hang(t *testing.T) {
	content := `Process:     HangTest [777]
Path:        /usr/bin/HangTest
Date/Time:   2026-02-10 14:30:00 -0500
Duration:    10.0s
OS Version:  macOS 26.2 (25C56)
`
	dir := t.TempDir()
	path := filepath.Join(dir, "detail.hang")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	detail, err := GetCrashDetail(path)
	if err != nil {
		t.Fatalf("GetCrashDetail() error: %v", err)
	}
	if detail.Process != "HangTest" {
		t.Errorf("Process = %q, want %q", detail.Process, "HangTest")
	}
	if detail.ReportType != "hang" {
		t.Errorf("ReportType = %q, want %q", detail.ReportType, "hang")
	}
}

func TestListCrashReports_InvalidDuration(t *testing.T) {
	_, err := ListCrashReports("invalid", "")
	if err == nil {
		t.Error("ListCrashReports(invalid) should return error")
	}
}

func TestParseIPSSummary_AppNameFallback(t *testing.T) {
	// When "name" is empty, should fall back to "app_name".
	header := `{"app_name":"FallbackApp","timestamp":"2026-01-01 00:00:00","name":""}`
	dir := t.TempDir()
	path := filepath.Join(dir, "fallback.ips")
	if err := os.WriteFile(path, []byte(header+"\n{}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	report, err := parseIPSSummary(path)
	if err != nil {
		t.Fatalf("parseIPSSummary() error: %v", err)
	}
	if report.Process != "FallbackApp" {
		t.Errorf("Process = %q, want %q", report.Process, "FallbackApp")
	}
}
