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
