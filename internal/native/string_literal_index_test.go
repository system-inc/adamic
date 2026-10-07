package native

import (
	"path/filepath"
	"strings"
	"testing"
)

// A long non-ASCII constant builds its position index the first time it's read by position, and keeps
// it for the program's life (string_index.c, usable: a constant's index starts as ADAMIC_LITERAL_INDEX,
// not NULL, so it may be built though its count is 0). Without that, every charCodeAt on a constant
// walks from its start, which is what took a scan of one from 7.2 s to 12 ms. Every answer is the same
// either way, so only the index itself can show it: integration's reading of 4ddd17f (#fxspptb) found
// nothing failing when the constant case was turned off.
func TestLiteralIndexIsBuiltOnceAndKept(t *testing.T) {
	t.Parallel()
	text := strings.Repeat("é😀A", 40) + "Z"
	source := `#include "adamic.h"
#include <stdio.h>
static adamic_string literal = ADAMIC_STRING(@TEXT@);
int main(void) {
    const char *before = literal.index == ADAMIC_LITERAL_INDEX ? "unbuilt" : "other";
    double last = adamic_string_char_code(&literal, adamic_string_length(&literal) - 1);
    struct adamic_string_index *built = literal.index;
    const char *after = built == NULL ? "null" : built == ADAMIC_LITERAL_INDEX ? "unbuilt" : "built";
    double middle = adamic_string_char_code(&literal, 100);
    printf("%s %s %s %.0f %.0f\n", before, after, literal.index == built ? "kept" : "rebuilt", last, middle);
    return 0;
}`
	binary := filepath.Join(t.TempDir(), "literal-index")
	if err := Build(strings.ReplaceAll(source, "@TEXT@", cString(text))+"\n", binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	node := runWithInput(t, "", "node", "--eval", `const text = "é😀A".repeat(40) + "Z"; console.log("unbuilt built kept", text.charCodeAt(text.length - 1), text.charCodeAt(100));`)
	got := runWithInput(t, "", binary)
	if got != node {
		t.Fatalf("native %q, want %q", got, node)
	}
}
