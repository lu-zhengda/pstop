package process

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CrashReport holds summary information about a crash, hang, spin, or panic report.
type CrashReport struct {
	Timestamp  string `json:"timestamp"`
	Process    string `json:"process"`
	PID        int    `json:"pid"`
	ExceptType string `json:"exception_type"`
	Signal     string `json:"signal"`
	Path       string `json:"path"`
	ReportType string `json:"report_type"` // crash, hang, spin, panic
}

// CrashDetail holds extended information about a specific crash report.
type CrashDetail struct {
	CrashReport
	Version     string   `json:"version,omitempty"`
	OSVersion   string   `json:"os_version,omitempty"`
	CrashThread int      `json:"crash_thread,omitempty"`
	Backtrace   []string `json:"backtrace,omitempty"`
}

// diagnosticDirs returns the directories to scan for diagnostic reports.
func diagnosticDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{"/Library/Logs/DiagnosticReports"}
	}
	return []string{
		filepath.Join(home, "Library", "Logs", "DiagnosticReports"),
		"/Library/Logs/DiagnosticReports",
	}
}

// ListCrashReports scans DiagnosticReports directories and returns recent reports.
func ListCrashReports(lastDuration string, processFilter string) ([]CrashReport, error) {
	cutoff, err := parseDuration(lastDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration %q: %w", lastDuration, err)
	}

	var reports []CrashReport

	for _, dir := range diagnosticDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Directory may not exist or be inaccessible; skip it.
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			var reportType string
			switch ext {
			case ".ips":
				reportType = "crash"
			case ".hang":
				reportType = "hang"
			case ".spin":
				reportType = "spin"
			case ".panic":
				reportType = "panic"
			default:
				continue
			}

			fullPath := filepath.Join(dir, name)
			report, err := parseCrashSummary(fullPath, reportType)
			if err != nil {
				continue
			}

			// Apply time filter.
			ts, err := parseCrashTimestamp(report.Timestamp)
			if err == nil && ts.Before(cutoff) {
				continue
			}

			// Apply process filter.
			if processFilter != "" && !strings.EqualFold(report.Process, processFilter) {
				continue
			}

			reports = append(reports, report)
		}
	}

	// Sort by timestamp descending (most recent first).
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp > reports[j].Timestamp
	})

	return reports, nil
}

// GetCrashDetail reads and parses a specific crash report file.
func GetCrashDetail(path string) (*CrashDetail, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ips":
		return parseIPSDetail(path)
	case ".hang":
		return parseHangDetail(path)
	case ".spin":
		return parseSpinDetail(path)
	case ".panic":
		return parsePanicDetail(path)
	default:
		return nil, fmt.Errorf("unsupported report format: %s", ext)
	}
}

// parseDuration parses a human-readable duration string like "24h", "7d", "30d".
// Returns the cutoff time (now - duration).
func parseDuration(s string) (time.Time, error) {
	if s == "" {
		s = "7d"
	}

	s = strings.TrimSpace(strings.ToLower(s))
	if len(s) < 2 {
		return time.Time{}, fmt.Errorf("invalid duration: %q", s)
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration number: %w", err)
	}
	if num <= 0 {
		return time.Time{}, fmt.Errorf("duration must be positive: %q", s)
	}

	now := time.Now()
	switch unit {
	case 'h':
		return now.Add(-time.Duration(num) * time.Hour), nil
	case 'd':
		return now.AddDate(0, 0, -num), nil
	default:
		return time.Time{}, fmt.Errorf("unknown duration unit %q (use h or d)", string(unit))
	}
}

// parseCrashTimestamp attempts to parse a timestamp from a crash report.
func parseCrashTimestamp(ts string) (time.Time, error) {
	// .ips files use: "2026-02-08 22:58:16.00 -0500"
	formats := []string{
		"2006-01-02 15:04:05.00 -0700",
		"2006-01-02 15:04:05.0000 -0700",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %q", ts)
}

// ipsHeader represents the first-line JSON header of an .ips file.
type ipsHeader struct {
	AppName   string `json:"app_name"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	BugType   string `json:"bug_type"`
	OSVersion string `json:"os_version"`
}

// ipsBody represents key fields in the JSON body of an .ips file.
type ipsBody struct {
	PID            int `json:"pid"`
	FaultingThread int `json:"faultingThread"`
	Exception      struct {
		Type   string `json:"type"`
		Signal string `json:"signal"`
	} `json:"exception"`
	OSVersion struct {
		Train string `json:"train"`
		Build string `json:"build"`
	} `json:"osVersion"`
	Threads []ipsThread `json:"threads"`
}

// ipsThread represents a thread in the .ips crash report.
type ipsThread struct {
	Triggered bool       `json:"triggered"`
	Frames    []ipsFrame `json:"frames"`
}

// ipsFrame represents a stack frame in the .ips crash report.
type ipsFrame struct {
	Symbol         string `json:"symbol"`
	SymbolLocation int    `json:"symbolLocation"`
	ImageOffset    int    `json:"imageOffset"`
	ImageIndex     int    `json:"imageIndex"`
}

// parseCrashSummary reads the first line(s) of a report file and extracts summary info.
func parseCrashSummary(path string, reportType string) (CrashReport, error) {
	switch reportType {
	case "crash":
		return parseIPSSummary(path)
	case "hang":
		return parseHangSummary(path)
	case "spin":
		return parseSpinSummary(path)
	case "panic":
		return parsePanicSummary(path)
	default:
		return CrashReport{}, fmt.Errorf("unknown report type: %s", reportType)
	}
}

// parseIPSSummary parses the header line of an .ips file to extract summary info.
func parseIPSSummary(path string) (CrashReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return CrashReport{}, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read header line.
	if !scanner.Scan() {
		return CrashReport{}, fmt.Errorf("empty file: %s", path)
	}
	headerLine := scanner.Text()

	var header ipsHeader
	if err := json.Unmarshal([]byte(headerLine), &header); err != nil {
		return CrashReport{}, fmt.Errorf("failed to parse .ips header: %w", err)
	}

	processName := header.Name
	if processName == "" {
		processName = header.AppName
	}

	// Read body to get PID and exception info.
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	report := CrashReport{
		Timestamp:  header.Timestamp,
		Process:    processName,
		Path:       path,
		ReportType: "crash",
	}

	if len(bodyLines) > 0 {
		bodyStr := strings.Join(bodyLines, "\n")
		var body ipsBody
		if err := json.Unmarshal([]byte(bodyStr), &body); err == nil {
			report.PID = body.PID
			report.ExceptType = body.Exception.Type
			report.Signal = body.Exception.Signal
		}
	}

	return report, nil
}

// parseIPSDetail parses an .ips file in full detail.
func parseIPSDetail(path string) (*CrashDetail, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Read header line.
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file: %s", path)
	}
	headerLine := scanner.Text()

	var header ipsHeader
	if err := json.Unmarshal([]byte(headerLine), &header); err != nil {
		return nil, fmt.Errorf("failed to parse .ips header: %w", err)
	}

	// Read body.
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	processName := header.Name
	if processName == "" {
		processName = header.AppName
	}

	detail := &CrashDetail{
		CrashReport: CrashReport{
			Timestamp:  header.Timestamp,
			Process:    processName,
			Path:       path,
			ReportType: "crash",
		},
		OSVersion: header.OSVersion,
	}

	if len(bodyLines) > 0 {
		bodyStr := strings.Join(bodyLines, "\n")
		var body ipsBody
		if err := json.Unmarshal([]byte(bodyStr), &body); err == nil {
			detail.PID = body.PID
			detail.ExceptType = body.Exception.Type
			detail.Signal = body.Exception.Signal
			detail.CrashThread = body.FaultingThread

			if body.OSVersion.Train != "" {
				detail.OSVersion = body.OSVersion.Train
				if body.OSVersion.Build != "" {
					detail.OSVersion += " (" + body.OSVersion.Build + ")"
				}
			}

			// Extract backtrace from the faulting thread.
			if body.FaultingThread >= 0 && body.FaultingThread < len(body.Threads) {
				thread := body.Threads[body.FaultingThread]
				for _, frame := range thread.Frames {
					if frame.Symbol != "" {
						detail.Backtrace = append(detail.Backtrace,
							fmt.Sprintf("%s+%d", frame.Symbol, frame.SymbolLocation))
					} else {
						detail.Backtrace = append(detail.Backtrace,
							fmt.Sprintf("0x%x", frame.ImageOffset))
					}
				}
			}
		}
	}

	return detail, nil
}

// hangHeaderRegex matches key-value pairs in .hang file headers.
var hangProcessRegex = regexp.MustCompile(`(?i)^Process:\s+(.+?)(?:\s+\[(\d+)\])?$`)
var hangDateRegex = regexp.MustCompile(`(?i)^Date\/Time:\s+(.+)$`)
var hangDurationRegex = regexp.MustCompile(`(?i)^Duration:\s+(.+)$`)
var hangOSRegex = regexp.MustCompile(`(?i)^OS Version:\s+(.+)$`)

// parseHangSummary extracts summary info from a .hang file.
func parseHangSummary(path string) (CrashReport, error) {
	return parseTextReportSummary(path, "hang")
}

// parseSpinSummary extracts summary info from a .spin file.
func parseSpinSummary(path string) (CrashReport, error) {
	return parseTextReportSummary(path, "spin")
}

// parsePanicSummary extracts summary info from a .panic file.
func parsePanicSummary(path string) (CrashReport, error) {
	return parseTextReportSummary(path, "panic")
}

// parseTextReportSummary parses text-based report files (.hang, .spin, .panic).
func parseTextReportSummary(path string, reportType string) (CrashReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return CrashReport{}, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	report := CrashReport{
		Path:       path,
		ReportType: reportType,
	}

	scanner := bufio.NewScanner(f)
	linesRead := 0
	for scanner.Scan() && linesRead < 50 {
		line := scanner.Text()
		linesRead++

		if m := hangProcessRegex.FindStringSubmatch(line); m != nil {
			report.Process = strings.TrimSpace(m[1])
			if len(m) > 2 && m[2] != "" {
				pid, err := strconv.Atoi(m[2])
				if err == nil {
					report.PID = pid
				}
			}
		}

		if m := hangDateRegex.FindStringSubmatch(line); m != nil {
			report.Timestamp = strings.TrimSpace(m[1])
		}
	}

	return report, nil
}

// parseHangDetail parses a .hang file in full detail.
func parseHangDetail(path string) (*CrashDetail, error) {
	return parseTextReportDetail(path, "hang")
}

// parseSpinDetail parses a .spin file in full detail.
func parseSpinDetail(path string) (*CrashDetail, error) {
	return parseTextReportDetail(path, "spin")
}

// parsePanicDetail parses a .panic file in full detail.
func parsePanicDetail(path string) (*CrashDetail, error) {
	return parseTextReportDetail(path, "panic")
}

// parseTextReportDetail parses a text-based report file in full detail.
func parseTextReportDetail(path string, reportType string) (*CrashDetail, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	detail := &CrashDetail{
		CrashReport: CrashReport{
			Path:       path,
			ReportType: reportType,
		},
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := hangProcessRegex.FindStringSubmatch(line); m != nil {
			detail.Process = strings.TrimSpace(m[1])
			if len(m) > 2 && m[2] != "" {
				pid, err := strconv.Atoi(m[2])
				if err == nil {
					detail.PID = pid
				}
			}
		}

		if m := hangDateRegex.FindStringSubmatch(line); m != nil {
			detail.Timestamp = strings.TrimSpace(m[1])
		}

		if m := hangDurationRegex.FindStringSubmatch(line); m != nil {
			detail.ExceptType = reportType + " (" + strings.TrimSpace(m[1]) + ")"
		}

		if m := hangOSRegex.FindStringSubmatch(line); m != nil {
			detail.OSVersion = strings.TrimSpace(m[1])
		}
	}

	return detail, nil
}
