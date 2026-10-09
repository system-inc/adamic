package native

import (
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
)

// maybeHarness packs number | undefined into one double and unpacks it, for values given as a present
// flag and hexadecimal bits on stdin, and prints the packed bits and what was read back.
const maybeHarness = `#include "adamic.h"

#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
	char line[64];
	while (fgets(line, sizeof line, stdin) != NULL) {
		char *rest;
		bool present = strtoul(line, &rest, 10) != 0;
		uint64_t bits = strtoull(rest, NULL, 16);
		double number;
		memcpy(&number, &bits, sizeof number);
		double packed = adamic_maybe_number_pack((adamic_maybe_number){present, number});
		uint64_t packed_bits;
		memcpy(&packed_bits, &packed, sizeof packed_bits);
		adamic_maybe_number back = adamic_maybe_number_unpack(packed);
		uint64_t back_bits;
		memcpy(&back_bits, &back.number, sizeof back_bits);
		printf("%016" PRIx64 " %d %016" PRIx64 "\n", packed_bits, back.present, back.present ? back_bits : 0);
	}
	return 0;
}
`

// A field, an element, a cell or a function value holds number | undefined as one double, undefined
// a reserved NaN. Every NaN a program can hold, the reserved pattern itself included, must come back
// present (as the ordinary quiet NaN), undefined must come back undefined, and every other number must
// come back with exactly its bits.
func TestMaybeNumbersPackIntoOneDouble(t *testing.T) {
	t.Parallel()
	type packing struct {
		present bool
		bits    uint64
	}
	const quiet, reserved = 0x7ff8000000000000, 0x7ff8000000000001
	values := []packing{
		{false, 0}, {false, math.Float64bits(1.5)},
		{true, quiet}, {true, 0xfff8000000000000}, {true, reserved}, {true, 0xfff8000000000001},
		{true, 0x7ff0000000000001}, {true, 0x7fffffffffffffff}, {true, 0xffffffffffffffff},
		{true, 0}, {true, math.Float64bits(math.Copysign(0, -1))}, {true, math.Float64bits(math.Inf(1))},
		{true, math.Float64bits(math.Inf(-1))}, {true, math.Float64bits(math.MaxFloat64)}, {true, 1},
	}
	random := rand.New(rand.NewPCG(1984, 1))
	for range 10_000 {
		values = append(values, packing{true, random.Uint64()})
		// NaNs with every kind of payload and sign.
		values = append(values, packing{true, 0x7ff0000000000000 | random.Uint64()&0x800fffffffffffff | 1})
	}
	var input strings.Builder
	for _, value := range values {
		present := 0
		if value.present {
			present = 1
		}
		fmt.Fprintf(&input, "%d %016x\n", present, value.bits)
	}
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(maybeHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	if len(lines) != len(values) {
		t.Fatalf("got %d lines for %d values", len(lines), len(values))
	}
	mismatches := 0
	for index, value := range values {
		want := fmt.Sprintf("%016x 0 %016x", uint64(reserved), 0)
		if value.present {
			kept := value.bits
			if math.IsNaN(math.Float64frombits(kept)) {
				kept = quiet
			}
			want = fmt.Sprintf("%016x 1 %016x", kept, kept)
		}
		if lines[index] != want {
			mismatches++
			if mismatches <= 10 {
				t.Errorf("present %t, bits %016x: got %q, want %q", value.present, value.bits, lines[index], want)
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d values pack wrong", mismatches, len(values))
	}
}
