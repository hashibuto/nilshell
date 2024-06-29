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

// CalculateColumnWidth returns the column width and number of columns per row
func CalculateColumnWidth(allText []string, screenWidth int, minColumns int, gutterWidth int) (int, int) {
	maxWidth := 0
	for _, text := range allText {
		if len(text) > maxWidth {
			maxWidth = len(text)
		}
	}

	maxWidth += gutterWidth

	maxColumnWidth := int(screenWidth / minColumns)
	if maxColumnWidth < 1 {
		maxColumnWidth = 1
	}
	if maxWidth > maxColumnWidth {
		maxWidth = maxColumnWidth
	}

	return maxWidth, int(screenWidth) / maxWidth
}
