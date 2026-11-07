package native

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

// Every power of two a double can be, from 2^-1074 to 2^1023, with the two doubles on each side and
// all their negatives. A power of two is where the doubles below are twice as dense as the doubles
// above, so the shortest digits that read back aren't always the closest digits of that length:
// String(2 ** 976) is 6.386688990511104e+293, and asking for the closest 16-digit decimal (as
// number.c once did) gives one that reads back as the double below, then 17 digits. The 255,599
// values of TestNumbersFormatExactlyAsJavaScriptDoes never landed on one of the 46 that differed.
func TestPowersOfTwoFormatAsJavaScriptDoes(test *testing.T) {
	test.Parallel()
	values := []float64{}
	for exponent := -1074; exponent <= 1023; exponent++ {
		power := math.Ldexp(1, exponent)
		below, above := math.Nextafter(power, 0), math.Nextafter(power, math.Inf(1))
		values = append(values, power, below, above, math.Nextafter(below, 0), math.Nextafter(above, math.Inf(1)))
	}
	for _, value := range values[:len(values):len(values)] {
		values = append(values, -value)
	}
	var input strings.Builder
	for _, value := range values {
		fmt.Fprintf(&input, "%016x\n", math.Float64bits(value))
	}

	binary := filepath.Join(test.TempDir(), "harness")
	if buildError := Build(numberHarness, binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(test, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(
		strings.TrimSuffix(runWithInput(test, input.String(), "node", "--eval", numberOracle), "\n"),
		"\n",
	)
	if len(native) != len(values) || len(oracle) != len(values) {
		test.Fatalf("got %d native lines and %d from Node for %d values", len(native), len(oracle), len(values))
	}
	mismatches, powers := 0, 0
	for index, value := range values {
		if native[index] == oracle[index] {
			continue
		}
		mismatches++
		if _, exponent := math.Frexp(math.Abs(value)); math.Abs(value) == math.Ldexp(0.5, exponent) {
			powers++
		}
		if mismatches <= 10 {
			test.Errorf("%016x: native %q, Node %q", math.Float64bits(value), native[index], oracle[index])
		}
	}
	if mismatches > 0 {
		test.Errorf("%d of %d values format differently (%d of them powers of two)", mismatches, len(values), powers)
	}
	test.Logf(
		"%d values (2,098 powers of two, two neighbors each side, and their negatives), %d mismatches",
		len(values),
		mismatches,
	)
}
