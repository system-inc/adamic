package native

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The normalization sweep: native normalize against Node's, in all four forms, byte for byte as
// WTF-8, over every code point on its own and in four places composition looks at its neighbours
// (between a starter and a mark, before marks out of canonical order, after a Hangul leading
// consonant, after a Hangul syllable), and over every short string of the characters that decide
// composition. The harness and the oracle are case_test.go's, asking normalize instead.

const normalizeHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void put_hex(const adamic_string *string) {
	static const char digits[] = "0123456789abcdef";
	for (size_t index = 0; index < string->length; index++) {
		unsigned char byte = (unsigned char)string->bytes[index];
		putchar(digits[byte >> 4]);
		putchar(digits[byte & 15]);
	}
}

// encode writes a code point as WTF-8: a surrogate as its three bytes, as the runtime stores one.
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

// text is the string of these code points, made the way the runtime makes strings: a lone high
// surrogate met by a lone low one becomes the pair's character, as in JavaScript.
static adamic_string *text(const uint32_t *points, size_t count) {
	adamic_string *parts[8];
	char bytes[8][4];
	adamic_string pieces[8];
	for (size_t index = 0; index < count; index++) {
		size_t size = encode(points[index], bytes[index]);
		adamic_string piece = {{0, adamic_kind_string, 0}, size, bytes[index], 0, NULL, NULL, 0};
		pieces[index] = piece;
		parts[index] = &pieces[index];
	}
	return adamic_string_concat(count, parts);
}

static adamic_string forms[4] = {ADAMIC_STRING("NFC"), ADAMIC_STRING("NFD"), ADAMIC_STRING("NFKC"), ADAMIC_STRING("NFKD")};

static void answer(const uint32_t *points, size_t count) {
	adamic_string *string = text(points, count);
	for (size_t form = 0; form < 4; form++) {
		adamic_string *normalized = adamic_string_normalize(string, &forms[form]);
		putchar(' ');
		put_hex(normalized);
		adamic_release(normalized);
	}
	adamic_release(string);
}

static const uint32_t alphabet[] = {ALPHABET};
#define LETTERS (sizeof alphabet / sizeof alphabet[0])

int main(int count, char **arguments) {
	static char buffer[1 << 20];
	setvbuf(stdout, buffer, _IOFBF, sizeof buffer);
	if (count >= 2 && strcmp(arguments[1], "points") == 0) {
		uint32_t first = count == 4 ? (uint32_t)strtoul(arguments[2], NULL, 10) : 0;
		uint32_t last = count == 4 ? (uint32_t)strtoul(arguments[3], NULL, 10) : 0x110000;
		for (uint32_t point = first; point < last; point++) {
			printf("%x", point);
			answer((uint32_t[]){point}, 1);
			answer((uint32_t[]){'a', point, 0x301}, 3);
			answer((uint32_t[]){point, 0x301, 0x323}, 3);
			answer((uint32_t[]){0x1100, point}, 2);
			answer((uint32_t[]){0xac00, point}, 2);
			putchar('\n');
		}
		return 0;
	}
	// Every string of one to four letters of the alphabet.
	for (size_t length = 1; length <= 4; length++) {
		size_t total = 1;
		for (size_t index = 0; index < length; index++) {
			total *= LETTERS;
		}
		for (size_t number = 0; number < total; number++) {
			uint32_t points[4];
			size_t rest = number;
			for (size_t index = length; index-- > 0;) {
				points[index] = alphabet[rest % LETTERS];
				rest /= LETTERS;
			}
			adamic_string *string = text(points, length);
			put_hex(string);
			adamic_release(string);
			answer(points, length);
			putchar('\n');
		}
	}
	return 0;
}
`

const normalizeOracle = `const hex = [];
for (let byte = 0; byte < 256; byte++) {
	hex.push(byte.toString(16).padStart(2, '0'));
}
// wtf8 is a string's UTF-16 as WTF-8 bytes in hexadecimal: pairs joined, lone surrogates as themselves.
function wtf8(string) {
	let out = '';
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
const character = (point) => (point >= 0xd800 && point <= 0xdfff ? String.fromCharCode(point) : String.fromCodePoint(point));
const answer = (points) => {
	const string = points.map(character).join('');
	return ' ' + ['NFC', 'NFD', 'NFKC', 'NFKD'].map((form) => wtf8(string.normalize(form))).join(' ');
};
const alphabet = [ALPHABET];
let chunk = '';
const flush = () => {
	process.stdout.write(chunk);
	chunk = '';
};
if (process.argv[1] === 'points') {
	const first = process.argv.length === 4 ? Number(process.argv[2]) : 0;
	const last = process.argv.length === 4 ? Number(process.argv[3]) : 0x110000;
	for (let point = first; point < last; point++) {
		chunk += point.toString(16) + answer([point]) + answer([0x61, point, 0x301]) + answer([point, 0x301, 0x323]) +
			answer([0x1100, point]) + answer([0xac00, point]) + '\n';
		if (chunk.length > 1 << 20) {
			flush();
		}
	}
} else {
	for (let length = 1; length <= 4; length++) {
		const total = alphabet.length ** length;
		for (let number = 0; number < total; number++) {
			const points = [];
			let rest = number;
			for (let index = 0; index < length; index++) {
				points.unshift(alphabet[rest % alphabet.length]);
				rest = Math.floor(rest / alphabet.length);
			}
			chunk += wtf8(points.map(character).join('')) + answer(points) + '\n';
			if (chunk.length > 1 << 20) {
				flush();
			}
		}
	}
}
flush();
`

// normalizeAlphabet is starters that compose with marks (a, A), marks of four classes (U+0300 and
// U+0301 and U+0302 at 230, U+0323 at 220, U+0338 at 1, U+0345 at 240), Hangul's leading consonant,
// vowel and trailing consonant and two syllables, two Bengali starters that compose with each other,
// a singleton (Å as the Angstrom sign), decompositions that start with a non-starter (U+0344,
// U+0F73), a composition exclusion beyond the BMP (U+1D15E), compatibility characters (ﬁ, ẛ), and
// both halves of a lone surrogate, which meet as a pair.
var normalizeAlphabet = []string{
	"0x61", "0x41",
	"0x300", "0x301", "0x302", "0x323", "0x338", "0x345",
	"0x1100", "0x1161", "0x11a8", "0xac00", "0xac01",
	"0x9c7", "0x9be",
	"0x212b", "0x344", "0xf73", "0x1d15e",
	"0xfb01", "0x1e9b",
	"0xd800", "0xdc00", "0xc5",
}

func TestNormalizeMatchesNode(t *testing.T) {
	t.Parallel()
	alphabet := strings.Join(normalizeAlphabet, ", ")
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cache, err := newTestBuildCache(repository)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(normalizeHarness, "ALPHABET", alphabet, 1)
	built, err := cache.Tree("normalization-probe", [][]byte{[]byte(source)}, func(destination string) error {
		return Build(source, filepath.Join(destination, "harness"), Options{Sanitize: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(built, "harness")
	oracle := strings.Replace(normalizeOracle, "ALPHABET", alphabet, 1)
	shard := currentTestShard(t)
	pieces := unitRanges(0x110000, normalizePointUnitSize)
	for index, piece := range pieces {
		if !shard.owns(index) {
			continue
		}
		t.Run("points/"+piece.name(), func(t *testing.T) {
			t.Parallel()
			first, last := strconv.Itoa(piece.first), strconv.Itoa(piece.last)
			compareStreams(t, piece.last-piece.first, []string{binary, "points", first, last}, []string{"node", "--eval", oracle, "points", first, last})
		})
	}
	letters := len(normalizeAlphabet)
	if !shard.owns(len(pieces)) {
		return
	}
	t.Run("contexts", func(t *testing.T) {
		t.Parallel()
		compareStreams(t, letters+letters*letters+letters*letters*letters+letters*letters*letters*letters,
			[]string{binary, "contexts"}, []string{"node", "--eval", oracle, "contexts"})
	})
}
