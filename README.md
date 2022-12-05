# nilshell
Shell for golang which provides only the bare interface for processing a command.  NilShell is a bare implementation of what's required to build a shell/command processor.  It handles things like entering/exiting raw terminal mode and terminal resize events.  It also provides hooks for command completion as well as execution, and command history.

## Usage

```
import "github.com/hashibuto/nilshell"

ns := NewNilShell(
    "» ", 
    func(beforeCursor, afterCursor string, full string) []*ns.AutoComplete {
        // Autocompletion happens here, perhaps tokenization, and the last token before the cursor is
        // fed to a lookup to find potential matches
        return nil
    },
    func(ns *ns.NilShell, cmd string) {
        // Perform tokenization, command lookup, and execution
    },
)

// Attach saved command history
ns.History = NewHistory(myLoadedHistory)

ns.ReadUntilTerm()

// Save command history
myHistory := ns.History.Export()
// write myHistory it to disk
```