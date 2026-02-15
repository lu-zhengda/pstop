package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DetailedInfo holds extended information about a single process.
type DetailedInfo struct {
	PID       int
	Name      string
	User      string
	CPU       float64
	Mem       float64
	OpenFiles int
	Ports     []int
	Children  []int
}

// GetInfo returns detailed information about a process by PID.
func GetInfo(pid int) (*DetailedInfo, error) {
	info := &DetailedInfo{PID: pid}

	// Get basic info from ps.
	if err := info.fetchBasicInfo(); err != nil {
		return nil, fmt.Errorf("failed to get basic info for PID %d: %w", pid, err)
	}

	// Get open files and ports from lsof.
	info.fetchLsofInfo()

	// Get child processes from pgrep.
	info.fetchChildren()

	return info, nil
}

func (d *DetailedInfo) fetchBasicInfo() error {
	out, err := exec.Command("ps", "-p", strconv.Itoa(d.PID), "-o", "pid,user,%cpu,rss,comm").Output()
	if err != nil {
		return fmt.Errorf("failed to run ps: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("process %d not found", d.PID)
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return fmt.Errorf("unexpected ps output: %q", lines[1])
	}

	d.User = fields[1]

	cpu, err := strconv.ParseFloat(fields[2], 64)
	if err == nil {
		d.CPU = cpu
	}

	rss, err := strconv.ParseFloat(fields[3], 64)
	if err == nil {
		d.Mem = rss / (16 * 1024 * 1024) * 100.0
	}

	command := strings.Join(fields[4:], " ")
	d.Name = command
	if idx := strings.LastIndex(d.Name, "/"); idx >= 0 {
		d.Name = d.Name[idx+1:]
	}

	return nil
}

func (d *DetailedInfo) fetchLsofInfo() {
	out, err := exec.Command("lsof", "-p", strconv.Itoa(d.PID)).Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(out), "\n")
	fileCount := 0
	portSet := make(map[int]bool)

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fileCount++

		// Look for TCP/UDP listening ports.
		if strings.Contains(line, "TCP") || strings.Contains(line, "UDP") {
			port := extractPort(line)
			if port > 0 {
				portSet[port] = true
			}
		}
	}

	d.OpenFiles = fileCount
	for p := range portSet {
		d.Ports = append(d.Ports, p)
	}
}

func (d *DetailedInfo) fetchChildren() {
	out, err := exec.Command("pgrep", "-P", strconv.Itoa(d.PID)).Output()
	if err != nil {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			d.Children = append(d.Children, pid)
		}
	}
}

func extractPort(line string) int {
	// lsof output has format: ... TCP *:8080 (LISTEN)
	// or: ... TCP localhost:3000 (LISTEN)
	fields := strings.Fields(line)
	for _, f := range fields {
		if idx := strings.LastIndex(f, ":"); idx >= 0 {
			portStr := f[idx+1:]
			port, err := strconv.Atoi(portStr)
			if err == nil && port > 0 && port < 65536 {
				return port
			}
		}
	}
	return 0
}
