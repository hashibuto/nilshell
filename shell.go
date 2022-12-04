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

type Completer func(beforeCursor string, afterCursor string) []*AutoComplete
type Executor func(ns *NilShell, cmd string)

type NilShell struct {
	preState   *term.State
	sigs       chan os.Signal
	lineReader *LineReader
	prompt     string
	onExecute  Executor
}

func NewShell(prompt string, onComplete Completer, onExecute Executor) *NilShell {
	sigs := make(chan os.Signal, 1)
	ns := &NilShell{
		sigs:       sigs,
		lineReader: NewLineReader(onComplete, sigs),
		prompt:     prompt,
		onExecute:  onExecute,
	}
	signal.Notify(ns.sigs, syscall.SIGWINCH)

	return ns
}

func (n *NilShell) ReadUntilTerm() error {
	fd := int(os.Stdin.Fd())
	preState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	n.preState = preState
	defer term.Restore(fd, n.preState)

	for {
		cmdString, isTerminate, err := n.lineReader.Read(n.prompt)
		if err != nil {
			return err
		}

		if isTerminate {
			return nil
		}

		n.onExecute(n, cmdString)
	}

}
