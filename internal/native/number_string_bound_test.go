package native

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The runtime reports its own bound so this test follows changes to the formatter's buffer.
func TestNumberStringBound(t *testing.T) {
	t.Parallel()
	values := []string{
		"-1.2345678901234567e-308", "-2.2250738585072014e-308", "-5e-324",
		"-1.7976931348623157e+308", "-0.0000012345678901234567", "-123456789012345680000",
		"-1.2345678901234567e+21", "-0.000001", "NaN", "-Infinity",
		// Notation boundaries, including values that need seventeen significant digits.
		"-999999999999999900000", "-1.0000000000000001e+21",
		"-9.999999999999997e-7", "-0.0000010000000000000002",
		"-9.999999999999998e-8", "-1.0000000000000001e-7",
	}
	const harness = `#include "adamic.h"

#include <stdio.h>
#include <stdlib.h>

int main(void) {
	printf("%d\n", ADAMIC_NUMBER_STRING_MAX_UNITS);
	char value[64];
	while (scanf("%63s", value) == 1) {
		adamic_string *text = adamic_string_from_number(strtod(value, NULL));
		fwrite(text->bytes, 1, text->length, stdout);
		fputc('\n', stdout);
		adamic_release(text);
	}
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	input := strings.Join(values, "\n") + "\n"
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input, binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input, "node", "--eval", `for (const value of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) console.log(String(Number(value)));`), "\n"), "\n")
	if len(native) != len(values)+1 || len(oracle) != len(values) {
		t.Fatalf("got %d native and %d Node lines for %d values", len(native), len(oracle), len(values))
	}
	bound, err := strconv.Atoi(native[0])
	if err != nil {
		t.Fatal(err)
	}
	longest := 0
	for index, value := range values {
		text := native[index+1]
		t.Logf("%s: native %q, Node %q, length %d", value, text, oracle[index], len(text))
		if text != oracle[index] {
			t.Errorf("%s: native %q, Node %q", value, text, oracle[index])
		}
		// Number text is ascii, so each byte is one code unit.
		for _, character := range text {
			if character > 127 {
				t.Fatalf("non-ascii number text %q", text)
			}
		}
		if len(text) > bound {
			t.Errorf("%s: length %d exceeds bound %d", value, len(text), bound)
		}
		longest = max(longest, len(text))
	}
	if longest != bound {
		t.Errorf("maximum length %d differs from bound %d", longest, bound)
	}
}
