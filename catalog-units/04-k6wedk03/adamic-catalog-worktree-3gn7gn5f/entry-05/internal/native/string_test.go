package native

import (
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

// stringHarness answers one question per line: an operation, two strings as hexadecimal UTF-8, and
// two numbers as hexadecimal bits.
const stringHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static adamic_string *text(const char *hex) {
	size_t length = strlen(hex) / 2;
	char *bytes = malloc(length + 1);
	for (size_t index = 0; index < length; index++) {
		unsigned byte;
		sscanf(hex + index * 2, "%2x", &byte);
		bytes[index] = (char)byte;
	}
	adamic_string piece = {{0, adamic_kind_string, 0}, length, bytes, 0, NULL, NULL, 0};
	adamic_string *string = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(bytes);
	return string;
}

static double number(const char *hex) {
	uint64_t bits = strtoull(hex, NULL, 16);
	double value;
	memcpy(&value, &bits, sizeof value);
	return value;
}

static void say_number(double value) {
	adamic_string *written = adamic_string_from_number(value);
	adamic_write_line(adamic_stdout, written);
	adamic_release(written);
}

static void say(adamic_string *string) {
	adamic_write_line(adamic_stdout, string);
	adamic_release(string);
}

int main(void) {
	char operation[16], left_hex[256], right_hex[256], first_hex[32], second_hex[32];
	while (scanf("%15s %255s %255s %31s %31s", operation, left_hex, right_hex, first_hex, second_hex) == 5) {
		adamic_string *left = text(strcmp(left_hex, "-") == 0 ? "" : left_hex);
		adamic_string *right = text(strcmp(right_hex, "-") == 0 ? "" : right_hex);
		double first = number(first_hex), second = number(second_hex);
		if (strcmp(operation, "length") == 0) {
			say_number(adamic_string_length(left));
		} else if (strcmp(operation, "charCodeAt") == 0) {
			say_number(adamic_string_char_code_at(left, first));
		} else if (strcmp(operation, "codeCounted") == 0) {
			// The length counted first, as a loop's condition does, so an ASCII string is read inline.
			(void)adamic_string_length(left);
			say_number(adamic_string_char_code_at(left, first));
		} else if (strcmp(operation, "codePointAt") == 0) {
			adamic_maybe_number point = adamic_string_code_point_at(left, first);
			if (point.present) {
				say_number(point.number);
			} else {
				say(adamic_string_concat(1, (adamic_string *const[]){&(adamic_string)ADAMIC_STRING("undefined")}));
			}
		} else if (strcmp(operation, "slice") == 0) {
			say(adamic_string_slice(left, first, second, true));
		} else if (strcmp(operation, "sliceFrom") == 0) {
			say(adamic_string_slice(left, first, 0, false));
		} else if (strcmp(operation, "rejoin") == 0) {
			adamic_string *head = adamic_string_slice(left, 0, first, true);
			adamic_string *tail = adamic_string_slice(left, first, 0, false);
			adamic_string *joined = adamic_string_concat(2, (adamic_string *const[]){head, tail});
			say_number(adamic_string_equal(joined, left) ? 1 : 0);
			adamic_release(head);
			adamic_release(tail);
			adamic_release(joined);
		} else if (strcmp(operation, "compare") == 0) {
			say_number(adamic_string_compare(left, right));
		} else if (strcmp(operation, "indexOf") == 0) {
			say_number(adamic_string_index_of(left, right));
		} else if (strcmp(operation, "startsWith") == 0) {
			say_number(adamic_string_starts_with(left, right) ? 1 : 0);
		} else if (strcmp(operation, "endsWith") == 0) {
			say_number(adamic_string_ends_with(left, right) ? 1 : 0);
		} else if (strcmp(operation, "padStart") == 0) {
			say(adamic_string_pad(left, first, right, true));
		} else if (strcmp(operation, "padEnd") == 0) {
			say(adamic_string_pad(left, first, right, false));
		} else if (strcmp(operation, "trim") == 0) {
			say(adamic_string_trim(left));
		} else if (strcmp(operation, "points") == 0) {
			adamic_array *points = adamic_string_code_points(left);
			say_number((double)points->length);
			adamic_release(points);
		}
		adamic_release(left);
		adamic_release(right);
	}
	return 0;
}
`

const stringOracle = `const view = new DataView(new ArrayBuffer(8));
const number = (hex) => { view.setBigUint64(0, BigInt('0x' + hex)); return view.getFloat64(0); };
const text = (hex) => hex === '-' ? '' : Buffer.from(hex, 'hex').toString('utf8');
const out = [];
for (const line of require('node:fs').readFileSync(0, 'utf8').trim().split('\n')) {
	const [operation, leftHex, rightHex, firstHex, secondHex] = line.split(' ');
	const left = text(leftHex), right = text(rightHex), first = number(firstHex), second = number(secondHex);
	switch (operation) {
		case 'length': out.push(String(left.length)); break;
		case 'charCodeAt': out.push(String(left.charCodeAt(first))); break;
		case 'codeCounted': out.push(String(left.charCodeAt(first))); break;
		case 'codePointAt': out.push(String(left.codePointAt(first))); break;
		case 'slice': out.push(left.slice(first, second)); break;
		case 'sliceFrom': out.push(left.slice(first)); break;
		case 'rejoin': out.push(left.slice(0, first) + left.slice(first) === left ? '1' : '0'); break;
		case 'compare': out.push(String(left < right ? -1 : left > right ? 1 : 0)); break;
		case 'indexOf': out.push(String(left.indexOf(right))); break;
		case 'startsWith': out.push(left.startsWith(right) ? '1' : '0'); break;
		case 'endsWith': out.push(left.endsWith(right) ? '1' : '0'); break;
		case 'padStart': out.push(left.padStart(first, right)); break;
		case 'padEnd': out.push(left.padEnd(first, right)); break;
		case 'trim': out.push(left.trim()); break;
		case 'points': out.push(String([...left].length)); break;
	}
}
process.stdout.write(out.join('\n') + '\n');
`

func TestStringsMatchJavaScript(t *testing.T) {
	t.Parallel()
	texts := []string{"", "a", "ab", "abc", "héllo, 世界 🌍", "🌍🌍", "a🌍b", "\uffff", "𝄞x", "e\u0301", "  trim\t ", "\u3000\u00a0 wide \ufeff",
		"ä", "zz", "Z", "日本語のテキスト", "🌍", "x🌍", "-", "ab🌍cd🌍"}
	numbers := []float64{-3, -1, 0, 1, 2, 3, 4, 5, 7, 9, 10, 11, 12, 20, 1.7, -0.5, math.NaN(), math.Inf(1), math.Inf(-1)}
	encode := func(value string) string {
		if value == "" {
			return "-"
		}
		return hex.EncodeToString([]byte(value))
	}
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	var input strings.Builder
	lines := 0
	ask := func(operation string, left string, right string, first float64, second float64) {
		fmt.Fprintf(&input, "%s %s %s %s %s\n", operation, encode(left), encode(right), bits(first), bits(second))
		lines++
	}
	for _, left := range texts {
		ask("length", left, "", 0, 0)
		ask("trim", left, "", 0, 0)
		ask("points", left, "", 0, 0)
		for _, first := range numbers {
			ask("charCodeAt", left, "", first, 0)
			ask("codeCounted", left, "", first, 0)
			ask("codePointAt", left, "", first, 0)
			ask("sliceFrom", left, "", first, 0)
			ask("rejoin", left, "", first, 0)
			if !math.IsInf(first, 0) {
				// Padding to an infinite length throws on both sides, with different words.
				ask("padStart", left, "-", first, 0)
				ask("padEnd", left, "🌍.", first, 0)
			}
			for _, second := range numbers {
				ask("slice", left, "", first, second)
			}
		}
		for _, right := range texts {
			ask("compare", left, right, 0, 0)
			ask("indexOf", left, right, 0, 0)
			ask("startsWith", left, right, 0, 0)
			ask("endsWith", left, right, 0, 0)
		}
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(stringHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), binary), "\n"), "\n")
	oracle := strings.Split(strings.TrimSuffix(runWithInput(t, input.String(), "node", "--eval", stringOracle), "\n"), "\n")
	asked := strings.Split(strings.TrimSuffix(input.String(), "\n"), "\n")
	if len(native) != lines || len(oracle) != lines {
		t.Fatalf("got %d native and %d Node answers for %d questions", len(native), len(oracle), lines)
	}
	mismatches := 0
	for index := range asked {
		if native[index] != oracle[index] {
			mismatches++
			if mismatches <= 12 {
				t.Errorf("%s: native %q, Node %q", asked[index], native[index], oracle[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d answers differ", mismatches, lines)
	}
	t.Logf("%d answers, %d mismatches", lines, mismatches)
}
