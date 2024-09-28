package nimble

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// SigFigDur renders duration to a fixed number of significant figures, independent of the
// unit of measure.
func SigFigDur(t time.Duration, sigFigs int) string {
	d := t.String()
	parts := strings.Split(d, ".")
	if len(parts) == 1 {
		return d
	}

	for i := 0; i < len(parts[1]); i++ {
		char := parts[1][i]
		if char < 48 || char > 57 {
			// Suffix starts here
			decStr := parts[1][:i]
			dec, _ := strconv.ParseFloat("."+decStr, 64)
			dec *= math.Pow(10, float64(sigFigs))
			wholeNum, _ := strconv.ParseFloat(fmt.Sprintf("%s.%d", parts[0], int64(math.Round(dec))), 64)
			suffix := parts[1][i:]
			return fmt.Sprintf("%g%s", wholeNum, suffix)
		}
	}

	// Should never get to this point of execution
	return d
}
