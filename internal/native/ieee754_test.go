package native

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// ieee754Harness applies one Math function per line: its name, how many operands, then each operand
// as hexadecimal bits. It answers with the result's bits, so the sweep compares bits, not text.
const ieee754Harness = `#include "adamic.h"

#include <inttypes.h>
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

static const struct {
	const char *name;
	double (*function)(double);
} unary[] = {
	{"sin", adamic_math_sin}, {"cos", adamic_math_cos}, {"tan", adamic_math_tan},
	{"asin", adamic_math_asin}, {"acos", adamic_math_acos}, {"atan", adamic_math_atan},
	{"sinh", adamic_math_sinh}, {"cosh", adamic_math_cosh}, {"tanh", adamic_math_tanh},
	{"asinh", adamic_math_asinh}, {"acosh", adamic_math_acosh}, {"atanh", adamic_math_atanh},
	{"exp", adamic_math_exp}, {"expm1", adamic_math_expm1}, {"log", adamic_math_log},
	{"log1p", adamic_math_log1p}, {"log2", adamic_math_log2}, {"log10", adamic_math_log10},
	{"cbrt", adamic_math_cbrt},
};

int main(void) {
	char operation[16], operand[32];
	int count;
	double values[8];
	while (scanf("%15s %d", operation, &count) == 2) {
		if (count < 0 || count > 8) {
			return 1;
		}
		for (int index = 0; index < count; index++) {
			if (scanf("%31s", operand) != 1) {
				return 1;
			}
			values[index] = read(operand);
		}
		double result = NAN;
		if (strcmp(operation, "atan2") == 0) {
			result = adamic_math_atan2(values[0], values[1]);
		} else if (strcmp(operation, "hypot") == 0) {
			result = adamic_math_hypot((size_t)count, values);
		} else {
			size_t index = 0;
			while (index < sizeof unary / sizeof unary[0] && strcmp(unary[index].name, operation) != 0) {
				index++;
			}
			if (index == sizeof unary / sizeof unary[0]) {
				return 1;
			}
			result = unary[index].function(values[0]);
		}
		uint64_t bits;
		memcpy(&bits, &result, sizeof bits);
		printf("%016" PRIx64 "\n", bits);
	}
	return 0;
}
`

const ieee754Oracle = `const view = new DataView(new ArrayBuffer(8));
const read = (text) => { view.setBigUint64(0, BigInt('0x' + text)); return view.getFloat64(0); };
const bits = (value) => { view.setFloat64(0, value); return view.getBigUint64(0).toString(16).padStart(16, '0'); };
const out = [];
for (const line of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) {
	const [operation, , ...operands] = line.split(' ');
	out.push(bits(Math[operation](...operands.map(read))));
}
process.stdout.write(out.join('\n') + '\n');
`

// ieee754Unary are the one-argument functions ported from V8's ieee754.cc.
var ieee754Unary = []string{"sin", "cos", "tan", "asin", "acos", "atan", "sinh", "cosh", "tanh",
	"asinh", "acosh", "atanh", "exp", "expm1", "log", "log1p", "log2", "log10", "cbrt"}

// ieee754BranchPoints are doubles on and beside every branch in the ported code. The code decides by
// comparing a double's high word (and sometimes its low word) with a constant, so every 32-bit
// constant in runtime/ieee754.c is read out of the file itself (the thresholds, and the high words
// its comments give for each coefficient, which cost nothing to try), and the doubles at, just under
// and just over it are tried, with both signs. Decimal thresholds (exp's overflow at
// 709.78...) are found the same way, with their neighbors a few ulps out.
func ieee754BranchPoints(t *testing.T, random *rand.Rand) []float64 {
	source, err := runtime.ReadFile("runtime/ieee754.c")
	if err != nil {
		t.Fatal(err)
	}
	words := map[uint32]bool{}
	for _, match := range regexp.MustCompile(`\b(0x[0-9A-Fa-f]{7,8})\b`).FindAllSubmatch(source, -1) {
		word, err := strconv.ParseUint(string(match[1]), 0, 64)
		if err == nil && word <= math.MaxUint32 {
			words[uint32(word)] = true
		}
	}
	// In order, so the seeded random low words go to the same high words every run: a map's order
	// would make the sweep, and every mismatch it finds, change from run to run.
	ordered := slices.Sorted(maps.Keys(words))
	values := []float64{}
	for _, word := range ordered {
		// The comparisons are against a high word, mostly with the sign masked off, so it means the
		// same as a magnitude; a few (log1p's 0xBFD2BEC4) keep the sign in it.
		for _, high := range []uint32{word - 1, word, word + 1} {
			// The edges of the high word, and a handful of low words between them: a branch that
			// moves by one high word changes answers only somewhere in the doubles it moves.
			lows := []uint32{0, 1, 0x7FFFFFFF, 0xFFFFFFFF}
			for range 12 {
				lows = append(lows, random.Uint32())
			}
			for _, low := range lows {
				value := math.Float64frombits(uint64(high)<<32 | uint64(low))
				values = append(values, value, -value)
			}
		}
	}
	decimals := regexp.MustCompile(`[<>]=?\s*-?([0-9]+\.[0-9]+(?:[eE][-+]?[0-9]+)?)`)
	for _, match := range decimals.FindAllSubmatch(source, -1) {
		value, err := strconv.ParseFloat(string(match[1]), 64)
		if err != nil {
			continue
		}
		for _, step := range []float64{-3, -2, -1, 0, 1, 2, 3} {
			neighbor := math.Float64frombits(uint64(int64(math.Float64bits(value)) + int64(step)))
			values = append(values, neighbor, -neighbor)
		}
	}
	if len(words) < 150 {
		t.Fatalf("found only %d branch constants in runtime/ieee754.c; the pattern has gone stale", len(words))
	}
	return values
}

// ieee754Specials are the values every Math function is tried on by name.
func ieee754Specials() []float64 {
	values := []float64{0, math.Copysign(0, -1), math.NaN(), math.Inf(1), math.Inf(-1),
		math.SmallestNonzeroFloat64, math.Float64frombits(0x000FFFFFFFFFFFFF), math.Float64frombits(0x0010000000000000),
		math.Float64frombits(2), math.Float64frombits(0x0008000000000000), math.MaxFloat64, math.Nextafter(math.MaxFloat64, 0),
		1, 0.5, 2, 3, 10, 0.1, 1e-300, 1e300, 1e-8, 1e-20, 1e16, 1e22, 1e100, 1e200, 1e308,
		math.Pi, math.Pi / 2, math.Pi / 4, 3 * math.Pi / 4, math.E, math.Ln2, math.Log2E, math.Sqrt2, 1 / math.Sqrt2,
		709.782712893384, 709.7822265625, 710.4758600739439, -745.1332191019411, -745.1332191019412, 22, 0.25, 0.75,
		0.5493061443340548, 1.0000000000000002, 0.9999999999999999, 0.49999999999999994, 2.0000000000000004,
		// Trigonometric argument reduction: the doubles nearest the multiples of π/2 that are hardest to
		// reduce (6381956970095103 × 2^797 is the worst case known), large powers of two, and π/2 times
		// small and huge integers.
		math.Ldexp(6381956970095103, 797), math.Ldexp(5706606302225723, 819), math.Ldexp(1, 1023), math.Ldexp(1, 1000),
		math.Ldexp(1, 500), math.Ldexp(1, 100), math.Ldexp(1, 60), math.Ldexp(1, 30), math.Ldexp(1, 20), 1048576 * math.Pi / 2,
		8.98846567431158e307, 1.5707963267948966e300, 3.14159265358979e100, 1e10 * math.Pi,
	}
	for multiple := 1; multiple <= 64; multiple++ {
		values = append(values, float64(multiple)*math.Pi/2, float64(multiple)*math.Pi/4)
	}
	for _, multiple := range []float64{1e3, 1e5, 1e6, 1e8, 1e9, 1e12, 1e15} {
		values = append(values, multiple*math.Pi/2, (multiple+1)*math.Pi/2)
	}
	all := []float64{}
	for _, value := range values {
		all = append(all, value, -value, math.Nextafter(value, math.Inf(1)), math.Nextafter(value, math.Inf(-1)))
	}
	return all
}

func TestIeee754MatchesNodeBitForBit(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewPCG(1993, 2016))
	specials := ieee754Specials()
	values := append(append([]float64{}, specials...), ieee754BranchPoints(t, random)...)
	for range 8000 {
		// Any bits at all: every exponent, NaNs and subnormals included.
		values = append(values, math.Float64frombits(random.Uint64()))
		// The ranges people compute in: around ±1, ±10, ±1000, and every magnitude from 2^-60 to 2^60.
		values = append(values, random.Float64()*2-1, random.Float64()*20-10, random.Float64()*2000-1000)
		values = append(values, math.Ldexp(random.Float64()+0.5, random.IntN(121)-60)*float64(1-2*random.IntN(2)))
	}

	var input strings.Builder
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	lines := 0
	ask := func(operation string, operands ...float64) {
		fmt.Fprintf(&input, "%s %d", operation, len(operands))
		for _, operand := range operands {
			input.WriteString(" " + bits(operand))
		}
		input.WriteByte('\n')
		lines++
	}
	for _, value := range values {
		for _, operation := range ieee754Unary {
			ask(operation, value)
		}
	}
	// Two arguments: every pair of the named values, then random pairs.
	pairs := ieee754Specials()[:4*28]
	for _, left := range pairs {
		for _, right := range pairs {
			ask("atan2", left, right)
			ask("hypot", left, right)
		}
	}
	for range 20000 {
		left, right := values[random.IntN(len(values))], values[random.IntN(len(values))]
		ask("atan2", left, right)
		ask("hypot", left, right)
		ask("hypot", left, right, values[random.IntN(len(values))])
		ask("hypot", random.Float64()*10-5, random.Float64()*10-5, random.Float64()*10-5, random.Float64()*10-5)
	}
	ask("hypot")
	for _, value := range specials {
		ask("hypot", value)
		ask("hypot", value, math.NaN(), math.Inf(-1), 1, 2)
		ask("hypot", 1, 2, 3, 4, 5, value)
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(ieee754Harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), "node", "--eval", ieee754Oracle), "\n"), "\n")
	asked := strings.Split(strings.TrimSuffix(input.String(), "\n"), "\n")
	if len(native) != lines || len(oracle) != lines {
		t.Fatalf("got %d native and %d Node answers for %d questions", len(native), len(oracle), lines)
	}
	same := func(left, right string) bool {
		if left == right {
			return true
		}
		// A program can't see a NaN's payload (0.1 has no typed arrays), so every NaN is the same NaN.
		leftBits, _ := strconv.ParseUint(left, 16, 64)
		rightBits, _ := strconv.ParseUint(right, 16, 64)
		return math.IsNaN(math.Float64frombits(leftBits)) && math.IsNaN(math.Float64frombits(rightBits))
	}
	// Where this Node's V8 contracts multiply-adds, a difference a fused build explains is forgiven
	// (fused_test.go); every other one fails.
	explained, unexplained := contraction(native, oracle, same, func() []string { return fusedAnswers(t, ieee754Harness, input.String(), lines) })
	mismatches := map[string]int{}
	for count, index := range unexplained {
		operation, _, _ := strings.Cut(asked[index], " ")
		mismatches[operation]++
		if count < 20 {
			t.Errorf("%s: native %s, Node %s", asked[index], native[index], oracle[index])
		}
	}
	if len(unexplained) > 0 {
		t.Errorf("%d of %d answers differ, by function: %v", len(unexplained), lines, mismatches)
	}
	t.Logf("%d answers (%d values through each of %d functions, and atan2 and hypot), %d mismatches, %d more that only contraction in this Node explains", lines, len(values), len(ieee754Unary), len(unexplained), explained)
}
