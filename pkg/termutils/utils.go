package termutils

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	escapeFinder    = regexp.MustCompile("\x1b\\[[^a-zA-Z]+[a-zA-Z]")
	positionMatcher = regexp.MustCompile("\x1b\\[(\\d*);(\\d*)R")
)

const (
	TERM_CLEAR               = "\x1B[2J"
	TERM_CLEAR_END_OF_SCREEN = "\x1B[0J"
	TERM_CLEAR_END_OF_LINE   = "\x1B[0K"
	STYLE_RESET              = "\x1b[0m"
	STYLE_BOLD               = "\x1b[1m"
)

var (
	ErrNoMatch = errors.New("no match")
)

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

// ClearTerminal clears the terminal without repositioning the cursor
func ClearTerminal() {
	fmt.Printf("%s", TERM_CLEAR)
}

// GetWindowSize returns the size of the window (row, col)
func GetWindowSize() (int, int) {
	winsize, _ := unix.IoctlGetWinsize(int(os.Stdout.Fd()), syscall.TIOCGWINSZ)
	return int(winsize.Row), int(winsize.Col)
}

// SetCursorPos sets the current cursor position.  row and col start from 1
func SetCursorPos(row int, col int) {
	os.Stdout.WriteString(fmt.Sprintf("\x1b[%d;%dH", row, col))
}

func CreateFgColor(red int, green int, blue int) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", red, green, blue)
}

func SetFgColor(red int, green int, blue int) {
	os.Stdout.WriteString(CreateFgColor(red, green, blue))
}

func SetBgColor(red int, green int, blue int) {
	os.Stdout.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", red, green, blue))
}

func ResetStyle() {
	os.Stdout.WriteString(STYLE_RESET)
}

func HideCursor() {
	os.Stdout.WriteString("\x1b[?25l")
}

func ShowCursor() {
	os.Stdout.WriteString("\x1b[?25h")
}

func RequestCursorPos() {
	os.Stdout.WriteString("\x1b[6n")
}

// GetCursorPosition returns row, column or an error
func GetCursorPosition(value string) (int, int, error) {
	matches := positionMatcher.FindStringSubmatch(value)
	if len(matches) == 0 {
		return 0, 0, ErrNoMatch
	}

	row, _ := strconv.Atoi(matches[1])
	col, _ := strconv.Atoi(matches[2])

	return row, col, nil
}

func ClearTerminalFromCursor() {
	os.Stdout.WriteString(TERM_CLEAR_END_OF_SCREEN)
}

func ClearLineFromCursor() {
	os.Stdout.WriteString(TERM_CLEAR_END_OF_LINE)
}
