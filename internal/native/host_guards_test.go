package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hostGuardsHarness calls one of the runtime's readers by name directly on a Stats-shaped host object,
// its shape on adamic_host_methods as node_fs_file.c's stat_shape is, so a guard no program reaches
// first is still held to its stop. Given "read", it's a read an in test would have narrowed
// (adamic_dynamic_property); given "own", ToPrimitive's own-member test (adamic_object_own_function).
// Each must stop before a slot is read, in the runtime value form naming the key.
const hostGuardsHarness = `#include "adamic.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
static const char *const names[] = {"size", "_fsFileMode"};
static const bool references[] = {false, false};
static const adamic_field_kind kinds[] = {adamic_field_number, adamic_field_number};
static const adamic_shape stats_shape = {2, names, references, &adamic_host_methods, kinds};
int main(int count, char **arguments) {
 assert(count == 3);
 adamic_stack_thread_start();
 adamic_object *stats = adamic_object_new(&stats_shape);
 stats->slots[0].number = 4096;
 stats->slots[1].number = 16877;
 if (strcmp(arguments[1], "read") == 0) {
  adamic_heap *value = adamic_dynamic_property(&stats->heap, arguments[2]);
  printf("read %s\n", value == NULL ? "undefined" : "a value");
 } else {
  printf("own %s\n", adamic_object_own_function(&stats->heap, arguments[2]) ? "true" : "false");
 }
 return 0;
}
`

// The two guards that no reflect_get fixture reaches first, held directly (Runtime's should on
// #ey7aysy, 11:41Z): a program's in test stops before the read it narrows, and ToPrimitive's own
// toString lookup is backed by adamic_object_has's stop, so each fixture stops in the same words with
// either guard gone. Mutants, run by hand: host_stop dropped from adamic_dynamic_property, and the
// read row answers; adamic_host_stop dropped from adamic_object_own_function, and the own row does.
func TestHostGuardsStopOnAStatsShape(test *testing.T) {
	test.Parallel()
	binary := filepath.Join(test.TempDir(), "host-guards")
	if err := Build(hostGuardsHarness, binary, Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	for _, probe := range []struct{ reader, key string }{
		{"read", "_fsFileMode"},
		{"read", "size"},
		{"own", "toString"},
		{"own", "valueOf"},
	} {
		test.Run(probe.reader+"/"+probe.key, func(test *testing.T) {
			test.Parallel()
			command := exec.Command(binary, probe.reader, probe.key)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			exit, ok := err.(*exec.ExitError)
			want := "adamic: panic: this runtime value's properties aren't read by a run-time key yet (reading '" +
				probe.key +
				"')\n"
			if !ok ||
				exit.ExitCode() != 70 ||
				stdout.Len() != 0 ||
				stderr.String() != want ||
				strings.Contains(stderr.String(), "Sanitizer") {
				test.Fatalf(
					"want the host stop and exit 70: %v\nstdout %q\nstderr %q",
					err,
					stdout.String(),
					stderr.String(),
				)
			}
		})
	}
}
