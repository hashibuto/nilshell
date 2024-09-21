package ns

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashibuto/nilshell/pkg/termutils"
	"golang.org/x/term"
)

var (
	ErrInterrupt = errors.New("interrupt")
	ErrEof       = errors.New("eof")
)

type Suggestion struct {
	Display string
	Value   string
}

type Suggestions struct {
	Count int
	Items []*Suggestion
}

type Reader struct {
	config            ReaderConfig
	editOffset        int
	prevEditOffset    int
	historyBuffer     []string
	logFile           *os.File
	readBuffer        []rune
	requireFullRender bool
	searchMode        bool
	signalChan        chan os.Signal
	windowSize        *Size
	renderPosition    Position
	windowSizeLock    sync.Mutex
	waitGroup         sync.WaitGroup
}

type Size struct {
	Rows    int
	Columns int
}

type Position struct {
	Row    int
	Column int
}

type CompletionFunc func(beforeCursor string, afterCursor string, full string) *Suggestions

type ReaderConfig struct {
	CompletionFunction  CompletionFunc
	ProcessFunction     func(string) error
	HistoryAddFunction  func(string)
	HistoryLoadFunction func() []string
	PromptFunction      func() string
	Debug               bool
	LogFile             string
}

func NewReader(config ReaderConfig) *Reader {
	if config.CompletionFunction == nil {
		config.CompletionFunction = func(beforeCursor, afterCursor, full string) *Suggestions {
			return &Suggestions{
				Count: 0,
				Items: []*Suggestion{},
			}
		}

		if config.ProcessFunction == nil {
			config.ProcessFunction = func(s string) error {
				return nil
			}
		}

		if config.HistoryAddFunction == nil {
			config.HistoryAddFunction = func(s string) {
			}
		}

		if config.HistoryLoadFunction == nil {
			config.HistoryLoadFunction = func() []string {
				return []string{}
			}
		}

		if config.PromptFunction == nil {
			config.PromptFunction = func() string {
				return "$ "
			}
		}
	}

	return &Reader{
		config:        config,
		historyBuffer: config.HistoryLoadFunction(),
		signalChan:    make(chan os.Signal, 10),
		readBuffer:    []rune{},
	}
}

// ReadLoop reads commands from the standard input and blocks until exit
func (r *Reader) ReadLoop() error {
	if r.config.LogFile != "" {
		var err error
		r.logFile, err = os.OpenFile(r.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer r.logFile.Close()
	}

	r.waitGroup.Add(1)
	go r.sigThread()
	defer r.waitGroup.Wait()

	r.GetWindowSize()
	termutils.RequestCursorPos()
	r.requireFullRender = true

	for {
		value, err := r.readLine()
		switch err {
		case ErrEof:
			r.signalChan <- syscall.SIGHUP
			return nil
		case ErrInterrupt:
			continue
		}

		if err != nil {
			return err
		}

		// do some processing on value if anything to process
		if len(value) == 0 {
			continue
		}

		err = r.config.ProcessFunction(value)
		if err == ErrEof {
			r.signalChan <- syscall.SIGHUP
			break
		}

		if err != nil {
			return err
		}
		termutils.RequestCursorPos()
	}

	return nil
}

func (r *Reader) GetWindowSize() *Size {
	r.windowSizeLock.Lock()
	defer r.windowSizeLock.Unlock()

	if r.windowSize == nil {
		r.windowSize = &Size{}
		r.windowSize.Rows, r.windowSize.Columns = termutils.GetWindowSize()
	}

	return r.windowSize
}

func (r *Reader) sigThread() {
	defer r.waitGroup.Done()

	for sig := range r.signalChan {
		switch sig {
		case syscall.SIGHUP:
			// This comes from within the application
			return
		case syscall.SIGINT, syscall.SIGTERM:
			// This should cause the main stdin loop to exit
			_, err := os.Stdin.WriteString(KEY_CTRL_D)
			if err != nil {
				slog.Error(err.Error())
			}
		case syscall.SIGWINCH:
			// System indicates that a resize of the terminal window occurred
			rows, cols := termutils.GetWindowSize()
			r.windowSizeLock.Lock()
			if r.windowSize == nil {
				r.windowSize = &Size{}
			}

			r.windowSize.Rows = rows
			r.windowSize.Columns = cols
			termutils.RequestCursorPos()
			r.windowSizeLock.Unlock()
		}
	}
}

// ReadInput displays the prompt and reads input until a terminating character is encountered.  Terminating characters include <Enter>,
// <Ctrl+D>, and <Ctrl+C>.
func (r *Reader) readLine() (string, error) {
	renderLines := 0
	r.editOffset = 0
	r.prevEditOffset = 0
	isNewLine := true

	stdioFd := int(os.Stdin.Fd())
	preState, err := term.MakeRaw(stdioFd)
	if err != nil {
		return "", err
	}

	defer func() {
		r.readBuffer = []rune{}
		rErr := recover()

		err = term.Restore(stdioFd, preState)
		if err != nil {
			log.Fatalf("fatal error: unable to restore terminal: %v", err)
		}

		if rErr != nil {
			fmt.Printf("caught panic in called function\n%v\n", err)
			if r.config.Debug {
				fmt.Println(string(debug.Stack()))
			}
		}

		fmt.Printf("\n")
		r.renderPosition.Row += renderLines + 1
		if r.renderPosition.Row > r.windowSize.Rows {
			r.renderPosition.Row = r.windowSize.Rows
		}
		r.searchMode = false
	}()

	stdinBuf := make([]byte, 100)
	for {
		termutils.HideCursor()
		renderLines = r.render(isNewLine)
		isNewLine = false
		r.SetEditCursorPosition()
		termutils.ShowCursor()

		nBytesRead, err := os.Stdin.Read(stdinBuf)
		if err != nil {
			return "", err
		}
		inputData := string(stdinBuf[:nBytesRead])

		switch inputData {
		case KEY_CTRL_C:
			return "", ErrInterrupt
		case KEY_CTRL_D:
			return "", ErrEof
		case KEY_ENTER:
			if r.searchMode {
				r.requireFullRender = true
				r.searchMode = false
				continue
			}
			return strings.Trim(string(r.readBuffer), " \t\r\n"), nil
		case KEY_CTRL_R:
			if r.searchMode {
				continue
			}
			r.requireFullRender = true
			r.searchMode = true
		case KEY_LEFT_ARROW:
			if r.editOffset > 0 {
				r.editOffset--
			}
		case KEY_RIGHT_ARROW:
			if r.editOffset < len(r.readBuffer) {
				r.editOffset++
			}
		case KEY_CTRL_L:
			termutils.ClearTerminal()
			termutils.SetCursorPos(1, 1)
		default:
			if r.parseControlSequence(inputData) {
				continue
			}

			if !r.searchMode {
				r.updateBuffer(inputData)
			}
		}
	}
}

func (r *Reader) updateBuffer(data string) {
	cutBegin := r.editOffset
	cutEnd := r.editOffset

	switch data {
	case KEY_DEL:
		cutEnd++
		data = ""
	case KEY_BACKSPACE:
		if cutBegin > 0 {
			cutBegin--
		}
		data = ""
	}

	newRunes := []rune{}
	if cutBegin > 0 {
		newRunes = append(newRunes, r.readBuffer[:cutBegin]...)
	}
	if len(data) > 0 {
		newRunes = append(newRunes, []rune(data)...)
	}
	if cutEnd < len(r.readBuffer) {
		newRunes = append(newRunes, r.readBuffer[cutEnd:]...)
	}

	if cutBegin != r.editOffset {
		r.editOffset = cutBegin
	} else {
		r.editOffset += termutils.Measure(data)
	}

	r.readBuffer = newRunes
}

// render renders the edit "line" and returns the number of screen rows used in the render
func (r *Reader) render(isNewLine bool) int {
	length := 0
	prompt := r.getCurrentPrompt()
	suffix := ""
	if r.searchMode {
		suffix = "\""
	}

	if r.requireFullRender {
		fmt.Printf("%s%s%s", prompt, string(r.readBuffer), suffix)
		termutils.ClearTerminalFromCursor()
		r.requireFullRender = false
		length = termutils.Measure(prompt) + termutils.Measure(string(r.readBuffer)) + termutils.Measure(suffix)
	} else if isNewLine {
		// this is the first time rendering this line, we want to render the prompt
		fmt.Printf("%s", prompt)
		length = termutils.Measure(prompt)
	} else {
		pos := r.prevEditOffset
		if r.editOffset < pos {
			pos = r.editOffset
		}
		r.SetEditCursorPosition(pos)
		r.log(fmt.Sprintf("POS: %d  BUF: %s", pos, string(r.readBuffer)))

		fmt.Printf("%s", string(r.readBuffer[pos:]))
		termutils.ClearLineFromCursor()
		length = termutils.Measure(prompt) + termutils.Measure(string(r.readBuffer))
	}
	r.prevEditOffset = r.editOffset

	return length / r.windowSize.Columns
}

func (r *Reader) parseControlSequence(input string) bool {
	if !strings.HasPrefix(input, "\x1B") {
		return false
	}

	row, col, err := termutils.GetCursorPosition(input)
	if err == nil {
		// we reset the cursor position right after we receive a new one, b/c this indicates that a terminal resize
		// occurred and we need to perform a full render from the beginning of the current input.
		r.resetStartingCursorPosition(row, col)
		return true
	}

	return true
}

func (r *Reader) getCurrentPrompt() string {
	if !r.searchMode {
		return r.config.PromptFunction()
	}

	return "(reverse-i-search) \""
}

// resetsCursorPosition sets the cursor position to the beginning of the current rendering position.
// It calculates the position based on the current
func (r *Reader) resetStartingCursorPosition(row int, col int) {
	r.log(fmt.Sprintf("CURSOR POS: r:%d c:%d", row, col))
	if len(r.readBuffer) > 0 {
		promptLen := termutils.Measure(r.getCurrentPrompt())
		col -= (termutils.Measure(string(r.readBuffer)) + promptLen)
	}

	for col < 1 {
		col += r.windowSize.Columns
		row--
	}

	if row < 1 {
		row = 1
	}

	col = 1

	r.renderPosition.Row = row
	r.renderPosition.Column = col
	r.log(fmt.Sprintf("CURSOR POS NEW: r:%d c:%d", row, col))

	r.requireFullRender = true
	termutils.SetCursorPos(r.renderPosition.Row, r.renderPosition.Column)
}

func (r *Reader) SetEditCursorPosition(offset ...int) {
	pos := r.editOffset
	if len(offset) > 0 {
		pos = offset[0]
	}
	numCols := termutils.Measure(r.getCurrentPrompt()) + pos
	col := 1 + (numCols % r.windowSize.Columns)
	row := r.renderPosition.Row + (numCols / r.windowSize.Columns)
	termutils.SetCursorPos(row, col)
}

func (r *Reader) log(msg string) {
	if r.logFile == nil {
		return
	}

	r.logFile.WriteString(fmt.Sprintf("%s: %s\n", time.Now(), msg))
	r.logFile.Sync()
}
