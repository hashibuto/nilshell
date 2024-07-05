package term

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestMeasureNormalString(t *testing.T) {
	length := Measure("hello world")
	assert.Equal(t, 11, length)
}

func TestMeasureWithTerminalCodes(t *testing.T) {
	length := Measure("\x1b[31mhello world\x1b[0m")
	assert.Equal(t, 11, length)
}

func TestMeasureWithTerminalCodesAndUnicode(t *testing.T) {
	length := Measure("\x1b[31mhello world 日本語\x1b[0m")
	assert.Equal(t, 15, length)
}

func TestCropRegularString(t *testing.T) {
	cropped, _ := Crop("hello", 5)
	assert.Equal(t, "hello", cropped)
}

func TestCropRegularString2(t *testing.T) {
	cropped, _ := Crop("hello", 4)
	assert.Equal(t, "h...", cropped)
}

func TestCropWithTerminalCodes(t *testing.T) {
	cropped, _ := Crop("\x1b[31mhello", 5)
	assert.True(t, strings.HasSuffix(cropped, "hello"))
	assert.Equal(t, 5, Measure(cropped))
}

func TestCropWithTerminalCodes2(t *testing.T) {
	cropped, _ := Crop("\x1b[31mhello\x1b[0m", 4)
	assert.True(t, strings.Contains(cropped, "h..."))
	assert.Equal(t, 4, Measure(cropped))
}

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

func TestPaddingRightGuttersRune(t *testing.T) {
	x := PadRight("hello\x1b[0m", 7, 2)
	assert.True(t, strings.HasPrefix(x, "hello"))
	assert.True(t, utf8.RuneCountInString(x) == 7)
}

func TestRemoveEscapeSequences(t *testing.T) {
	src := []byte("\x1b[31mhello\x1b[0m")
	stripped := StripTerminalEscapeSequences(src)
	strippedString := string(stripped)
	assert.Equal(t, "hello", strippedString)
}
