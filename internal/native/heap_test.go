package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The size-class allocator (heap.c) must not hide a use after a free from the address sanitizer.
// Sanitized builds take every value from malloc, so ASan sees it as it always has; with the classes
// kept on (ADAMIC_SLABS), a freed slot is poisoned, the first word too, where a retain or release
// would touch it. Each use here is made on a string the runtime made and freed, through the runtime's
// own release, and each is caught; and the same program that doesn't touch the freed string runs
// clean, so the harness can pass.
func TestFreedValuesAreCaughtWithSlabs(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <string.h>

int main(int count, char **arguments) {
	(void)count;
	static adamic_string left = ADAMIC_STRING("built ");
	static adamic_string right = ADAMIC_STRING("at runtime");
	adamic_string *text = adamic_string_concat(2, (adamic_string *const[]){&left, &right});
	adamic_release(text);
	if (strcmp(arguments[1], "retain") == 0) {
		adamic_retain(text);
	} else if (strcmp(arguments[1], "length") == 0) {
		return (int)text->length;
	}
	return 0;
}
`
	for _, build := range []struct {
		name   string
		slabs  bool
		report string
	}{
		{"malloc", false, "heap-use-after-free"},
		{"slabs", true, "use-after-poison"},
	} {
		binary := filepath.Join(t.TempDir(), "harness-"+build.name)
		if err := Build(harness, binary, Options{Sanitize: true, slabs: build.slabs}); err != nil {
			t.Fatal(err)
		}
		for _, use := range []string{"retain", "length", "none"} {
			output, err := exec.Command(binary, use).CombinedOutput()
			caught := strings.Contains(string(output), "ERROR: AddressSanitizer: "+build.report)
			switch {
			case use == "none" && err != nil:
				t.Errorf("%s, untouched: want a clean run, got %v\n%s", build.name, err, output)
			case use != "none" && (err == nil || !caught):
				t.Errorf("%s, %s after the free: want AddressSanitizer's %s, got %v\n%.400s", build.name, use, build.report, err, output)
			}
		}
	}
}
