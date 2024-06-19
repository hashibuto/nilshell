package ns

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaddingRight(t *testing.T) {
	x := PadRight("hello", 20, 2)
	assert.True(t, strings.HasPrefix(x, "hello "))
	assert.True(t, len(x) == 20)
}

func TestPaddingRightGutters(t *testing.T) {
	x := PadRight("hello", 7, 2)
	assert.True(t, strings.HasPrefix(x, "hello "))
	assert.True(t, len(x) == 7)
}

func TestPaddingRightGutters2(t *testing.T) {
	x := PadRight("hello", 6, 2)
	assert.True(t, strings.HasPrefix(x, "h... "))
	assert.True(t, len(x) == 6)
}
