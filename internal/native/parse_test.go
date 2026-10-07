package native

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
)

// parseHarness parses each line: its kind ('i' for parseInt, 'f' for parseFloat), the radix as 16
// hexadecimal digits of its bits, then the text as hexadecimal bytes. It prints each result's bits,
// or NaN.
const parseHarness = `#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int nibble(char character) {
	return character <= '9' ? character - '0' : character - 'a' + 10;
}

int main(void) {
	static char line[8192];
	while (fgets(line, sizeof line, stdin) != NULL) {
		size_t length = strcspn(line, "\n");
		char radix_bits[17];
		memcpy(radix_bits, line + 1, 16);
		radix_bits[16] = '\0';
		uint64_t bits = strtoull(radix_bits, NULL, 16);
		double radix;
		memcpy(&radix, &bits, sizeof radix);
		size_t size = (length - 17) / 2;
		char *bytes = malloc(size + 1);
		for (size_t index = 0; index < size; index++) {
			bytes[index] = (char)(nibble(line[17 + 2 * index]) * 16 + nibble(line[18 + 2 * index]));
		}
		adamic_string text = {{0, adamic_kind_string, 0}, size, bytes, 0, NULL, NULL, 0};
		double result = line[0] == 'i' ? adamic_number_parse_int(&text, radix) : adamic_number_parse_float(&text);
		free(bytes);
		if (isnan(result)) {
			puts("NaN");
			continue;
		}
		uint64_t out;
		memcpy(&out, &result, sizeof out);
		printf("%016llx\n", (unsigned long long)out);
	}
	return 0;
}
`

// parseOracle is the same on Node: Number.parseInt and Number.parseFloat, JavaScript's own answers.
const parseOracle = `const view = new DataView(new ArrayBuffer(8));
const lines = require('node:fs').readFileSync(0, 'utf8').trim().split('\n');
const out = lines.map((line) => {
	view.setBigUint64(0, BigInt('0x' + line.slice(1, 17)));
	const radix = view.getFloat64(0);
	const text = Buffer.from(line.slice(17), 'hex').toString('utf8');
	const result = line[0] === 'i' ? Number.parseInt(text, radix) : Number.parseFloat(text);
	if (Number.isNaN(result)) return 'NaN';
	view.setFloat64(0, result);
	return view.getBigUint64(0).toString(16).padStart(16, '0');
});
process.stdout.write(out.join('\n') + '\n');
`

type parseCase struct {
	parseInt bool
	radix    float64
	text     string
}

// parseSweep is the texts the parsers are held to: every edge the grammar has, and random texts from a
// fixed seed, long enough to reach past 2^53 in every radix, where the rounding lives.
func parseSweep() []parseCase {
	spaces := []string{"", " ", "\t\n", string(rune(0xa0)), string(rune(0x3000)), string(rune(0x2028)), string(rune(0xfeff))}
	cases := []parseCase{}
	integers := []string{
		"", "-", "+", "0", "-0", "+0", "00", "0x", "0X", "0x1f", "0XaB", "-0x10", "0x-1", "0b11", "0o17", "08", "1e21",
		"12abc", "abc", "- 5", "1_000", "9007199254740993", "-9007199254740993", "18446744073709551616",
		"zz", "ZZ", "1.9", ".5", "Infinity",
	}
	radixes := []float64{0, 2, 8, 10, 16, 36, 1, 37, -1, math.NaN(), 16.9, 4294967312, -4294967280, math.Copysign(0, -1), math.Inf(1)}
	for _, text := range integers {
		for _, radix := range radixes {
			cases = append(cases, parseCase{true, radix, text})
		}
	}
	floats := []string{
		"", "-", "+", ".", ".5", "5.", "-.5e-3", "+.e1", "Infinity", "-Infinity", "+Infinity", "Infinityx", "infinity",
		"Inf", "NaN", "0x10", "1e", "1e+", "1e-x", "1E5", "1_000", "-0", "0.0e0", "00012.50",
		"9007199254740993", "9007199254740995", "2.4703282292062327e-324", "2.4703282292062328e-324",
		"4.9e-324", "5e-324", "2e-324", "3e-324", "1.7976931348623157e308", "1.7976931348623158e308",
		"1.7976931348623159e308", "1e309", "1e-400", "2.2250738585072011e-308", "2.2250738585072012e-308",
		"1e99999999999999999999", "0.1", "0.30000000000000004",
	}
	for _, text := range floats {
		for _, space := range spaces {
			cases = append(cases, parseCase{false, 0, space + text}, parseCase{false, 0, space + text + "junk"})
		}
	}
	random := rand.New(rand.NewPCG(1969, 1984))
	digitsOf := func(radix int, count int) string {
		var builder strings.Builder
		for range count {
			digit := random.IntN(radix)
			character := "0123456789abcdefghijklmnopqrstuvwxyz"[digit]
			if random.IntN(2) == 0 && digit >= 10 {
				character -= 'a' - 'A'
			}
			builder.WriteByte(character)
		}
		return builder.String()
	}
	for range 40_000 {
		radix := 2 + random.IntN(35)
		count := 1 + random.IntN(40)
		if radix == 2 || radix == 4 || radix == 8 || radix == 16 || radix == 32 {
			count = 1 + random.IntN(90)
		}
		text := spaces[random.IntN(len(spaces))] + []string{"", "-", "+"}[random.IntN(3)] + digitsOf(radix, count) + []string{"", "", "g!", " 7", "."}[random.IntN(5)]
		given := float64(radix)
		if radix == 10 && random.IntN(2) == 0 {
			given = 0
		}
		cases = append(cases, parseCase{true, given, text})
	}
	for range 4_000 {
		cases = append(cases, parseCase{true, 10, digitsOf(10, 300+random.IntN(120))})
	}
	for range 60_000 {
		var builder strings.Builder
		builder.WriteString([]string{"", "-", "+"}[random.IntN(3)])
		builder.WriteString(digitsOf(10, random.IntN(25)))
		if random.IntN(3) > 0 {
			builder.WriteString(".")
			builder.WriteString(digitsOf(10, random.IntN(25)))
		}
		if random.IntN(2) == 0 {
			builder.WriteString([]string{"e", "E", "e+", "e-"}[random.IntN(4)])
			builder.WriteString(fmt.Sprint(random.IntN(340)))
		}
		cases = append(cases, parseCase{false, 0, builder.String()})
	}
	for range 3_000 {
		// Long mantissas, where only correct rounding gets every bit.
		cases = append(cases, parseCase{false, 0, digitsOf(10, 1+random.IntN(30)) + "." + digitsOf(10, 100+random.IntN(700)) + "e-" + fmt.Sprint(random.IntN(330))})
	}
	return cases
}

func TestNumbersParseExactlyAsJavaScriptDoes(t *testing.T) {
	t.Parallel()
	cases := parseSweep()
	var input strings.Builder
	for _, parse := range cases {
		kind := "f"
		if parse.parseInt {
			kind = "i"
		}
		fmt.Fprintf(&input, "%s%016x%s\n", kind, math.Float64bits(parse.radix), hex.EncodeToString([]byte(parse.text)))
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(parseHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), "node", "--eval", parseOracle), "\n"), "\n")
	if len(native) != len(cases) || len(oracle) != len(cases) {
		t.Fatalf("got %d native lines and %d from Node for %d texts", len(native), len(oracle), len(cases))
	}
	mismatches := 0
	for index, parse := range cases {
		if native[index] != oracle[index] {
			mismatches++
			if mismatches <= 10 {
				t.Errorf("parseInt %t, radix %v, text %q: native %s, Node %s", parse.parseInt, parse.radix, parse.text, native[index], oracle[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d texts parse differently", mismatches, len(cases))
	}
	t.Logf("%d texts, %d mismatches", len(cases), mismatches)
}
