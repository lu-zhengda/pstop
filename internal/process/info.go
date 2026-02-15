package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

// Connection represents a network connection for a process.
type Connection struct {
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	State      string `json:"state"`
}

// DetailedInfo holds extended information about a single process.
type DetailedInfo struct {
	PID         int               `json:"pid"`
	Name        string            `json:"name"`
	User        string            `json:"user"`
	CPU         float64           `json:"cpu"`
	Mem         float64           `json:"mem"`
	OpenFiles   int               `json:"open_files"`
	Ports       []int             `json:"ports"`
	Children    []int             `json:"children"`
	Connections []Connection      `json:"connections,omitempty"`
	EnvVars     map[string]string `json:"env_vars,omitempty"`
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

	// Get network connections.
	info.fetchConnections()

	// Get environment variables.
	info.fetchEnvVars()

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

func (d *DetailedInfo) fetchConnections() {
	out, err := exec.Command("lsof", "-i", "-P", "-n", "-p", strconv.Itoa(d.PID)).Output()
	if err != nil {
		return
	}
	d.Connections = ParseLsofConnections(string(out))
}

func (d *DetailedInfo) fetchEnvVars() {
	out, err := exec.Command("ps", "eww", "-p", strconv.Itoa(d.PID), "-o", "command=").Output()
	if err != nil {
		return
	}
	d.EnvVars = ParseEnvVars(string(out))
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

// ParseLsofConnections parses the output of `lsof -i -P -n -p <pid>`.
func ParseLsofConnections(output string) []Connection {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var conns []Connection
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// lsof -i output format:
		// COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		// The protocol is in the NODE field (index 7), and the address info is in NAME (index 8).
		protocol := fields[7]
		if protocol != "TCP" && protocol != "UDP" {
			continue
		}

		name := fields[8]
		// Extract state if present (e.g., "(LISTEN)", "(ESTABLISHED)")
		state := ""
		if len(fields) > 9 {
			state = strings.Trim(fields[9], "()")
		}

		// Parse local and remote addresses from NAME field.
		// Format: local->remote or just local (for LISTEN)
		local := name
		remote := ""
		if idx := strings.Index(name, "->"); idx >= 0 {
			local = name[:idx]
			remote = name[idx+2:]
		}

		conns = append(conns, Connection{
			Protocol:   protocol,
			LocalAddr:  local,
			RemoteAddr: remote,
			State:      state,
		})
	}
	return conns
}

// ParseEnvVars parses the output of `ps eww -p <pid> -o command=` to extract environment variables.
func ParseEnvVars(output string) map[string]string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	// The output is: command arg1 arg2 ... ENV1=val1 ENV2=val2
	// We split by spaces and look for KEY=VALUE patterns where KEY looks like an env var.
	parts := strings.Fields(output)
	envVars := make(map[string]string)

	for _, part := range parts {
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		key := part[:idx]
		value := part[idx+1:]
		if isEnvVarKey(key) {
			envVars[key] = value
		}
	}

	if len(envVars) == 0 {
		return nil
	}
	return envVars
}

// isEnvVarKey checks if a string looks like an environment variable key.
// Heuristic: uppercase alphanumeric + underscore, minimum 2 characters.
func isEnvVarKey(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, r := range s {
		if !unicode.IsUpper(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
