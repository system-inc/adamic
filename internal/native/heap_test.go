package native

import (
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
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
		if err := Build(harness, binary, Options{Sanitize: true, Slabs: build.slabs}); err != nil {
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

// The size classes share their chunks: memory one class emptied is carved by the next, so a program
// that builds its values at one size, then another, peaks at what it holds at once (reviewer R2's
// finding, review/r2/alloc/, was five times main's peak when each class kept its own). The harness
// builds generations of strings through the runtime at sizes across every class, down and up again,
// and a control builds the same generations at one size; both are release builds, where the classes
// are on, and the churn's peak resident set must stay near the control's.
func TestSizeClassesShareTheirChunks(t *testing.T) {
	t.Parallel()
	const harness = `#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>

static adamic_string *items[20000];

int main(int count, char **arguments) {
	(void)count;
	bool churn = strcmp(arguments[1], "churn") == 0;
	static const size_t lengths[] = {8, 40, 72, 104, 136, 168, 192, 168, 136, 104, 72, 40, 8, 192, 8};
	for (size_t round = 0; round < sizeof lengths / sizeof lengths[0]; round++) {
		size_t length = churn ? lengths[round] : 192;
		// The control must use a slab too: 200 bytes plus the string header exceeds 256.
		if (sizeof(adamic_string) + length > 256) { return 2; }
		for (size_t index = 0; index < 20000; index++) {
			items[index] = adamic_string_allocate(length);
			memset((char *)items[index]->bytes, 'a' + (int)(index % 26), length);
		}
		for (size_t index = 0; index < 20000; index++) {
			if (items[index]->bytes[length - 1] != 'a' + (int)(index % 26)) {
				puts("corrupt");
				return 1;
			}
			adamic_release(items[index]);
		}
	}
	struct rusage usage;
	getrusage(RUSAGE_SELF, &usage);
	printf("%ld\n", usage.ru_maxrss);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(harness, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	peak := func(mode string) int {
		// Fork from a small shell after exec, so the child's high-water mark doesn't inherit
		// the Go test process's changing resident set while other tests run in parallel.
		output, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary, mode).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", mode, err, output)
		}
		kilobytes, err := strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			t.Fatalf("%s: %q", mode, output)
		}
		return residentKibibytes(kilobytes, goruntime.GOOS)
	}
	control, churn := peak("control"), peak("churn")
	// Fifteen generations at up to fourteen sizes: kept apart, they'd peak several times the control.
	if churn > control*3/2 {
		t.Errorf("the churn through every class peaked at %d KiB, the same churn at one size at %d KiB: the classes aren't sharing their chunks", churn, control)
	}
	t.Logf("peak %d KiB through every class, %d KiB at one size", churn, control)
}

// getrusage reports bytes on macOS and KiB on Linux. Normalize before reporting or comparing.
func residentKibibytes(value int, operatingSystem string) int {
	if operatingSystem == "darwin" {
		return value / 1024
	}
	return value
}

func TestResidentSetUnits(t *testing.T) {
	t.Parallel()
	for _, given := range []struct {
		operatingSystem string
		value           int
	}{
		{"linux", 8192}, {"darwin", 8192 * 1024},
	} {
		if got := residentKibibytes(given.value, given.operatingSystem); got != 8192 {
			t.Errorf("%s: got %d KiB, want 8192", given.operatingSystem, got)
		}
	}
}
