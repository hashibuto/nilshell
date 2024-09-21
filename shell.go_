package ns

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

type AutoComplete struct {
	Value   string
	Display string
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
	DumpFile          string

	AutoCompleteSuggestStyle string
	AutoCompleteTooMuchStyle string

	preState   *term.State
	sigs       chan os.Signal
	lineReader *LineReader
	onExecute  Executor
	isShutdown bool
	dumpChan   chan []byte
	wg         sync.WaitGroup
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
		Debug:             true,
		dumpChan:          make(chan []byte, 3000),
	}
	ns.lineReader = NewLineReader(onComplete, sigs, ns)
	signal.Notify(ns.sigs, syscall.SIGWINCH)

	return ns
}

// ReadUntilTerm blocks, receiving commands until the user requests termination.  Commands are processed via the executor callback
// provided at initialization time.  Likewise for command completion.
func (n *NilShell) ReadUntilTerm() error {
	if n.DumpFile != "" {
		n.wg.Add(1)
		n.lineReader.DumpChan = n.dumpChan

		ofile, err := os.OpenFile(n.DumpFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer ofile.Close()

		go func() {
			defer n.wg.Done()

			t := time.NewTicker(time.Second)
			var buffer bytes.Buffer
			for {
				select {
				case bArray := <-n.dumpChan:
					if bArray == nil {
						t.Stop()
						if buffer.Len() > 0 {
							ofile.Write(buffer.Bytes())
							buffer.Reset()
						}
						return
					}

					for _, b := range bArray {
						if b > 32 && b < 127 {
							buffer.WriteString(fmt.Sprintf("%c", b))
						} else {
							buffer.WriteString(fmt.Sprintf("<0x%02X>", b))
						}
					}
					buffer.WriteString("\n")
				case <-t.C:
					if buffer.Len() > 0 {
						ofile.Write(buffer.Bytes())
						buffer.Reset()
					}
				}
			}
		}()
	}

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
			fmt.Printf("\r")
			n.onExecute(n, cmdString)
		}
	}
	close(n.dumpChan)
	n.wg.Wait()

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
