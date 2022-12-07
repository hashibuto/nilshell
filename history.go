package ns

import "strings"

type History struct {
	commands     []string
	commandIndex int
	maxKeep      int
}

// NewHistory creates a new history object with optional pre-loaded history
func NewHistory(maxKeep int, commands ...string) *History {
	newCommands := []string{}
	newCommands = append(newCommands, commands...)
	if len(newCommands) > maxKeep {
		newCommands = newCommands[len(newCommands)-maxKeep:]
	}
	return &History{
		commands:     newCommands,
		commandIndex: -1,
		maxKeep:      maxKeep,
	}
}

// Older returns the next oldest command in the history
func (h *History) Older() string {
	if len(h.commands) == 0 {
		return ""
	}

	if h.commandIndex == -1 {
		h.commandIndex = len(h.commands) - 1
	} else {
		h.commandIndex--
		if h.commandIndex < 0 {
			h.commandIndex = 0
		}
	}

	return h.commands[h.commandIndex]
}

// Newer returns the next newest command in the history
func (h *History) Newer() string {
	if len(h.commands) == 0 {
		return ""
	}

	if h.commandIndex == -1 {
		h.commandIndex = len(h.commands) - 1
	} else {
		h.commandIndex++
		if h.commandIndex >= len(h.commands) {
			h.commandIndex = len(h.commands) - 1
			return ""
		}
	}

	return h.commands[h.commandIndex]
}

// Any returns true if there are any commands in the history
func (h *History) Any() bool {
	return len(h.commands) > 0
}

// Append appends another command to the history
func (h *History) Append(command string) {
	if len(h.commands) > 0 {
		// Don't append consecutive duplicates (this isn't an audit history, its a convenience history)
		if h.commands[len(h.commands)-1] == command {
			return
		}
	}
	h.commands = append(h.commands, command)
	if len(h.commands) > h.maxKeep {
		h.commands = h.commands[len(h.commands)-h.maxKeep:]
	}

	h.commandIndex = -1
}

// Export returns the command history
func (h *History) Export() []string {
	return h.commands
}

// FindMostRecentMatch returns the most recent command that contains subString, or an empty string
func (h *History) FindMostRecentMatch(subString string) string {
	if subString == "" {
		return ""
	}
	for i := len(h.commands) - 1; i >= 0; i-- {
		cmd := h.commands[i]
		if strings.Contains(cmd, subString) {
			return cmd
		}
	}
	return ""
}
