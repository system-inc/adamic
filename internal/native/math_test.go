package native

import (
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
)

// mathHarness applies one function per line: an operation, then operands as hexadecimal bits.
const mathHarness = `#include "adamic.h"

#include <math.h>
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

static void number(double value) {
	char buffer[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, buffer);
	// String(-0) is "0", so a sign that matters is written apart.
	fwrite(buffer, 1, length, stdout);
	fputs(value == 0 && signbit(value) ? " -0\n" : "\n", stdout);
}

int main(void) {
	char operation[16], first[32], second[32];
	while (scanf("%15s %31s %31s", operation, first, second) == 3) {
		double left = read(first), right = read(second);
		if (strcmp(operation, "round") == 0) {
			number(adamic_math_round(left));
		} else if (strcmp(operation, "sign") == 0) {
			number(adamic_math_sign(left));
		} else if (strcmp(operation, "max") == 0) {
			number(adamic_math_max(left, right));
		} else if (strcmp(operation, "min") == 0) {
			number(adamic_math_min(left, right));
		} else if (strcmp(operation, "pow") == 0) {
			number(adamic_power(left, right));
		} else {
			adamic_string *fixed = adamic_number_to_fixed(left, right);
			fwrite(fixed->bytes, 1, fixed->length, stdout);
			fputc('\n', stdout);
			adamic_release(fixed);
		}
	}
	return 0;
}
`

const mathOracle = `const view = new DataView(new ArrayBuffer(8));
const read = (text) => { view.setBigUint64(0, BigInt('0x' + text)); return view.getFloat64(0); };
const number = (value) => String(value) + (Object.is(value, -0) ? ' -0' : '');
const out = [];
for (const line of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) {
	const [operation, first, second] = line.split(' ');
	const left = read(first), right = read(second);
	switch (operation) {
		case 'round': out.push(number(Math.round(left))); break;
		case 'sign': out.push(number(Math.sign(left))); break;
		case 'max': out.push(number(Math.max(left, right))); break;
		case 'min': out.push(number(Math.min(left, right))); break;
		case 'pow': out.push(number(left ** right)); break;
		default: out.push(left.toFixed(right));
	}
}
process.stdout.write(out.join('\n') + '\n');
`

func TestMathAndToFixedMatchJavaScript(t *testing.T) {
	t.Parallel()
	special := []float64{0, math.Copysign(0, -1), math.NaN(), math.Inf(1), math.Inf(-1), 0.5, -0.5, 1.5, -1.5, 2.5, -2.5,
		0.49999999999999994, -0.49999999999999994, 4503599627370495.5, 4503599627370497, 1e21, -1e21, 1e20, 0.1, 0.125, 1.005,
		1.45, 1.55, 8.345, 26.568583470577035, 7.0685834705770345, 123.456, 999.995, 0.000001, 5e-324, math.MaxFloat64, -3.7, 1, -1}
	random := rand.New(rand.NewPCG(628, 1984))
	values := append([]float64{}, special...)
	for range 3000 {
		values = append(values, float64(random.IntN(2_000_000)-1_000_000)/math.Pow(10, float64(random.IntN(7))))
		values = append(values, math.Float64frombits(random.Uint64()))
	}
	var input strings.Builder
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	lines := 0
	for _, value := range values {
		for _, operation := range []string{"round", "sign"} {
			fmt.Fprintf(&input, "%s %s %s\n", operation, bits(value), bits(0))
			lines++
		}
		for _, digits := range []float64{0, 1, 2, 3, 5, 10, 20, 100, 2.7, math.NaN()} {
			if !math.IsNaN(value) && math.Abs(value) < 1e21 || digits == 2 {
				fmt.Fprintf(&input, "fixed %s %s\n", bits(value), bits(digits))
				lines++
			}
		}
	}
	exponents := []float64{-10, -3, -2, -1, -0.5, 0, 0.5, 1, 2, 3, 10, 1.0 / 3, 2.5, -2.5, 0.1}
	for range 6000 {
		base := float64(random.IntN(20_000)-10_000) / math.Pow(10, float64(random.IntN(5)))
		exponent := exponents[random.IntN(len(exponents))]
		if random.IntN(3) == 0 {
			exponent = float64(random.IntN(4000)-2000) / 100
		}
		fmt.Fprintf(&input, "pow %s %s\n", bits(base), bits(exponent))
		lines++
	}
	for _, left := range special {
		for _, right := range special {
			for _, operation := range []string{"max", "min", "pow"} {
				fmt.Fprintf(&input, "%s %s %s\n", operation, bits(left), bits(right))
				lines++
			}
		}
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(mathHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), "node", "--eval", mathOracle), "\n"), "\n")
	asked := strings.Split(strings.TrimSuffix(input.String(), "\n"), "\n")
	if len(native) != lines || len(oracle) != lines {
		t.Fatalf("got %d native and %d Node answers for %d questions", len(native), len(oracle), lines)
	}
	mismatches := 0
	for index := range asked {
		if native[index] != oracle[index] {
			mismatches++
			if mismatches <= 10 {
				t.Errorf("%s: native %q, Node %q", asked[index], native[index], oracle[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d answers differ", mismatches, lines)
	}
	t.Logf("%d answers, %d mismatches", lines, mismatches)
}
