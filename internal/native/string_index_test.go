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
		adamic_string piece = {{0, adamic_kind_string}, encode(pattern[(state >> 8) % PATTERN_LENGTH], bytes), bytes, 0, NULL};
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
	// One-, two-, three- and four-byte code points, and lone surrogates of each half, which meet in
	// some shifts and stay apart in others.
	replacer := strings.NewReplacer(
		"@PATTERN@", "0x61, 0xe9, 0x4e16, 0x1f30d, 0x62, 0xd800, 0x63, 0x1f600, 0x1f600, 0xdc00, 0x3b1, 0x20, 0xffff, 0x10ffff, 0xdbff, 0xdfff",
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
