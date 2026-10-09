package native

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Keep ordinary runs small. The worker validation uses ADAMIC_STRING_OPERATIONS=100000.
// The strings stay bounded, but reads build indexes between appends and slices so caches can stale.
func TestStringBuildingMatchesNode(t *testing.T) {
	t.Parallel()
	operations := 2000
	if value := os.Getenv("ADAMIC_STRING_OPERATIONS"); value != "" {
		var err error
		operations, err = strconv.Atoi(value)
		if err != nil || operations < 1 {
			t.Fatal("ADAMIC_STRING_OPERATIONS must be a positive integer")
		}
	}
	prefix := strings.Split(indexHarness, "// pattern is")[0]
	harness := prefix + `
static uint32_t state = 19;
static size_t next(size_t limit) {
    state = state * 1664525u + 1013904223u;
    return (state >> 8) % limit;
}
int main(void) {
    const uint32_t points[] = {0x61, 0x7f, 0xe9, 0xff, 0x1f30d, 0xd83c, 0xdf0d, 0xd800, 0xdc00};
    adamic_string *text = adamic_string_concat(0, NULL);
    for (size_t step = 0; step < @OPERATIONS@; step++) {
        size_t operation = next(6);
        double from = (double)next(adamic_string_units(text) + 5) - 2;
        double to = (double)next(adamic_string_units(text) + 5) - 2;
        char bytes[4];
        adamic_string piece = {{0, adamic_kind_string, 0}, encode(points[next(9)], bytes), bytes, 0, NULL, NULL, 0};
        if (operation == 0) {
            text = adamic_string_append(text, 1, (adamic_string *const[]){&piece});
            put_hex(text);
        } else if (operation == 1) {
            adamic_string *slice = adamic_string_slice(text, from, to, true);
            adamic_release(text);
            text = slice;
            put_hex(text);
        } else if (operation == 2) {
            put_number(adamic_string_index_of_from(text, &piece, from));
        } else if (operation == 3) {
            put_number(adamic_string_char_code_at(text, from));
        } else if (operation == 4) {
            adamic_maybe_number point = adamic_string_code_point_at(text, from);
            put_number(point.present ? point.number : -1);
        } else {
            put_number(adamic_string_length(text));
        }
        putchar('\n');
        // Grow long ASCII, Latin, emoji and lone-surrogate texts too: random short slices alone
        // seldom reach the index threshold. Reads precede the append, exercising a live cache.
        if (step % 97 == 0) {
            static adamic_string patterns[] = {ADAMIC_STRING("ascii"), ADAMIC_STRING("\xc3\xa9"), ADAMIC_STRING("\xf0\x9f\x8c\x8d"), ADAMIC_STRING("\xed\xa0\xbc")};
            adamic_string *longer = adamic_string_repeat(&patterns[next(4)], 80);
            adamic_release(text);
            text = longer;
        }
        if (adamic_string_units(text) > 512) {
            adamic_string *slice = adamic_string_slice(text, 0, 200, true);
            adamic_release(text);
            text = slice;
        }
    }
    adamic_release(text);
    return 0;
}
`
	oracle := strings.Split(indexOracle, "const pattern =")[0] + `
let state = 19;
function next(limit) {
    state = (Math.imul(state, 1664525) + 1013904223) >>> 0;
    return (state >>> 8) % limit;
}
const pieces = ['a', '\x7f', 'é', 'ÿ', '🌍', '\ud83c', '\udf0d', '\ud800', '\udc00'];
const patterns = ['ascii', 'é', '🌍', '\ud83c'];
let text = '';
const lines = [];
for (let step = 0; step < @OPERATIONS@; step++) {
    const operation = next(6);
    const from = next(text.length + 5) - 2;
    const to = next(text.length + 5) - 2;
    const piece = pieces[next(9)];
    if (operation === 0) { text += piece; lines.push(wtf8(text)); }
    else if (operation === 1) { text = text.slice(from, to); lines.push(wtf8(text)); }
    else if (operation === 2) lines.push(' ' + text.indexOf(piece, from));
    else if (operation === 3) lines.push(' ' + text.charCodeAt(from));
    else if (operation === 4) lines.push(' ' + (text.codePointAt(from) ?? -1));
    else lines.push(' ' + text.length);
    if (step % 97 === 0) text = patterns[next(4)].repeat(80);
    if (text.length > 512) text = text.slice(0, 200);
}
process.stdout.write(lines.join('\n') + '\n');
`
	replace := func(source string) string {
		return strings.ReplaceAll(source, "@OPERATIONS@", strconv.Itoa(operations))
	}
	binary := filepath.Join(t.TempDir(), "string-build")
	if err := Build(replace(harness), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	want := strings.Split(runWithInput(t, "", "node", "--eval", replace(oracle)), "\n")
	got := strings.Split(runWithInput(t, "", binary), "\n")
	if len(got) != len(want) {
		t.Fatalf("native has %d lines, Node %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("operation %d: native %.200s; Node %.200s", index, got[index], want[index])
		}
	}
	t.Logf("%d stateful operations match Node", operations)
}
