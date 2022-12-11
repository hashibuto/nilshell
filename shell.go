package ns

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

type AutoComplete struct {
	Name string
}

// Completer receives a string of everything before the cursor, after the cursor, and the entire command string.  It
// returns a list of potential suggestions according to the available command set.  Completers are invoked when the user
// presses <tab>, the completion key.
type Completer func(beforeCursor string, afterCursor string, full string) []*AutoComplete

// Executor is called when the <enter> key is pressed after inputting a command
type Executor func(ns *NilShell, cmd string)

type NilShell struct {
	Prompt            string
	History           *History
	AutoCompleteLimit int // Maximum number of autocompletes to display
	Debug             bool

	AutoCompleteSuggestStyle string
	AutoCompleteTooMuchStyle string

	preState   *term.State
	sigs       chan os.Signal
	lineReader *LineReader
	onExecute  Executor
	isShutdown bool
}

// NewShell constructs a NilShell
func NewShell(prompt string, onComplete Completer, onExecute Executor) *NilShell {
	sigs := make(chan os.Signal, 1)
	ns := &NilShell{
		Prompt:            prompt,
		History:           NewHistory(100),
		AutoCompleteLimit: 20,
		sigs:              sigs,
		onExecute:         onExecute,
	}
	ns.lineReader = NewLineReader(onComplete, sigs, ns)
	signal.Notify(ns.sigs, syscall.SIGWINCH)

	return ns
}

// ReadUntilTerm blocks, receiving commands until the user requests termination.  Commands are processed via the executor callback
// provided at initialization time.  Likewise for command completion.
func (n *NilShell) ReadUntilTerm() error {
	for !n.isShutdown {
		cmdString, isTerminate, err := n.lineReader.Read()
		if err != nil {
			return err
		}

		if isTerminate {
			return nil
		}

		if len(cmdString) > 0 {
			n.History.Append(cmdString)
			n.onExecute(n, cmdString)
		}
	}

	return nil
}

// Exit instructs the shell to gracefully exit - this can be safely invoked in an OnExecute method
// to implement a exit command
func (n *NilShell) Shutdown() {
	n.sigs <- syscall.SIGTERM
	n.isShutdown = true
}

// Clear will clear the terminal - this can be safely invoked in an OnExecute method
// to implement a clear command
func (n *NilShell) Clear() {
	clear()
}
