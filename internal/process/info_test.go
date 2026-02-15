package process

import (
	"os"
	"testing"
)

func TestGetInfoCurrentProcess(t *testing.T) {
	pid := os.Getpid()
	info, err := GetInfo(pid)
	if err != nil {
		t.Fatalf("GetInfo(%d) error: %v", pid, err)
	}

	if info.PID != pid {
		t.Errorf("PID = %d, want %d", info.PID, pid)
	}

	if info.Name == "" {
		t.Error("Name is empty")
	}

	if info.User == "" {
		t.Error("User is empty")
	}
}

func TestGetInfoNonExistentProcess(t *testing.T) {
	// PID 999999 is very unlikely to exist.
	_, err := GetInfo(999999)
	if err == nil {
		t.Error("GetInfo(999999) should return error for non-existent process")
	}
}

func TestExtractPort(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		{
			name: "TCP listening port",
			line: "node    12345 user   10u  IPv4 0x1234  0t0  TCP *:8080 (LISTEN)",
			want: 8080,
		},
		{
			name: "TCP localhost port",
			line: "node    12345 user   10u  IPv4 0x1234  0t0  TCP localhost:3000 (LISTEN)",
			want: 3000,
		},
		{
			name: "no port",
			line: "node    12345 user    0r  REG   1,18 12345 /usr/bin/node",
			want: 0,
		},
		{
			name: "empty line",
			line: "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPort(tt.line)
			if got != tt.want {
				t.Errorf("extractPort() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseLsofConnections(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
		check func(t *testing.T, conns []Connection)
	}{
		{
			name:  "empty output",
			input: "",
			want:  0,
		},
		{
			name:  "header only",
			input: "COMMAND     PID   USER   FD   TYPE DEVICE SIZE/OFF NODE NAME",
			want:  0,
		},
		{
			name: "single TCP LISTEN",
			input: `COMMAND     PID   USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node      12345 user   10u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)`,
			want: 1,
			check: func(t *testing.T, conns []Connection) {
				c := conns[0]
				if c.Protocol != "TCP" {
					t.Errorf("Protocol = %q, want TCP", c.Protocol)
				}
				if c.LocalAddr != "*:8080" {
					t.Errorf("LocalAddr = %q, want *:8080", c.LocalAddr)
				}
				if c.RemoteAddr != "" {
					t.Errorf("RemoteAddr = %q, want empty", c.RemoteAddr)
				}
				if c.State != "LISTEN" {
					t.Errorf("State = %q, want LISTEN", c.State)
				}
			},
		},
		{
			name: "TCP ESTABLISHED with remote",
			input: `COMMAND     PID   USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node      12345 user   11u  IPv4 0x5678      0t0  TCP 127.0.0.1:3000->10.0.0.1:54321 (ESTABLISHED)`,
			want: 1,
			check: func(t *testing.T, conns []Connection) {
				c := conns[0]
				if c.Protocol != "TCP" {
					t.Errorf("Protocol = %q, want TCP", c.Protocol)
				}
				if c.LocalAddr != "127.0.0.1:3000" {
					t.Errorf("LocalAddr = %q, want 127.0.0.1:3000", c.LocalAddr)
				}
				if c.RemoteAddr != "10.0.0.1:54321" {
					t.Errorf("RemoteAddr = %q, want 10.0.0.1:54321", c.RemoteAddr)
				}
				if c.State != "ESTABLISHED" {
					t.Errorf("State = %q, want ESTABLISHED", c.State)
				}
			},
		},
		{
			name: "multiple connections",
			input: `COMMAND     PID   USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node      12345 user   10u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)
node      12345 user   11u  IPv4 0x5678      0t0  TCP 127.0.0.1:3000->10.0.0.1:54321 (ESTABLISHED)
node      12345 user   12u  IPv4 0x9abc      0t0  UDP *:5353`,
			want: 3,
		},
		{
			name: "non-network lines are skipped",
			input: `COMMAND     PID   USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node      12345 user    0r   REG    1,18    12345  REG /usr/bin/node
node      12345 user   10u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)`,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLsofConnections(tt.input)
			if len(got) != tt.want {
				t.Errorf("ParseLsofConnections() returned %d connections, want %d", len(got), tt.want)
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseEnvVars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name:  "command only no env vars",
			input: "/usr/local/bin/node server.js",
			want:  nil,
		},
		{
			name:  "command with env vars",
			input: "/usr/local/bin/node server.js HOME=/Users/test PATH=/usr/bin:/usr/local/bin NODE_ENV=production",
			want: map[string]string{
				"HOME":     "/Users/test",
				"PATH":     "/usr/bin:/usr/local/bin",
				"NODE_ENV": "production",
			},
		},
		{
			name:  "env var with empty value",
			input: "/bin/sh DEBUG=",
			want: map[string]string{
				"DEBUG": "",
			},
		},
		{
			name:  "lowercase keys are not env vars",
			input: "/usr/bin/app --config=file.json port=8080",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEnvVars(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseEnvVars() = %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseEnvVars() returned %d vars, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseEnvVars()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestIsEnvVarKey(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"HOME", true},
		{"PATH", true},
		{"NODE_ENV", true},
		{"A1", true},
		{"__", true},
		{"A", false},        // too short
		{"", false},         // empty
		{"home", false},     // lowercase
		{"Home", false},     // mixed case
		{"MY-VAR", false},   // contains hyphen
		{"MY.VAR", false},   // contains dot
		{"MY VAR", false},   // contains space
		{"123", true},       // digits only, 3+ chars
		{"A_B_C_D", true},   // underscores
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isEnvVarKey(tt.input)
			if got != tt.want {
				t.Errorf("isEnvVarKey(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
