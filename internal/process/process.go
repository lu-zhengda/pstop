package process

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Info holds basic process information.
type Info struct {
	PID     int
	PPID    int
	Name    string
	CPU     float64
	Mem     float64
	User    string
	State   string
	Command string
}

// List returns all running processes.
func List() ([]Info, error) {
	out, err := exec.Command("ps", "-eo", "pid,ppid,user,stat,%cpu,rss,comm").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}
	return ParsePSOutput(string(out))
}

// Top returns the top N processes sorted by CPU usage.
func Top(n int) ([]Info, error) {
	out, err := exec.Command("ps", "-eo", "pid,ppid,user,stat,%cpu,rss,comm", "-r").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get top processes: %w", err)
	}
	procs, err := ParsePSOutput(string(out))
	if err != nil {
		return nil, err
	}
	if n > 0 && n < len(procs) {
		procs = procs[:n]
	}
	return procs, nil
}

// Find searches processes by name or command substring.
func Find(query string) ([]Info, error) {
	procs, err := List()
	if err != nil {
		return nil, fmt.Errorf("failed to find processes: %w", err)
	}
	query = strings.ToLower(query)
	var result []Info
	for _, p := range procs {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Command), query) {
			result = append(result, p)
		}
	}
	return result, nil
}

// ParsePSOutput parses the output of `ps -eo pid,ppid,user,stat,%cpu,rss,comm`.
func ParsePSOutput(output string) ([]Info, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, nil
	}

	var procs []Info
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p, err := parsePSLine(line)
		if err != nil {
			continue // skip unparseable lines
		}
		procs = append(procs, p)
	}
	return procs, nil
}

func parsePSLine(line string) (Info, error) {
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return Info{}, fmt.Errorf("not enough fields: %q", line)
	}

	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return Info{}, fmt.Errorf("failed to parse PID: %w", err)
	}

	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return Info{}, fmt.Errorf("failed to parse PPID: %w", err)
	}

	user := fields[2]
	state := fields[3]

	cpu, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return Info{}, fmt.Errorf("failed to parse CPU: %w", err)
	}

	rss, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return Info{}, fmt.Errorf("failed to parse RSS: %w", err)
	}
	// Convert RSS (KB) to approximate percentage of total memory.
	// Use a rough estimate: assume 16GB total.
	mem := rss / (16 * 1024 * 1024) * 100.0

	// Command is everything from field 6 onwards.
	command := strings.Join(fields[6:], " ")
	// Name is the basename of the executable (first field only).
	name := fields[6]
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	return Info{
		PID:     pid,
		PPID:    ppid,
		Name:    name,
		CPU:     cpu,
		Mem:     mem,
		User:    user,
		State:   state,
		Command: command,
	}, nil
}

// Sort sorts a slice of Info by the given field.
func Sort(procs []Info, field string) {
	sort.Slice(procs, func(i, j int) bool {
		switch field {
		case "mem":
			return procs[i].Mem > procs[j].Mem
		case "pid":
			return procs[i].PID < procs[j].PID
		case "name":
			return strings.ToLower(procs[i].Name) < strings.ToLower(procs[j].Name)
		default: // "cpu"
			return procs[i].CPU > procs[j].CPU
		}
	})
}
