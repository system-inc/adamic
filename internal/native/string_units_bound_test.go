package native

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// A string's index trusts its units only as far as it counts them (string_index.c's build): the
// checkpoint walk counts every unit of the bytes, and a string whose units say otherwise stops by name
// before its UTF-16 view is allocated, in every build, rather than decoding past the end of a view
// sized from the wrong count. No producer sets units wrong today, so the harness sets them itself:
// right, where the first indexed read is through the view, and one short, where it stops. The mutant
// is the shipping runtime with the check removed, where the short row writes past the view, as
// Library's mutant did (J-rjgc4x6, #11np4d1).

// stringUnitsHarness makes a heap string long enough for an index, of one-, two-, three- and
// four-byte code points, sets its units right or one short as its argument says, and reads it.
const stringUnitsHarness = `#include "adamic.h"

#include <stdio.h>
#include <string.h>

int main(int count, char **arguments) {
	adamic_start(count, arguments);
	// a, é and € are a unit each, and the emoji a pair: five units in ten bytes.
	static const char group[] = "a\xc3\xa9\xe2\x82\xac\xf0\x9f\x98\x80";
	size_t size = sizeof group - 1, groups = 20, units = 5 * groups;
	adamic_string *text = adamic_string_allocate(size * groups);
	for (size_t index = 0; index < groups; index++) {
		memcpy((char *)text->bytes + index * size, group, size);
	}
	bool short_by_one = count > 1 && strcmp(arguments[1], "short") == 0;
	text->units = short_by_one ? units : units + 1;
	printf("%.0f %.0f\n", adamic_string_char_code_at(text, 2), adamic_string_char_code_at(text, (double)(units - 1)));
	adamic_release(text);
	return 0;
}
`

// stringUnitsCheck is build's check, the one line the mutant removes.
const stringUnitsCheck = `if (unit != units) { adamic_invariant("a string's unit count is wrong"); }`

// stringUnitsStop is all the short row writes to stderr: the runtime's own stop, and no sanitizer report.
const stringUnitsStop = "adamic: panic: runtime invariant broken: a string's unit count is wrong\n"

func TestStringUnitsRightReadThroughTheView(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "units")
	if err := Build(stringUnitsHarness, binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	// € is U+20AC, and the last unit is the low half of U+1F600's pair.
	result := inlineArrayRun(t, binary, "right")
	if result.exit != 0 || string(result.stdout) != "8364 56832\n" || len(result.stderr) != 0 {
		t.Fatalf("units right: exit %d stdout %q stderr %s", result.exit, result.stdout, result.stderr)
	}
	result = inlineArrayRun(t, binary, "short")
	if result.exit != 70 || len(result.stdout) != 0 || string(result.stderr) != stringUnitsStop {
		t.Fatalf("units one short: exit %d stdout %q stderr %s", result.exit, result.stdout, result.stderr)
	}
	t.Logf("right: %q; one short: exit %d, %q", result.stdout, result.exit, result.stderr)
}

func TestStringUnitsOneShortMutantOverflowsTheView(t *testing.T) {
	t.Parallel()
	options := native_runtime.Options{Sanitize: true}
	files, flags, compiler, version, err := native_runtime.RuntimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	for index := range files {
		if files[index].Name != "string_index.c" {
			continue
		}
		source := string(files[index].Contents)
		if strings.Count(source, stringUnitsCheck) != 1 {
			t.Fatal("string_index.c must hold build's unit check exactly once")
		}
		files[index].Contents = []byte(strings.Replace(source, stringUnitsCheck, "/* unit check removed */", 1))
		mutated = true
	}
	if !mutated {
		t.Fatal("no string_index.c in the runtime")
	}
	cache, err := native_runtime.CacheDirectory("string-units-mutant")
	if err != nil {
		t.Fatal(err)
	}
	library, err := native_runtime.CachedRuntime(files, flags, compiler, version, cache)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := BuildWithRuntime(stringUnitsHarness, binary, options, library); err != nil {
		t.Fatal(err)
	}
	result := inlineArrayRun(t, binary, "short")
	stderr := string(result.stderr)
	if result.exit == 0 || result.exit == 70 || strings.Contains(stderr, "runtime invariant broken") ||
		!strings.Contains(stderr, "AddressSanitizer: heap-buffer-overflow") || !strings.Contains(stderr, "WRITE of size 2") {
		t.Fatalf("mutant survived or failed for another reason: exit %d stdout %q stderr %s", result.exit, result.stdout, stderr)
	}
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "AddressSanitizer: heap-buffer-overflow") || strings.Contains(line, "WRITE of size 2") || strings.Contains(line, "string_index.c") {
			t.Log(line)
		}
	}
}
