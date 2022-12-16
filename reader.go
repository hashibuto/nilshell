package ns

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/term"
)

type ProcessingCode int8

var EscapeFinder = regexp.MustCompile("\033\\[[^m]+m")

const (
	CodeContinue ProcessingCode = iota
	CodeComplete
	CodeCancel
	CodeTerminate
)

type LineReader struct {
	lastSearchText  []rune
	nilShell        *NilShell
	isReverseSearch bool
	prompt          string
	promptLength    int
	completer       Completer
	bufferOffset    int
	resizeChan      chan os.Signal
	buffer          []rune
	lock            *sync.Mutex
	winWidth        int
	winHeight       int
	cursorRow       int
}

var reverseSearchPrompt = "(reverse search: `"

// NewLineReader creates a new LineReader object
func NewLineReader(completer Completer, resizeChan chan os.Signal, nilShell *NilShell) *LineReader {
	lr := &LineReader{
		completer:  completer,
		resizeChan: resizeChan,
		buffer:     []rune{},
		lock:       &sync.Mutex{},
		nilShell:   nilShell,
	}

	go lr.resizeWatcher()
	lr.resizeWindow(false)

	return lr
}

// Read will read a single command from the command line and can be interrupted by pressing <enter>, <ctrl+c>, or <ctrl+d>.
// Read responds to changes in the terminal window size.
func (lr *LineReader) Read() (string, bool, error) {
	fd := int(os.Stdin.Fd())
	preState, err := term.MakeRaw(fd)
	if err != nil {
		return "", false, err
	}
	lr.nilShell.preState = preState

	// Try our best not to leave the terminal in raw mode
	defer func() {
		err := recover()
		term.Restore(fd, lr.nilShell.preState)
		if err != nil {
			fmt.Printf("Caught panic before exiting\n%v\n", err)
			if lr.nilShell.Debug {
				fmt.Println(string(debug.Stack()))
			}
		}
		if err != nil {
			os.Exit(1)
		}
	}()

	cursorRow, _ := getCursorPos()
	setCursorPos(cursorRow, 1)
	lr.cursorRow = cursorRow
	lr.bufferOffset = 0
	lr.buffer = []rune{}
	fmt.Printf("%s", lr.nilShell.Prompt)

	lr.promptLength = len([]rune(EscapeFinder.ReplaceAllString(lr.nilShell.Prompt, "")))

	iBuf := make([]byte, 20)
	for {
		n, err := os.Stdin.Read(iBuf)
		if err != nil {
			return "", false, err
		}
		iString := string(iBuf[:n])
		code := lr.processInput(iString, lr.nilShell)
		switch code {
		case CodeComplete:
			lr.isReverseSearch = false
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
func (lr *LineReader) processInput(input string, n *NilShell) ProcessingCode {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	switch input {
	case KEY_CTRL_R:
		lr.isReverseSearch = true
		lr.renderComplete()
	case KEY_CTRL_T:
		lr.openInEditor()
	case KEY_UP_ARROW:
		if n.History.Any() {
			cmd := n.History.Older()
			lr.setText([]rune(cmd))
		}
	case KEY_DOWN_ARROW:
		if n.History.Any() {
			cmd := n.History.Newer()
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
		if lr.isReverseSearch {
			buffer := make([]rune, len(lr.lastSearchText))
			copy(buffer, lr.lastSearchText)
			lr.buffer = buffer
		}
		fmt.Printf("\r\n")
		return CodeComplete
	case KEY_CTRL_C:
		fmt.Printf("\r\n")
		return CodeCancel
	case KEY_CTRL_D:
		fmt.Printf("\r\n")
		return CodeTerminate
	case KEY_CTRL_L:
		clear()
		// Tell LineReader that we are back at the first row,
		// since it keeps track of the cursor itself.
		lr.cursorRow = 1
		lr.renderComplete()
	case KEY_DEL:
		lr.deleteAtCurrentPos()
	case KEY_TAB:
		if lr.isReverseSearch {
			buffer := make([]rune, len(lr.lastSearchText))
			copy(buffer, lr.lastSearchText)
			lr.buffer = buffer
			lr.isReverseSearch = false
			lr.bufferOffset = len(lr.buffer)
			lr.setCursorPos()
			lr.renderComplete()
		} else {
			autoComplete := lr.completer(string(lr.buffer[:lr.bufferOffset]), string(lr.buffer[lr.bufferOffset:]), string(lr.buffer))
			if autoComplete != nil {
				if len(autoComplete) == 1 {
					ac := autoComplete[0]
					lr.completeText([]rune(ac.Name))
				} else if len(autoComplete) > 1 && len(autoComplete) <= n.AutoCompleteLimit {
					lr.displayAutocomplete(autoComplete, n)
				} else if len(autoComplete) > n.AutoCompleteLimit {
					lr.displayTooManyAutocomplete(autoComplete, n)
				}
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
func (lr *LineReader) displayTooManyAutocomplete(autoComplete []*AutoComplete, ns *NilShell) {
	y, _ := getCursorPos()
	fmt.Printf("\r\n")
	y++
	fmt.Printf("%s%d suggestions, too many to display...%s", ns.AutoCompleteTooMuchStyle, len(autoComplete), CODE_RESET)
	y++
	fmt.Printf("\r\n")
	if y > lr.winHeight {
		y = lr.winHeight
	}
	lr.cursorRow = y
	lr.renderComplete()
}

// displayAutocomplete displays the autocomplete suggestions
func (lr *LineReader) displayAutocomplete(autoComplete []*AutoComplete, ns *NilShell) {
	y, _ := getCursorPos()
	fmt.Printf("\r\n%s", ns.AutoCompleteSuggestStyle)
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
	fmt.Printf("%s\n\r", CODE_RESET)
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
	if lr.isReverseSearch {
		lr.renderComplete()
	} else {
		lr.setCursorPos()
		fmt.Printf("%s", string(lr.buffer[lr.bufferOffset:]))
		lr.renderEraseForward(true)
	}
}

// renderEraseForward renders the erase forward pattern so that input does not "drag" when deletion occurs
func (lr *LineReader) renderEraseForward(justOne bool) {
	var totalOffset int
	if lr.isReverseSearch {
		totalOffset = len(reverseSearchPrompt) + 4 + len(lr.buffer) + len(lr.lastSearchText)
	} else {
		totalOffset = lr.promptLength + len(lr.buffer)
	}
	remainder := lr.winWidth - (totalOffset % lr.winWidth)
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
	if lr.isReverseSearch {
		lr.lastSearchText = []rune(lr.nilShell.History.FindMostRecentMatch(string(lr.buffer)))
		setCursorPos(lr.cursorRow, 1)
		fmt.Printf("%s%s`): %s", reverseSearchPrompt, string(lr.buffer), string(lr.lastSearchText))
		lr.renderEraseForward(false)
		lr.setCursorPos()
	} else {
		setCursorPos(lr.cursorRow, 1)
		fmt.Printf("%s", lr.nilShell.Prompt)
		fmt.Printf("%s", string(lr.buffer))
		lr.renderEraseForward(false)
		lr.setCursorPos()
	}
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
	var promptOffset int
	if lr.isReverseSearch {
		promptOffset = len(reverseSearchPrompt)
	} else {
		promptOffset = lr.promptLength
	}
	linearCursorPos := promptOffset + lr.bufferOffset
	curCursorRow := lr.cursorRow + int(linearCursorPos/lr.winWidth)
	curCursorCol := (linearCursorPos % lr.winWidth) + 1
	setCursorPos(curCursorRow, curCursorCol)
}

// Reset the LineReader buffer and return its contents.
func (lr *LineReader) reset() []rune {
	ret := make([]rune, len(lr.buffer))
	copy(ret, lr.buffer)

	lr.buffer = lr.buffer[:0]
	lr.bufferOffset = 0

	return ret
}

// openInEditor takes the contents of the LineReader buffer and stores it in a temp file, which is
// then opened in the user's $EDITOR. After the editor is closed the contents of the file are put
// back into the LineReader buffer.
func (lr *LineReader) openInEditor() {
	// Create a temp file to store the buffer in.
	tmpFile, err := os.CreateTemp("", "ns-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer os.Remove(tmpFile.Name())

	currentBuffer := lr.reset()
	tmpFile.WriteString(string(currentBuffer))
	tmpFile.Close()

	// Get the user's editor, and open the file in it.
	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		editorName = "vi"
	}

	editor := exec.Command(editorName, tmpFile.Name())
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr
	editor.Run()

	// Open that temp file again, and put its contents back into the buffer.
	tmpFile, err = os.Open(tmpFile.Name())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer tmpFile.Close()

	contents, err := io.ReadAll(tmpFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	contentsAsRunes := convertAndTrim(contents)

	lr.buffer = contentsAsRunes
	lr.bufferOffset = len(contentsAsRunes)
	lr.renderComplete()
}

// convertAndTrim returns a rune slice for a given byte slice, minus a final line feed character.
func convertAndTrim(b []byte) []rune {
	s := string(b)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return []rune(s)
}
