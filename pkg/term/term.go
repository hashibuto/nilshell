package term

import (
	"regexp"
	"strings"
)

var escapeFinder = regexp.MustCompile("\x1b\\[[^a-zA-Z]+[a-zA-Z]")

// Measure returns the number of horizonal space the provided text accounts for.  This will filter out escape characters, and treat multi-byte
// unicode characters as single space tenants.
func Measure(text string) int {
	isTermEsc := false
	length := 0
	for _, r := range text {
		if r == 0x1B {
			isTermEsc = true
			continue
		}

		if isTermEsc {
			if (r >= 65 && r <= 89) || (r >= 97 && r <= 122) {
				isTermEsc = false
			}
			continue
		}

		length++
	}

	return length
}

// Crop ensures that the supplied text is cropped to the target length.  If it exceeds the target length, the last (max) 3 characters
// are replaced with ellipsis.  Returned are the cropped string, and the final visible length after cropping.
func Crop(text string, length int) (string, int) {
	output := []rune{}
	ellipsisArray := []int{}

	isTermEsc := false
	visLength := 0
	for oIdx, r := range text {
		if r == 0x1B {
			output = append(output, r)
			isTermEsc = true
			continue
		}

		if isTermEsc {
			output = append(output, r)
			if (r >= 65 && r <= 89) || (r >= 97 && r <= 122) {
				isTermEsc = false
			}
			continue
		}

		if visLength < length {
			output = append(output, r)
			if visLength >= length-3 && visLength < length {
				ellipsisArray = append(ellipsisArray, oIdx)
			}
		}
		visLength++
	}

	if visLength > length {
		for _, idx := range ellipsisArray {
			output[idx] = '.'
		}
	}

	finalLength := visLength
	if visLength > length {
		finalLength = length
	}
	return string(output), finalLength
}

// PadRight pads text to the right, cropping anything over the supplied maxLength, preserving up to n gutter chars on the right.
func PadRight(text string, maxLength int, gutter int) string {
	croppedText, croppedLen := Crop(text, maxLength-gutter)

	return croppedText + strings.Repeat(" ", maxLength-croppedLen)
}

// StripTerminalEscapeSequences removes all terminal escape sequences from the provided string, and returns the remaining string
func StripTerminalEscapeSequences(data []byte) []byte {
	return escapeFinder.ReplaceAll(data, []byte{})
}
