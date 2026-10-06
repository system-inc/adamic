package native

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The sweep: native toUpperCase and toLowerCase against Node's, byte for byte, over every code point
// on its own and in the places case mapping looks at its neighbours, and over every short string of
// the characters that decide final sigma. Results are written as their WTF-8 bytes in hexadecimal, so
// a lone surrogate is compared as itself rather than as the U+FFFD stdout would make it.

// caseHarness enumerates the same strings caseOracle does, in the same order, and writes one line for
// each: the string, then what each mapping made of it.
const caseHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
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

static void answer(const uint32_t *points, size_t count, bool upper) {
	adamic_string *string = text(points, count);
	adamic_string *lower = adamic_string_to_lower(string);
	putchar(' ');
	put_hex(lower);
	adamic_release(lower);
	if (upper) {
		adamic_string *upper_case = adamic_string_to_upper(string);
		putchar(' ');
		put_hex(upper_case);
		adamic_release(upper_case);
	}
	adamic_release(string);
}

static const uint32_t alphabet[] = {ALPHABET};
#define LETTERS (sizeof alphabet / sizeof alphabet[0])

int main(int count, char **arguments) {
	static char buffer[1 << 20];
	setvbuf(stdout, buffer, _IOFBF, sizeof buffer);
	if (count == 2 && strcmp(arguments[1], "points") == 0) {
		for (uint32_t point = 0; point <= 0x10ffff; point++) {
			printf("%x", point);
			answer((uint32_t[]){point}, 1, true);
			answer((uint32_t[]){'x', 'Y', point, 'Z', 'w'}, 5, true);
			answer((uint32_t[]){point, 0x3a3}, 2, false);
			answer((uint32_t[]){'A', point, 0x3a3}, 3, false);
			answer((uint32_t[]){'A', 0x3a3, point}, 3, false);
			answer((uint32_t[]){'A', 0x3a3, point, 'a'}, 4, false);
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
			answer(points, length, true);
			putchar('\n');
		}
	}
	return 0;
}
`

const caseOracle = `const hex = [];
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
const answer = (points, upper) => {
	const string = points.map(character).join('');
	return ' ' + wtf8(string.toLowerCase()) + (upper ? ' ' + wtf8(string.toUpperCase()) : '');
};
const alphabet = [ALPHABET];
let chunk = '';
const flush = () => {
	process.stdout.write(chunk);
	chunk = '';
};
if (process.argv[1] === 'points') {
	for (let point = 0; point <= 0x10ffff; point++) {
		chunk += point.toString(16) + answer([point], true) + answer([0x78, 0x59, point, 0x5a, 0x77], true) +
			answer([point, 0x3a3], false) + answer([0x41, point, 0x3a3], false) + answer([0x41, 0x3a3, point], false) +
			answer([0x41, 0x3a3, point, 0x61], false) + '\n';
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
			chunk += wtf8(points.map(character).join('')) + answer(points, true) + '\n';
			if (chunk.length > 1 << 20) {
				flush();
			}
		}
	}
}
flush();
`

// caseAlphabet is the characters final sigma turns on, and the ones whose mappings are special, so
// that every string of up to four of them is every arrangement of those cases: cased (A, a, ǅ, and
// 𐐀 and 𐐨 beyond the BMP), case-ignorable (., U+0301, a soft hyphen, U+E0001 beyond the BMP), both
// at once (ʰ and U+0345, which Node skips as ignorable), neither (space, 1, 😀), lone surrogates of
// each half (which meet as a pair), the sigmas, and ß, İ and ΐ, which map to more than one.
var caseAlphabet = []string{
	"0x41", "0x61", "0x1c5", "0x10400", "0x10428",
	"0x2e", "0x301", "0xad", "0xe0001",
	"0x2b0", "0x345",
	"0x20", "0x31", "0x1f600",
	"0xd800", "0xdc00",
	"0x3a3", "0x3c3", "0x3c2",
	"0xdf", "0x130", "0x390",
}

func TestCaseMappingMatchesNode(t *testing.T) {
	t.Parallel()
	alphabet := strings.Join(caseAlphabet, ", ")
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(strings.Replace(caseHarness, "ALPHABET", alphabet, 1), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	oracle := strings.Replace(caseOracle, "ALPHABET", alphabet, 1)
	for _, sweep := range []struct {
		name  string
		lines int
	}{
		{"points", 0x110000},
		{"contexts", 22 + 22*22 + 22*22*22 + 22*22*22*22},
	} {
		t.Run(sweep.name, func(t *testing.T) {
			t.Parallel()
			compareStreams(t, sweep.lines, []string{binary, sweep.name}, []string{"node", "--eval", oracle, sweep.name})
		})
	}
}

// compareStreams runs two programs and compares what they write, line by line, as it comes, so a
// sweep of millions of lines never has to be held whole.
func compareStreams(t *testing.T, want int, native []string, node []string) {
	t.Helper()
	nativeLines, nativeWait := streamLines(t, native)
	nodeLines, nodeWait := streamLines(t, node)
	lines, mismatches := 0, 0
	for {
		nativeMore, nodeMore := nativeLines.Scan(), nodeLines.Scan()
		if !nativeMore || !nodeMore {
			if nativeMore != nodeMore {
				t.Errorf("after %d lines, only one side went on (native %t, Node %t)", lines, nativeMore, nodeMore)
			}
			break
		}
		lines++
		if nativeLines.Text() != nodeLines.Text() {
			mismatches++
			if mismatches <= 12 {
				t.Errorf("native %q\nNode   %q", nativeLines.Text(), nodeLines.Text())
			}
		}
	}
	for _, scanner := range []*bufio.Scanner{nativeLines, nodeLines} {
		if err := scanner.Err(); err != nil {
			t.Errorf("reading: %v", err)
		}
	}
	nativeWait()
	nodeWait()
	if lines != want {
		t.Errorf("compared %d lines, want %d", lines, want)
	}
	if mismatches > 0 {
		t.Errorf("%d of %d lines differ", mismatches, lines)
	}
	t.Logf("%d lines, %d mismatches", lines, mismatches)
}

func streamLines(t *testing.T, arguments []string) (*bufio.Scanner, func()) {
	t.Helper()
	// The normalization sweep streams every code point through Node and native. Alone it
	// finishes well inside five minutes, but in a full parallel gate it does not.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, arguments[0], arguments[1:]...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	var stderr strings.Builder
	command.Stderr = &stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1<<16), 1<<20)
	return scanner, func() {
		// Drain what's left, so a side that stopped early can't block the other on a full pipe.
		for scanner.Scan() {
		}
		if err := command.Wait(); err != nil {
			t.Errorf("%s: %v\n%s", arguments[0], err, stderr.String())
		}
	}
}

// The tables are Unicode's for the version the Node the oracle runs carries. When Node moves to a new
// Unicode version, this says to run case_generate.go for it, rather than the sweep failing on whatever
// characters changed.
func TestCaseTablesMatchNodesUnicode(t *testing.T) {
	t.Parallel()
	output, err := exec.Command("node", "--print", "process.versions.unicode").Output()
	if err != nil {
		t.Fatal(err)
	}
	node := strings.TrimSpace(string(output))
	for _, tables := range []struct{ file, define, generator string }{
		{"case_tables.h", "CASE_UNICODE_VERSION", "case_generate.go"},
		{"normalize_tables.h", "NORMALIZE_UNICODE_VERSION", "normalize_generate.go"},
	} {
		contents, err := runtime.ReadFile("runtime/" + tables.file)
		if err != nil {
			t.Fatal(err)
		}
		version := regexp.MustCompile(`#define ` + tables.define + ` "([0-9.]+)"`).FindSubmatch(contents)
		if version == nil {
			t.Fatalf("%s names no Unicode version", tables.file)
		}
		// Node says 17.0 for Unicode 17.0.0.
		if !strings.HasPrefix(string(version[1]), node+".") && string(version[1]) != node {
			t.Errorf("%s is Unicode %s and Node is Unicode %s: update %s and run go generate", tables.file, version[1], node, tables.generator)
		}
	}
}
