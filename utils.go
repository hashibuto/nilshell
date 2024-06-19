package ns

import "strings"

// PadRight will pad text with blank space on the right, leaving up to gutter characters unused.
// If text is longer than width - gutter, it will be terminated with an ellipsis, up to the gutter line.
func PadRight(text string, width int, gutter int) string {
	if len(text) > width-gutter {
		text = text[:width-gutter-3] + "..."
	}

	return text + strings.Repeat(" ", width-len(text))
}
