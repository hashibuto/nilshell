package ns

import (
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
	b := make([]byte, 10)
	// Request the cursor position
	fmt.Printf("\033[6n")
	bLen, _ := os.Stdin.Read(b)

	section := string(b[2 : bLen-1])
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
	fmt.Print("\033[H\033[2J")
}
