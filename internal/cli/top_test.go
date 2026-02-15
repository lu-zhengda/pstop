package cli

import "testing"

func TestSpark(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"zero", 0, "▁"},
		{"negative", -5.0, "▁"},
		{"low", 10.0, "▁"},
		{"quarter", 25.0, "▂"},
		{"half", 50.0, "▄"},
		{"high", 85.0, "▆"},
		{"full", 100.0, "█"},
		{"over", 150.0, "█"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spark(tt.pct)
			if got != tt.want {
				t.Errorf("spark(%.1f) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestSparkMonotonic(t *testing.T) {
	// Verify that spark output is monotonically non-decreasing.
	prev := spark(0)
	for pct := 1.0; pct <= 100.0; pct++ {
		curr := spark(pct)
		if curr < prev {
			t.Errorf("spark(%.0f) = %q < spark(%.0f) = %q, not monotonic", pct, curr, pct-1, prev)
		}
		prev = curr
	}
}
