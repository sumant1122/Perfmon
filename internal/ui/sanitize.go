package ui

import (
	"regexp"
)

// sanitizeOutput removes ANSI cursor movement and clear screen codes
// while preserving SGR (color/style) codes.
// Specifically, it removes CSI sequences ending in anything other than 'm'.
func sanitizeOutput(input string) string {
	// CSI sequence: ESC [ ... FinalByte
	// We want to remove all CSI sequences where FinalByte is NOT 'm'.
	// Common CSI finals: A-H (cursor move), J (clear), K (clear line), etc.
	// Range @-~ (0x40-0x7E).
	// 'm' is 0x6D.
	// Regex: \x1b\[[\d;?]*[@-ln-~]
	// \x1b matches ESC.
	// \[ matches [.
	// [\d;?]* matches parameters (digits, semicolon, question mark for private modes).
	// [@-ln-~] matches valid final bytes EXCEPT 'm'.

	// This regex is slightly simplified but covers most modern terminal usage
	re := regexp.MustCompile(`\x1b\[[\d;?]*[@-ln-~]`)
	return re.ReplaceAllString(input, "")
}
