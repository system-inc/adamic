package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// V8's longest string is 536,870,888 UTF-16 units: a string that long is made, and one unit more is
// RangeError: Invalid string length. Building either in a fixture takes a gigabyte, so the edge is
// held here, at the check itself: it says false and leaves the RangeError pending, its name and message
// apart. Nothing catches it here, so it ends as an uncaught error does, exit 70. A second failure keeps
// the first error pending, so none is lost (oracle's TestStringLengthLimitIsNodes holds the number to
// the pinned Node's).
func TestStringLengthLimit(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>

int main(int count, char **arguments) {
	(void)count;
	if (!adamic_string_check_length(strtod(arguments[1], NULL))) {
		adamic_object *first = adamic_thrown;
		bool again = adamic_string_check_length(strtod(arguments[1], NULL));
		const adamic_string *name = adamic_thrown->slots[0].reference;
		const adamic_string *message = adamic_thrown->slots[1].reference;
		fprintf(stderr, "thrown %.*s|%.*s|%s\n", (int)name->length, name->bytes, (int)message->length, message->bytes, !again && adamic_thrown == first ? "kept" : "replaced");
		adamic_uncaught();
	}
	puts("allowed");
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		units  string
		output string
		exit   int
	}{
		{"0", "allowed\n", 0},
		{"536870888", "allowed\n", 0},
		{"536870889", "thrown RangeError|Invalid string length|kept\nadamic: panic: RangeError: Invalid string length\n", 70},
		{"inf", "thrown RangeError|Invalid string length|kept\nadamic: panic: RangeError: Invalid string length\n", 70},
	} {
		output, err := exec.Command(binary, check.units).CombinedOutput()
		exit := 0
		if exitError, isExit := err.(*exec.ExitError); isExit {
			exit = exitError.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if exit != check.exit || !strings.Contains(string(output), check.output) {
			t.Errorf("%s units: exit %d, %q; want exit %d, %q", check.units, exit, output, check.exit, check.output)
		}
	}
}
