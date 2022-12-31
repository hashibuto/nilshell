package ns

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var CODE_RESET = "\033[0m"

// getCursorPos returns the current cursor position (row, col).  row and col start from 1
func getCursorPos() (int, int) {
	r := bufio.NewReader(os.Stdin)

	// Request the cursor position
	fmt.Printf("\033[6n")

	b, err := r.ReadString('R')
	if err != nil {
		panic(fmt.Sprintf("getCursorPos: %s", err.Error()))
	}

	// len(b) should be _at least_ 6, since the shortest possible valid response
	// would be `\033[1;1R`, and `\033` (the escape char) counts as 1.

	section := b[2 : len(b)-1]
	parts := strings.Split(section, ";")
	row, _ := strconv.Atoi(parts[0])
	col, _ := strconv.Atoi(parts[1])

	return row, col
}

// setCursorPos sets the current cursor position.  row and col start from 1
func setCursorPos(row int, col int) {
	os.Stdout.WriteString(fmt.Sprintf("\x1B[%d;%dH", row, col))
}

// getWindowDimensions returns the size of the window (row, col)
func getWindowDimensions() (int, int) {
	winsize, _ := unix.IoctlGetWinsize(int(os.Stdout.Fd()), syscall.TIOCGWINSZ)
	return int(winsize.Row), int(winsize.Col)
}

// clear clears the terminal
func clear() {
	os.Stdout.WriteString("\033[H\033[2J")
}

// hideCursor makes the cursor disappear
func hideCursor() {
	os.Stdout.WriteString("\x1B[?25l")
}

// showCursor makes the cursor appear
func showCursor() {
	os.Stdout.WriteString("\x1B[?25h")
}
