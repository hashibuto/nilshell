package ns

import (
	"fmt"
	"os"
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

func NewLineReader(
	completer Completer,
	resizeChan chan os.Signal,
) *LineReader {
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

func (lr *LineReader) Read(prompt string) (string, bool, error) {
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
		code := lr.processInput(iString)
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

func (lr *LineReader) resizeWatcher() {
	for {
		sig := <-lr.resizeChan
		if sig == syscall.SIGTERM {
			return
		}

		lr.resizeWindow(true)
	}
}

func (lr *LineReader) processInput(input string) ProcessingCode {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	switch input {
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

func (lr *LineReader) renderFromCursor() {
	lr.setCursorPos()
	fmt.Printf("%s", string(lr.buffer[lr.bufferOffset:]))
	lr.renderEraseForward(true)
}

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

func (lr *LineReader) renderComplete() {
	setCursorPos(lr.cursorRow, 1)
	fmt.Printf("%s", string(lr.prompt))
	fmt.Printf("%s", string(lr.buffer))
	lr.renderEraseForward(false)
	lr.setCursorPos()
}

func (lr *LineReader) resizeWindow(render bool) {
	lr.lock.Lock()
	defer lr.lock.Unlock()

	lr.winHeight, lr.winWidth = getWindowDimensions()
	if render {
		lr.renderComplete()
	}
}

func (lr *LineReader) setCursorPos() {
	// Determine the linear cursor position, including the prompt
	linearCursorPos := len(lr.prompt) + lr.bufferOffset
	curCursorRow := lr.cursorRow + int(linearCursorPos/lr.winWidth)
	curCursorCol := (linearCursorPos % lr.winWidth) + 1
	setCursorPos(curCursorRow, curCursorCol)
}
