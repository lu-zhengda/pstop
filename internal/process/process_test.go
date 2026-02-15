package process

import (
	"strings"
	"testing"
)

func TestParsePSOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name: "typical ps output",
			input: `  PID  PPID USER             STAT  %CPU   RSS COMMAND
    1     0 root             Ss     0.0  1234 /sbin/launchd
  501     1 zhengda          S      2.5  5678 /usr/bin/some_app
  502   501 zhengda          R     15.3 10240 /usr/local/bin/node server.js`,
			want:    3,
			wantErr: false,
		},
		{
			name:    "empty output",
			input:   "",
			want:    0,
			wantErr: false,
		},
		{
			name: "header only",
			input: `  PID  PPID USER             STAT  %CPU   RSS COMMAND`,
			want:    0,
			wantErr: false,
		},
		{
			name: "malformed line is skipped",
			input: `  PID  PPID USER             STAT  %CPU   RSS COMMAND
    1     0 root             Ss     0.0  1234 /sbin/launchd
badline
  502   501 zhengda          R     15.3 10240 /usr/local/bin/node`,
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePSOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePSOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("ParsePSOutput() returned %d processes, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParsePSLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantPID  int
		wantPPID int
		wantUser string
		wantCPU  float64
		wantName string
		wantErr  bool
	}{
		{
			name:     "standard process",
			line:     "  501     1 zhengda          S      2.5  5678 /usr/bin/some_app",
			wantPID:  501,
			wantPPID: 1,
			wantUser: "zhengda",
			wantCPU:  2.5,
			wantName: "some_app",
			wantErr:  false,
		},
		{
			name:     "process with spaces in command",
			line:     "  502   501 zhengda          R     15.3 10240 /usr/local/bin/node server.js",
			wantPID:  502,
			wantPPID: 501,
			wantUser: "zhengda",
			wantCPU:  15.3,
			wantName: "node",
			wantErr:  false,
		},
		{
			name:     "root process",
			line:     "    1     0 root             Ss     0.0  1234 /sbin/launchd",
			wantPID:  1,
			wantPPID: 0,
			wantUser: "root",
			wantCPU:  0.0,
			wantName: "launchd",
			wantErr:  false,
		},
		{
			name:    "too few fields",
			line:    "501 1 user",
			wantErr: true,
		},
		{
			name:    "non-numeric PID",
			line:    "abc 1 user S 0.0 1234 /bin/sh",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePSLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePSLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.PID != tt.wantPID {
				t.Errorf("PID = %d, want %d", got.PID, tt.wantPID)
			}
			if got.PPID != tt.wantPPID {
				t.Errorf("PPID = %d, want %d", got.PPID, tt.wantPPID)
			}
			if got.User != tt.wantUser {
				t.Errorf("User = %q, want %q", got.User, tt.wantUser)
			}
			if got.CPU != tt.wantCPU {
				t.Errorf("CPU = %f, want %f", got.CPU, tt.wantCPU)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

func TestSort(t *testing.T) {
	procs := []Info{
		{PID: 1, Name: "alpha", CPU: 5.0, Mem: 1.0},
		{PID: 3, Name: "charlie", CPU: 15.0, Mem: 3.0},
		{PID: 2, Name: "bravo", CPU: 10.0, Mem: 2.0},
	}

	tests := []struct {
		name  string
		field string
		first int // expected PID of first element
	}{
		{"sort by cpu descending", "cpu", 3},
		{"sort by mem descending", "mem", 3},
		{"sort by pid ascending", "pid", 1},
		{"sort by name ascending", "name", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy to avoid mutation between tests.
			p := make([]Info, len(procs))
			copy(p, procs)
			Sort(p, tt.field)
			if p[0].PID != tt.first {
				t.Errorf("Sort(%q) first PID = %d, want %d", tt.field, p[0].PID, tt.first)
			}
		})
	}
}

func TestList(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(procs) == 0 {
		t.Error("List() returned no processes")
	}

	// Verify at least one process has a non-empty name.
	found := false
	for _, p := range procs {
		if p.Name != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("List() returned no processes with names")
	}
}

func TestTop(t *testing.T) {
	procs, err := Top(5)
	if err != nil {
		t.Fatalf("Top(5) error: %v", err)
	}
	if len(procs) > 5 {
		t.Errorf("Top(5) returned %d processes, want <= 5", len(procs))
	}
	if len(procs) == 0 {
		t.Error("Top(5) returned no processes")
	}
}

func TestFind(t *testing.T) {
	// "launchd" should always be running on macOS.
	procs, err := Find("launchd")
	if err != nil {
		t.Fatalf("Find(launchd) error: %v", err)
	}
	if len(procs) == 0 {
		t.Error("Find(launchd) returned no results")
	}
	for _, p := range procs {
		if !strings.Contains(strings.ToLower(p.Name), "launchd") &&
			!strings.Contains(strings.ToLower(p.Command), "launchd") {
			t.Errorf("Find(launchd) returned unexpected process: %q", p.Name)
		}
	}
}

func TestFindNoResults(t *testing.T) {
	procs, err := Find("zzz_nonexistent_process_zzz")
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if len(procs) != 0 {
		t.Errorf("Find(nonexistent) returned %d results, want 0", len(procs))
	}
}
