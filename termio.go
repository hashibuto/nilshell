package ns

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var CODE_RESET = "\x1b[0m"

func requestCursorPos() {
	fmt.Printf("\x1b[6n")
}

// getCursorPos returns the current cursor position (row, col).  row and col start from 1
func getCursorPos() (int, int) {
	r := bufio.NewReader(os.Stdin)

	// Request the cursor position
	requestCursorPos()

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
	os.Stdout.WriteString(fmt.Sprintf("\x1b[%d;%dH", row, col))
}

// clear clears the terminal
func clear() {
	os.Stdout.WriteString("\x1b[H\x1b[2J")
}

// showCursor makes the cursor appear
func showCursor() {
	os.Stdout.WriteString("\x1b[?25h")
}
