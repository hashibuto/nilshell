package ns

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
)

type ProcessingCode int8

const (
	CodeContinue ProcessingCode = iota
	CodeComplete
	CodeCancel
	CodeTerminate
)

type LineReader struct {
	prompt       []rune
	completer    Completer
	bufferOffset int
	resizeChan   chan os.Signal
	buffer       []rune
	lock         *sync.Mutex
	winWidth     int
	winHeight    int
	cursorRow    int
}

// NewLineReader creates a new LineReader object
func NewLineReader(completer Completer, resizeChan chan os.Signal) *LineReader {
	lr := &LineReader{
		completer:  completer,
		resizeChan: resizeChan,
		buffer:     []rune{},
		lock:       &sync.Mutex{},
	}

	go lr.resizeWatcher()
	lr.resizeWindow(false)

	return lr
}

// Read will read a single command from the command line and can be interrupted by pressing <enter>, <ctrl+c>, or <ctrl+d>.
// Read responds to changes in the terminal window size.
func (lr *LineReader) Read(prompt string, history *History, completionLimit int) (string, bool, error) {
	cursorRow, _ := getCursorPos()
	setCursorPos(cursorRow, 1)
	lr.cursorRow = cursorRow
	lr.bufferOffset = 0
	lr.buffer = []rune{}
	fmt.Printf("%s", prompt)

	lr.prompt = []rune(prompt)
	iBuf := make([]byte, 20)
	for {
		n, err := os.Stdin.Read(iBuf)
		if err != nil {
			return "", false, err
		}
		iString := string(iBuf[:n])
		code := lr.processInput(iString, history, completionLimit)
		switch code {
		case CodeComplete:
			return string(lr.buffer), false, nil
		case CodeCancel:
			return "", false, nil
		case CodeTerminate:
			lr.resizeChan <- syscall.SIGTERM
			return "", true, nil
		}
	}
}

// resizeWatcher waits for input on the resize channel and resizes accordingly.  when the channel receives a SIGTERM the thread exists.
// the SIGTERM originates internally, not from outside the process.
func (lr *LineReader) resizeWatcher() {
	for {
		sig := <-lr.resizeChan
		if sig == syscall.SIGTERM {
			return
		}

		lr.resizeWindow(true)
	}
}

// processInput executes one iteration of input processing which would occur in the interactive read loop
func (lr *LineReader) processInput(input string, history *History, completionLimit int) ProcessingCode {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	switch input {
	case KEY_UP_ARROW:
		cmd := history.Older()
		if len(cmd) > 0 {
			lr.setText([]rune(cmd))
		}
	case KEY_DOWN_ARROW:
		cmd := history.Newer()
		if len(cmd) > 0 {
			lr.setText([]rune(cmd))
		}
	case KEY_LEFT_ARROW:
		lr.bufferOffset--
		if lr.bufferOffset < 0 {
			lr.bufferOffset = 0
		}
		lr.setCursorPos()
	case KEY_RIGHT_ARROW:
		lr.bufferOffset++
		if lr.bufferOffset > len(lr.buffer) {
			lr.bufferOffset = len(lr.buffer)
		}
		lr.setCursorPos()
	case KEY_HOME:
		lr.bufferOffset = 0
		lr.setCursorPos()
	case KEY_END:
		lr.bufferOffset = len(lr.buffer)
		lr.setCursorPos()
	case KEY_ENTER:
		fmt.Println()
		return CodeComplete
	case KEY_CTRL_C:
		fmt.Println()
		return CodeCancel
	case KEY_CTRL_D:
		fmt.Println()
		return CodeTerminate
	case KEY_DEL:
		lr.deleteAtCurrentPos()
	case KEY_TAB:
		autoComplete := lr.completer(string(lr.buffer[:lr.bufferOffset]), string(lr.buffer[lr.bufferOffset:]), string(lr.buffer))
		if autoComplete != nil {
			if len(autoComplete) == 1 {
				ac := autoComplete[0]
				lr.completeText([]rune(ac.Name))
			} else if len(autoComplete) > 1 && len(autoComplete) <= completionLimit {
				lr.displayAutocomplete(autoComplete)
			} else if len(autoComplete) > completionLimit {
				lr.displayTooManyAutocomplete(autoComplete)
			}
		}
	case KEY_BACKSPACE:
		if lr.bufferOffset > 0 {
			lr.bufferOffset--
			lr.deleteAtCurrentPos()
		}
	default:
		lr.insertText([]rune(input))
	}

	return CodeContinue
}

// displayTooManyAutocomplete displays the too many autocomplete suggestions message
func (lr *LineReader) displayTooManyAutocomplete(autoComplete []*AutoComplete) {
	y, _ := getCursorPos()
	fmt.Printf("\r\n")
	y++
	fmt.Printf("%d suggestions, too many to display...", len(autoComplete))
	y++
	fmt.Println()
	if y > lr.winHeight {
		y = lr.winHeight
	}
	lr.cursorRow = y
	lr.renderComplete()
}

// displayAutocomplete displays the autocomplete suggestions
func (lr *LineReader) displayAutocomplete(autoComplete []*AutoComplete) {
	y, _ := getCursorPos()
	fmt.Printf("\r\n")
	y++
	total := 0
	for _, ac := range autoComplete {
		text := ac.Name
		if len(text) > 12 {
			text = text[:12] + "..."
		}
		total += 18
		if total > lr.winWidth {
			y++
			fmt.Printf("\r\n")
			total = 18
		}
		fmt.Printf("%-20s", text)
	}
	y++
	fmt.Println()
	if y > lr.winHeight {
		y = lr.winHeight
	}
	lr.cursorRow = y
	lr.renderComplete()
}

// setText sets the current input text
func (lr *LineReader) setText(input []rune) {
	lr.buffer = input
	lr.bufferOffset = len(lr.buffer)
	lr.renderComplete()
}

// insertText inserts text at the current cursor position
func (lr *LineReader) insertText(input []rune) {
	runeBuffer := []rune{}
	runeBuffer = append(runeBuffer, lr.buffer[:lr.bufferOffset]...)
	runeBuffer = append(runeBuffer, input...)
	runeBuffer = append(runeBuffer, lr.buffer[lr.bufferOffset:]...)
	lr.buffer = runeBuffer

	newBufferOffset := lr.bufferOffset + len(input)

	lr.renderFromCursor()
	lr.bufferOffset = newBufferOffset
	lr.setCursorPos()
}

// completeText performs an autocomplete operation
func (lr *LineReader) completeText(input []rune) {
	// hunt back to the previous either space, or beginning of the text from the current cursor position
	inputStr := string(input)

	for i := lr.bufferOffset - 1; i >= 0; i-- {
		if lr.buffer[i] == ' ' || i == 0 {
			j := i
			if lr.buffer[i] == ' ' {
				j++
			}
			strPrefix := string(lr.buffer[j:lr.bufferOffset])

			if !strings.HasPrefix(inputStr, strPrefix) {
				return
			}

			runePrefix := []rune(strPrefix)
			lr.insertText(input[len(runePrefix):])
		}
	}
}

// deleteAtCurrentPos deletes a single character at the current cursor position
func (lr *LineReader) deleteAtCurrentPos() {
	if lr.bufferOffset < len(lr.buffer) {
		runeBuffer := []rune{}
		runeBuffer = append(runeBuffer, lr.buffer[:lr.bufferOffset]...)
		runeBuffer = append(runeBuffer, lr.buffer[lr.bufferOffset+1:]...)
		lr.buffer = runeBuffer

		lr.renderFromCursor()
		lr.setCursorPos()
	}
}

// renderFromCursor renders the input line starting from the current cursor position
func (lr *LineReader) renderFromCursor() {
	lr.setCursorPos()
	fmt.Printf("%s", string(lr.buffer[lr.bufferOffset:]))
	lr.renderEraseForward(true)
}

// renderEraseForward renders the erase forward pattern so that input does not "drag" when deletion occurs
func (lr *LineReader) renderEraseForward(justOne bool) {
	remainder := lr.winWidth - ((len(lr.buffer) + len(lr.prompt)) % lr.winWidth)
	if remainder > 0 {
		if justOne {
			remainder = 1
		}
		remBuf := make([]byte, remainder)
		for i := 0; i < remainder; i++ {
			remBuf[i] = ' '
		}
		os.Stdout.Write(remBuf)
	}
}

// renderComplete renders the complete input text regardless of the cursor position
func (lr *LineReader) renderComplete() {
	setCursorPos(lr.cursorRow, 1)
	fmt.Printf("%s", string(lr.prompt))
	fmt.Printf("%s", string(lr.buffer))
	lr.renderEraseForward(false)
	lr.setCursorPos()
}

// resizeWindow re-renders according to the window size
func (lr *LineReader) resizeWindow(render bool) {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	lr.winHeight, lr.winWidth = getWindowDimensions()
	if render {
		lr.renderComplete()
	}
}

// setCursorPos sets the current cursor position based on the linear offset in the command input
func (lr *LineReader) setCursorPos() {
	// Determine the linear cursor position, including the prompt
	linearCursorPos := len(lr.prompt) + lr.bufferOffset
	curCursorRow := lr.cursorRow + int(linearCursorPos/lr.winWidth)
	curCursorCol := (linearCursorPos % lr.winWidth) + 1
	setCursorPos(curCursorRow, curCursorCol)
}
