package ns

const (
	KEY_CTRL_C      = "\x03" // Signal interrupt
	KEY_CTRL_D      = "\x04" // Signal EOF
	KEY_CTRL_L      = "\x0C" // Clear terminal
	KEY_TAB         = "\x09"
	KEY_ENTER       = "\x0D"
	KEY_CTRL_R      = "\x12" // Search backward
	KEY_CTRL_T      = "\x14"
	KEY_ESCAPE      = "\x1B"
	KEY_BACKSPACE   = "\x7F"
	KEY_DEL         = "\x1B[3~"
	KEY_END         = "\x1B[F"
	KEY_HOME        = "\x1B[H"
	KEY_UP_ARROW    = "\x1B[A"
	KEY_DOWN_ARROW  = "\x1B[B"
	KEY_RIGHT_ARROW = "\x1B[C"
	KEY_LEFT_ARROW  = "\x1B[D"
)
