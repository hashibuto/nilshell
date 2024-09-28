package ns

import "github.com/hashibuto/nilshell/pkg/termutils"

// CalculateColumnWidth returns the column width and number of columns per row
func CalculateColumnWidth(allText []string, screenWidth int, minColumns int, gutterWidth int) (int, int) {
	maxWidth := 0
	for _, text := range allText {
		measuredWidth := termutils.Measure(text)
		if measuredWidth > maxWidth {
			maxWidth = measuredWidth
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
