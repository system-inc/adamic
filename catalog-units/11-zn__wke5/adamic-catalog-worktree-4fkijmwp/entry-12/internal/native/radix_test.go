package native

import (
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
)

// radixHarness writes value.toString(radix) for each line's radix and value, as hexadecimal bits.
const radixHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static double read(const char *text) {
	uint64_t bits = strtoull(text, NULL, 16);
	double value;
	memcpy(&value, &bits, sizeof value);
	return value;
}

int main(void) {
	char radix[32], value[32];
	while (scanf("%31s %31s", radix, value) == 2) {
		adamic_string *text = adamic_number_to_radix(read(value), read(radix));
		fwrite(text->bytes, 1, text->length, stdout);
		fputc('\n', stdout);
		adamic_release(text);
	}
	return 0;
}
`

const radixOracle = `const view = new DataView(new ArrayBuffer(8));
const read = (text) => { view.setBigUint64(0, BigInt('0x' + text)); return view.getFloat64(0); };
const out = [];
for (const line of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) {
	const [radix, value] = line.split(' ');
	out.push(read(value).toString(read(radix)));
}
process.stdout.write(out.join('\n') + '\n');
`

func TestToStringWithARadixMatchesNode(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewPCG(2200, 36))
	values := dtoaValues(random, 300)
	// Integers, which take V8's small-integer path on Node and the same digits here, and the values
	// whose fraction digits carry back into the integer.
	for _, value := range []float64{1, 2, 7, 35, 36, 37, 255, 256, 1295, 1296, 65535, 1 << 30, 1<<31 - 1, 1 << 31, 1<<32 - 1,
		1 << 53, 1<<53 + 2, 1 << 60, 1e21, 1e22, 0.999999999999999, 0.9999999999999999, 1.9999999999999998, 3.9999999999999996,
		0.5, 0.25, 1.0 / 3, 2.0 / 3, 0.7, 255.5, 1e-10, 1e-300} {
		values = append(values, value, -value)
	}
	for range 300 {
		values = append(values, float64(random.IntN(1<<30)), float64(random.Int64N(1<<62)), random.Float64())
	}
	var input strings.Builder
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	lines := 0
	for _, value := range values {
		for radix := 2; radix <= 36; radix++ {
			fmt.Fprintf(&input, "%s %s\n", bits(float64(radix)), bits(value))
			lines++
		}
		// Radixes that ToIntegerOrInfinity brings into range.
		for _, radix := range []float64{2.5, 36.9, 16.0000001} {
			fmt.Fprintf(&input, "%s %s\n", bits(radix), bits(value))
			lines++
		}
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(radixHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), "node", "--eval", radixOracle), "\n"), "\n")
	asked := strings.Split(strings.TrimSuffix(input.String(), "\n"), "\n")
	if len(native) != lines || len(oracle) != lines {
		t.Fatalf("got %d native and %d Node answers for %d questions", len(native), len(oracle), lines)
	}
	mismatches := 0
	for index := range asked {
		if native[index] != oracle[index] {
			mismatches++
			if mismatches <= 20 {
				t.Errorf("%s: native %q, Node %q", asked[index], native[index], oracle[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d answers differ", mismatches, lines)
	}
	t.Logf("%d answers (%d values, every radix from 2 to 36), %d mismatches", lines, len(values), mismatches)
}

// A radix out of range is a RangeError in Node and a panic in Adamic, checked before the value, so
// NaN and the infinities throw too.
func TestToStringWithARadixOutOfRangePanics(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"

#include <stdlib.h>

int main(int count, char **arguments) {
	(void)count;
	static adamic_string before = ADAMIC_STRING("before");
	adamic_write_line(adamic_stdout, &before);
	adamic_string *text = adamic_number_to_radix(strtod(arguments[1], NULL), strtod(arguments[2], NULL));
	adamic_write_line(adamic_stdout, text);
	adamic_release(text);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct {
		value, radix string
		panics       bool
	}{
		{"255", "1", true}, {"255", "37", true}, {"255", "0", true}, {"255", "-2", true}, {"255", "nan", true},
		{"255", "inf", true}, {"255", "1.9", true}, {"nan", "1", true}, {"inf", "40", true}, {"0", "37", true},
		{"255", "36.5", false}, {"nan", "2", false}, {"-inf", "16", false}, {"-0", "2", false},
	} {
		javascript := map[string]string{"nan": "NaN", "inf": "Infinity", "-inf": "-Infinity"}
		value, radix := probe.value, probe.radix
		if name, isNamed := javascript[value]; isNamed {
			value = name
		}
		if name, isNamed := javascript[radix]; isNamed {
			radix = name
		}
		script := fmt.Sprintf("console.log('before'); console.log((%s).toString(%s));", value, radix)
		native := outcome(t, binary, probe.value, probe.radix)
		node := outcome(t, "node", "--eval", script)
		if native.stdout != node.stdout || (native.exitCode == 70) != probe.panics || (node.exitCode != 0) != probe.panics {
			t.Errorf("(%s).toString(%s): native exit %d %q, Node exit %d %q", value, radix, native.exitCode, native.stdout, node.exitCode, node.stdout)
		}
		if probe.panics && (!strings.HasPrefix(native.stderr, "adamic: panic: RangeError: toString() radix argument must be between 2 and 36") || !strings.Contains(node.stderr, "RangeError: toString() radix argument must be between 2 and 36")) {
			t.Errorf("(%s).toString(%s): native stderr %q, Node stderr %q", value, radix, native.stderr, node.stderr)
		}
	}
}
