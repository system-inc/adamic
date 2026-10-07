package native

import (
	"path/filepath"
	"strings"
	"testing"
)

// The index sweep: a string's UTF-16 positions found through its index (string_index.c), against
// Node. Long strings are needed for there to be an index at all (64 bytes or more), and they mix
// one-, two-, three- and four-byte code points and lone surrogates of each half, so that pairs and
// lone halves fall on every side of a checkpoint (one every 32 units). Each is read in three orders,
// forward, backward and jumping, since the cursor serves one and the checkpoints the others, and two
// strings are read in turn, each with an index of its own. Every slice of the shorter ones is taken,
// and results are written as WTF-8 in hexadecimal, so a lone surrogate is compared as itself.

const indexHarness = `#include "adamic.h"

#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

static void put_hex(const adamic_string *string) {
	static const char digits[] = "0123456789abcdef";
	putchar(' ');
	for (size_t index = 0; index < string->length; index++) {
		unsigned char byte = (unsigned char)string->bytes[index];
		putchar(digits[byte >> 4]);
		putchar(digits[byte & 15]);
	}
}

static void put_number(double value) {
	adamic_string *written = adamic_string_from_number(value);
	putchar(' ');
	fwrite(written->bytes, 1, written->length, stdout);
	adamic_release(written);
}

static size_t encode(uint32_t point, char *out) {
	if (point < 0x80) {
		out[0] = (char)point;
		return 1;
	}
	if (point < 0x800) {
		out[0] = (char)(0xc0 | (point >> 6));
		out[1] = (char)(0x80 | (point & 0x3f));
		return 2;
	}
	if (point < 0x10000) {
		out[0] = (char)(0xe0 | (point >> 12));
		out[1] = (char)(0x80 | ((point >> 6) & 0x3f));
		out[2] = (char)(0x80 | (point & 0x3f));
		return 3;
	}
	out[0] = (char)(0xf0 | (point >> 18));
	out[1] = (char)(0x80 | ((point >> 12) & 0x3f));
	out[2] = (char)(0x80 | ((point >> 6) & 0x3f));
	out[3] = (char)(0x80 | (point & 0x3f));
	return 4;
}

// pattern is the code points strings are made from, in turn.
static const uint32_t pattern[] = {@PATTERN@};
#define PATTERN_LENGTH (sizeof pattern / sizeof pattern[0])

// make is a heap string of count code points from the pattern, chosen by a small generator seeded
// with shift, so a substring is seldom found anywhere but where it was taken from. It's made as the
// runtime makes strings, so lone halves that meet join into a pair, as they do in JavaScript.
static adamic_string *make(size_t count, size_t shift) {
	adamic_string *result = adamic_string_concat(0, NULL);
	size_t state = shift;
	for (size_t index = 0; index < count; index++) {
		state = (state * 25173 + 13849) % 65536;
		char bytes[4];
		adamic_string piece = {{0, adamic_kind_string, 0}, encode(pattern[(state >> 8) % PATTERN_LENGTH], bytes), bytes, 0, NULL, NULL, 0};
		adamic_string *longer = adamic_string_concat(2, (adamic_string *const[]){result, &piece});
		adamic_release(result);
		result = longer;
	}
	return result;
}

static void read_at(adamic_string *string, double index) {
	put_number(adamic_string_char_code_at(string, index));
	adamic_maybe_number point = adamic_string_code_point_at(string, index);
	put_number(point.present ? point.number : -1);
	adamic_string *at = adamic_string_at(string, index);
	if (at == NULL) {
		fputs(" undefined", stdout);
	} else {
		put_hex(at);
		adamic_release(at);
	}
}

int main(void) {
	static const size_t counts[] = {@COUNTS@};
	for (size_t which = 0; which < sizeof counts / sizeof counts[0]; which++) {
		for (size_t shift = 0; shift < @SHIFTS@; shift++) {
			adamic_string *first = make(counts[which], shift);
			adamic_string *second = make(counts[which] + 7, shift + 3);
			size_t length = (size_t)adamic_string_length(first);
			printf("%zu %zu", counts[which], shift);
			put_number((double)length);
			put_number(adamic_string_length(second));
			putchar('\n');
			// Forward, then backward, each unit of the first string and, in turn, of the second.
			for (size_t index = 0; index <= length; index++) {
				read_at(first, (double)index);
				read_at(second, (double)index);
			}
			putchar('\n');
			for (size_t index = length + 1; index-- > 0;) {
				read_at(first, (double)index);
			}
			putchar('\n');
			// Reads and slices disturb the same cursor. Visit both halves and step back nearby.
			for (size_t index = 1; index < length; index++) {
				read_at(first, (double)index);
				adamic_string *piece = adamic_string_slice(first, (double)(index - 1), (double)(index + 2), true);
				put_hex(piece);
				adamic_release(piece);
				read_at(first, (double)(index - 1));
			}
			putchar('\n');
			// Jumping: a step that wraps around the length, back and forth across checkpoints.
			for (size_t step = 0, at = 5; step < length; step++) {
				at = (at * 37 + 11) % (length + 1);
				read_at(first, (double)at);
			}
			putchar('\n');
			// indexOf of pieces taken from all over the string: its answer in units, from far in.
			for (size_t step = 0, at = 1; step < length; step++) {
				at = (at * 29 + 3) % (length + 1);
				adamic_string *piece = adamic_string_slice(first, (double)at, (double)(at + 1 + step % 6), true);
				put_number(adamic_string_index_of(first, piece));
				adamic_release(piece);
			}
			putchar('\n');
			if (length <= @SLICE_LIMIT@) {
				for (size_t from = 0; from <= length; from++) {
					for (size_t to = from; to <= length; to++) {
						adamic_string *piece = adamic_string_slice(first, (double)from, (double)to, true);
						put_hex(piece);
						adamic_release(piece);
					}
					putchar('\n');
				}
			} else {
				for (size_t step = 0, at = 3; step < length; step++) {
					at = (at * 41 + 7) % (length + 1);
					size_t span = (step * 13) % 70;
					adamic_string *piece = adamic_string_slice(first, (double)at, (double)(at + span), true);
					put_hex(piece);
					adamic_release(piece);
				}
				putchar('\n');
			}
			adamic_release(first);
			adamic_release(second);
		}
	}
	return 0;
}
`

const indexOracle = `const hex = [];
for (let byte = 0; byte < 256; byte++) {
	hex.push(byte.toString(16).padStart(2, '0'));
}
function wtf8(string) {
	let out = ' ';
	for (let index = 0; index < string.length; index++) {
		let unit = string.charCodeAt(index);
		if (unit >= 0xd800 && unit <= 0xdbff && index + 1 < string.length) {
			const low = string.charCodeAt(index + 1);
			if (low >= 0xdc00 && low <= 0xdfff) {
				unit = 0x10000 + ((unit - 0xd800) << 10) + (low - 0xdc00);
				index++;
			}
		}
		if (unit < 0x80) {
			out += hex[unit];
		} else if (unit < 0x800) {
			out += hex[0xc0 | (unit >> 6)] + hex[0x80 | (unit & 63)];
		} else if (unit < 0x10000) {
			out += hex[0xe0 | (unit >> 12)] + hex[0x80 | ((unit >> 6) & 63)] + hex[0x80 | (unit & 63)];
		} else {
			out += hex[0xf0 | (unit >> 18)] + hex[0x80 | ((unit >> 12) & 63)] + hex[0x80 | ((unit >> 6) & 63)] + hex[0x80 | (unit & 63)];
		}
	}
	return out;
}
const pattern = [@PATTERN@];
const character = (point) => (point >= 0xd800 && point <= 0xdfff ? String.fromCharCode(point) : String.fromCodePoint(point));
const make = (count, shift) => {
	let result = '';
	let state = shift;
	for (let index = 0; index < count; index++) {
		state = (state * 25173 + 13849) % 65536;
		result += character(pattern[(state >> 8) % pattern.length]);
	}
	return result;
};
const readAt = (string, index) => {
	const at = string[index];
	return ' ' + string.charCodeAt(index) + ' ' + (string.codePointAt(index) ?? -1) + (at === undefined ? ' undefined' : wtf8(at));
};
const lines = [];
for (const count of [@COUNTS@]) {
	for (let shift = 0; shift < @SHIFTS@; shift++) {
		const first = make(count, shift);
		const second = make(count + 7, shift + 3);
		const length = first.length;
		lines.push(count + ' ' + shift + ' ' + length + ' ' + second.length);
		let line = '';
		for (let index = 0; index <= length; index++) {
			line += readAt(first, index) + readAt(second, index);
		}
		lines.push(line);
		line = '';
		for (let index = length; index >= 0; index--) {
			line += readAt(first, index);
		}
		lines.push(line);
		line = '';
		for (let index = 1; index < length; index++) {
			line += readAt(first, index) + wtf8(first.slice(index - 1, index + 2)) + readAt(first, index - 1);
		}
		lines.push(line);
		line = '';
		for (let step = 0, at = 5; step < length; step++) {
			at = (at * 37 + 11) % (length + 1);
			line += readAt(first, at);
		}
		lines.push(line);
		line = '';
		for (let step = 0, at = 1; step < length; step++) {
			at = (at * 29 + 3) % (length + 1);
			line += ' ' + first.indexOf(first.slice(at, at + 1 + step % 6));
		}
		lines.push(line);
		if (length <= @SLICE_LIMIT@) {
			for (let from = 0; from <= length; from++) {
				line = '';
				for (let to = from; to <= length; to++) {
					line += wtf8(first.slice(from, to));
				}
				lines.push(line);
			}
		} else {
			line = '';
			for (let step = 0, at = 3; step < length; step++) {
				at = (at * 41 + 7) % (length + 1);
				const span = (step * 13) % 70;
				line += wtf8(first.slice(at, at + span));
			}
			lines.push(line);
		}
	}
}
process.stdout.write(lines.join('\n') + '\n');
`

func TestStringIndexMatchesNode(t *testing.T) {
	t.Parallel()
	for _, pattern := range []struct{ name, points string }{
		{"mixed", "0x61, 0xe9, 0x4e16, 0x1f30d, 0x62, 0xd800, 0x63, 0x1f600, 0x1f600, 0xdc00, 0x3b1, 0x20, 0xffff, 0x10ffff, 0xdbff, 0xdfff"},
		// Lone high halves separated from low halves: no supplementary points can form.
		{"BMP", "0x61, 0xe9, 0x4e16, 0x62, 0xd800, 0x63, 0x3b1, 0x20, 0xffff, 0xdbff"},
		{"BMP low halves", "0x61, 0xe9, 0x4e16, 0xdc00, 0xdfff, 0x3b1, 0xffff"},
		{"ASCII", "0x61, 0x62, 0x20, 0x7f"},
	} {
		t.Run(pattern.name, func(t *testing.T) {
			t.Parallel()
			checkStringIndex(t, pattern.points)
		})
	}
}

func checkStringIndex(t *testing.T, pattern string) {
	t.Helper()
	// One-, two-, three- and four-byte code points, and lone surrogates of each half, which meet in
	// some shifts and stay apart in others.
	replacer := strings.NewReplacer(
		"@PATTERN@", pattern,
		"@COUNTS@", "20, 40, 41, 63, 64, 65, 100, 400, 2000",
		"@SHIFTS@", "16",
		"@SLICE_LIMIT@", "140",
	)
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(replacer.Replace(indexHarness), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	oracle := replacer.Replace(indexOracle)
	native := strings.Split(strings.TrimSuffix(runWithInput(t, "", binary), "\n"), "\n")
	node := strings.Split(strings.TrimSuffix(runWithInput(t, "", "node", "--eval", oracle), "\n"), "\n")
	if len(native) != len(node) {
		t.Fatalf("%d native lines, %d from Node", len(native), len(node))
	}
	mismatches := 0
	for index := range native {
		if native[index] != node[index] {
			mismatches++
			if mismatches <= 6 {
				t.Errorf("line %d:\nnative %.300s\nNode   %.300s", index, native[index], node[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d lines differ", mismatches, len(native))
	}
	t.Logf("%d lines, %d mismatches", len(native), mismatches)
}

// Appending a low half joins the previous high half, changing the cached BMP fact as well as the
// units and byte checkpoints. The pointer comparison proves this exercises the in-place path.
func TestStringViewAfterAppendMatchesNode(t *testing.T) {
	t.Parallel()
	harness := strings.Split(indexHarness, "int main(void)")[0] + `int main(void) {
	static adamic_string literal = ADAMIC_STRING("@BMP_LITERAL@");
	adamic_string stack = {{0, adamic_kind_string, 0}, literal.length, literal.bytes, 0, NULL, NULL, 0};
	put_number(adamic_string_length(&literal));
	for (size_t at = 80; at-- > 0;) {
		read_at(&literal, (double)at);
		read_at(&stack, (double)at);
	}
	putchar('\n');
	adamic_string *text = adamic_string_allocate(160);
	for (size_t offset = 0; offset < 160; offset += 2) {
		((char *)text->bytes)[offset] = (char)0xc3;
		((char *)text->bytes)[offset + 1] = (char)0xa9;
	}
	static adamic_string high = ADAMIC_STRING("\xed\xa0\xbc");
	static adamic_string low = ADAMIC_STRING("\xed\xbc\x8d");
	text = adamic_string_append(text, 1, (adamic_string *const[]){&high});
	put_number(adamic_string_length(text));
	read_at(text, 80);
	putchar('\n');
	adamic_string *before = text;
	text = adamic_string_append(text, 1, (adamic_string *const[]){&low});
	put_number(text == before);
	put_number(adamic_string_length(text));
	read_at(text, 80);
	putchar('\n');
	static adamic_string more = ADAMIC_STRING("x\xf0\x9f\x8c\x8d\xc3\xa9");
	for (size_t round = 0; round < 40; round++) {
		text = adamic_string_append(text, 1, (adamic_string *const[]){&more});
		size_t length = adamic_string_units(text);
		put_number((double)length);
		for (size_t at = length + 1; at-- > 0;) {
			read_at(text, (double)at);
			adamic_string *piece = adamic_string_slice(text, (double)(at == 0 ? 0 : at - 1), (double)(at + 2), true);
			put_hex(piece);
			adamic_release(piece);
		}
		putchar('\n');
	}
	adamic_release(text);
	return 0;
}
`
	harness = strings.NewReplacer("@PATTERN@", "0x61", "@BMP_LITERAL@", strings.Repeat(`\303\251`, 80)).Replace(harness)
	// make is unused by this harness, so omit it instead of silencing -Werror.
	from := strings.Index(harness, "// pattern is")
	to := strings.Index(harness, "static void read_at")
	harness = harness[:from] + harness[to:]
	from = strings.Index(harness, "static size_t encode")
	to = strings.Index(harness, "static void read_at")
	harness = harness[:from] + harness[to:]
	oracle := strings.Split(indexOracle, "const pattern =")[0] + `
const readAt = (string, index) => {
	const at = string[index];
	return ' ' + string.charCodeAt(index) + ' ' + (string.codePointAt(index) ?? -1) + (at === undefined ? ' undefined' : wtf8(at));
};
let text = 'é'.repeat(80) + '\ud83c';
const literal = 'é'.repeat(80);
let literalLine = ' ' + literal.length;
for (let at = 79; at >= 0; at--) {
	literalLine += readAt(literal, at) + readAt(literal, at);
}
const lines = [literalLine, ' ' + text.length + readAt(text, 80)];
text += '\udf0d';
lines.push(' 1 ' + text.length + readAt(text, 80));
for (let round = 0; round < 40; round++) {
	text += 'x🌍é';
	let line = ' ' + text.length;
	for (let at = text.length; at >= 0; at--) {
		line += readAt(text, at) + wtf8(text.slice(at === 0 ? 0 : at - 1, at + 2));
	}
	lines.push(line);
}
process.stdout.write(lines.join('\n') + '\n');
`
	binary := filepath.Join(t.TempDir(), "append")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	node := runWithInput(t, "", "node", "--eval", oracle)
	native := runWithInput(t, "", binary)
	if native != node {
		t.Fatal("UTF-16 view after append differs from Node")
	}
	t.Log("43 lines match Node, including literal, stack and in-place append views")
}

// Exercise the literal sentinel, an unindexed stack string, and a heap cache before and after
// construction. indexOf runs before any length query, so units_before must accept an empty cache.
func TestStringIndexCacheStatesMatchNode(t *testing.T) {
	t.Parallel()
	text := strings.Repeat("é😀A", 24) + "Z"
	source := `#include "adamic.h"
#include <stdio.h>
static adamic_string literal = ADAMIC_STRING(@TEXT@);
static adamic_string suffix = ADAMIC_STRING("!");
static adamic_string needle = ADAMIC_STRING("Z");
int main(void) {
    adamic_string stack = literal;
    stack.index = NULL;
    adamic_string *heap = adamic_string_concat(2, (adamic_string *const[]){&literal, &suffix});
    printf("%.0f %.0f %.0f\n", adamic_string_index_of(heap, &needle), adamic_string_index_of(&literal, &needle), adamic_string_index_of(&stack, &needle));
    const adamic_string *texts[] = {&literal, &stack, heap};
    for (int pass = 0; pass < 2; pass++) {
        for (int which = 0; which < 3; which++) {
            for (double at = adamic_string_length(texts[which]) - 1; at >= 0; at--) {
                printf("%.0f ", adamic_string_char_code(texts[which], at));
            }
            putchar('\n');
        }
    }
    adamic_release(heap);
    return 0;
}`
	binary := filepath.Join(t.TempDir(), "cache-states")
	if err := Build(strings.ReplaceAll(source, "@TEXT@", cString(text))+"\n", binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	node := `const literal = "é😀A".repeat(24) + "Z";
const texts = [literal, literal, literal + "!"];
console.log(texts[2].indexOf("Z"), texts[0].indexOf("Z"), texts[1].indexOf("Z"));
for (let pass = 0; pass < 2; pass++) {
    for (const text of texts) {
        const units = [];
        for (let at = text.length - 1; at >= 0; at--) units.push(text.charCodeAt(at));
        console.log(units.join(" ") + " ");
    }
}`
	want := runWithInput(t, "", "node", "--eval", node)
	got := runWithInput(t, "", binary)
	if got != want {
		t.Fatalf("cache states differ: native %q; Node %q", got, want)
	}
}
