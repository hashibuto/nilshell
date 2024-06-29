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

func TestCalculateColumnWidth(t *testing.T) {
	width, numCols := CalculateColumnWidth([]string{"hello", "world", "dog", "cat"}, 80, 2, 2)
	// max column width here is 7 including gutters
	// 80 / 7 is 12
	// we should see 12 columns per row and a width of 7
	assert.Equal(t, 7, width)
	assert.Equal(t, 11, numCols)
}

func TestCalculateColumnWidth2(t *testing.T) {
	width, numCols := CalculateColumnWidth([]string{"this is a longer column", "world", "dog", "cat"}, 40, 2, 2)
	// max column width here is 25 including gutters
	// 40 with min 2 columns means no more than 20 per
	assert.Equal(t, 20, width)
	assert.Equal(t, 2, numCols)
}
