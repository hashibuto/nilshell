# nilshell
Command shell for golang which provides a minimal line editor and command processing loop.  Here's what you get with NilShell:

- Line editor (type, insert, delete)
- Command history (up/down to navigate, load/export)
- Reverse search (simple pattern match, most recent history first)
- Tab completion hook
- Handling of terminal resize

What it doesn't do

- Any sort of argument parsing / tokenization

For a full CLI parser implementation using nilshell, check out [Commander](https://github.com/hashibuto/commander)

## Usage

```
import (
    ns "github.com/hashibuto/nilshell"
)

config := ns.ReaderConfig{

    CompletionFunction: func(beforeCursor string, afterCursor string, full string) *ns.Suggestions {
        // This is where you would return tab completion suggestions based on the input before the cursor, perhaps after the
        // cursor, or even the entire line buffer.

        return &ns.Suggestions{
            Total: 0
            Items: []*ns.Suggestion{}
        }
    },

    ProcessFunction: func(text string) error {
        // text contains the command to be processed by your own command interpreter
        return nil
    }

    // implement your own history manager if you want to persist history across multiple invocations, or use the default (nil)
	HistoryManager: nil,

    PromptFunction: func() string {
        // Return a prompt with terminal escape chars to add style

        return "$ "
    },

    Debug: false,

    // enable the log file to dump debugging info to a tailable log file
    LogFile: "",
}

r := NewReader(config)

// block until the process captures SIGINT or SIGTERM
r.ReadLoop()
```
