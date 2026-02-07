package monitor

import (
	"testing"
)

func TestFormatRate(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0KB/s"},
		{500, "500KB/s"},
		{1023, "1023KB/s"},
		{1024, "1.0MB/s"},
		{2048, "2.0MB/s"},
		{1536, "1.5MB/s"},
	}

	for _, tt := range tests {
		result := FormatRate(tt.input)
		if result != tt.expected {
			t.Errorf("FormatRate(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTrimHistory(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		max      int
		expected int
	}{
		{"empty", []float64{}, 5, 0},
		{"under limit", []float64{1, 2, 3}, 5, 3},
		{"at limit", []float64{1, 2, 3, 4, 5}, 5, 5},
		{"over limit", []float64{1, 2, 3, 4, 5, 6}, 5, 5},
		{"much over limit", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimHistory(tt.input, tt.max)
			if len(result) != tt.expected {
				t.Errorf("len(trimHistory) = %d, want %d", len(result), tt.expected)
			}
			// check if we kept the *last* elements
			if len(result) > 0 {
				lastInput := tt.input[len(tt.input)-1]
				lastResult := result[len(result)-1]
				if lastInput != lastResult {
					t.Errorf("last element = %v, want %v", lastResult, lastInput)
				}
			}
		})
	}
}

func TestUpdateHistory(t *testing.T) {
	history := MetricHistory{}
	sample := MetricsSample{
		Load: 1.0, OkLoad: true,
		CPU: 50.0, OkCPU: true,
		Mem: 25.0, OkMem: true,
		NetKB: 100.0, OkNet: true,
	}

	// First update
	history = UpdateHistory(history, sample)
	if len(history.Load) != 1 || history.Load[0] != 1.0 {
		t.Errorf("UpdateHistory failed on first update")
	}

	// Add enough items to trigger trim (HistoryLength is 30)
	for i := 0; i < 40; i++ {
		history = UpdateHistory(history, sample)
	}

	if len(history.Load) != HistoryLength {
		t.Errorf("UpdateHistory should trim to %d, got %d", HistoryLength, len(history.Load))
	}
}
