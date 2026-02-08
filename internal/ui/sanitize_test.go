package ui

import (
	"testing"
)

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "color codes (SGR) should be kept",
			input:    "\x1b[31mred\x1b[0m",
			expected: "\x1b[31mred\x1b[0m",
		},
		{
			name:     "cursor move (H) should be removed",
			input:    "\x1b[Hheader",
			expected: "header",
		},
		{
			name:     "clear screen (2J) should be removed",
			input:    "\x1b[2Jclean",
			expected: "clean",
		},
		{
			name:     "complex mix",
			input:    "\x1b[2J\x1b[32mgreen\x1b[0m\x1b[10;20Hmove",
			expected: "\x1b[32mgreen\x1b[0mmove",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeOutput(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeOutput(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
