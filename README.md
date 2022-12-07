# nilshell
Command shell for golang which provides a minimal line editor and command processing loop.  Here's what you get with NilShell:

- Line editor (type, insert, delete)
- Command history (up/down to navigate, load/export)
- Reverse search (simple pattern match, most recent history first)
- Tab completion hook
- Handling of terminal resize

What it doesn't do

- Any sort of argument parsing / tokenization

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