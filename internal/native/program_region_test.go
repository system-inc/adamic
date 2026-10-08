package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

const programRegionHarness = `#include "adamic.h"
#include <stdio.h>
#include <string.h>
static const char *const names[] = {"payload"};
static const bool references[] = {true};
static const adamic_shape shape = {1, names, references, NULL};
int main(int argc, char **argv) {
 adamic_object *member = adamic_object_new(&shape);
 member = adamic_program_adopt_owned(member, sizeof *member + sizeof(adamic_value));
 if (!adamic_program_is(member) || member->heap.references != 0) return 2;
 // The same union pointer alternates between a member and a counted string.
 void *mixed = member;
 adamic_retain(mixed); adamic_release(mixed);
 if (member->heap.references != 0) return 3;
 static adamic_string left = ADAMIC_STRING("counted");
 static adamic_string right = ADAMIC_STRING(" payload");
 mixed = adamic_string_concat(2, (adamic_string *const[]){&left, &right});
 adamic_heap *counted = mixed;
 if (adamic_program_is(counted) || counted->references != 1) return 4;
 adamic_retain(mixed);
 if (counted->references != 2) return 5;
 adamic_release(mixed);
 member->slots[0].reference = mixed; // transfer a counted owner into the member
 adamic_weak *weak = adamic_weak_of(member);
 adamic_release(member); // lifetime ends only at the explicit region boundary
 if (adamic_weak_target(weak) != member) return 6;
 if (argc > 1 && strcmp(argv[1],"shared") == 0) { adamic_share(member); return 7; }
 adamic_program_region_end();
 if (adamic_weak_target(weak) != NULL) return 8;
 adamic_release(weak);
 adamic_program_region_end(); // idempotent; process cleanup also calls it
 puts("safe");
 return 0;
}
`

func TestProgramRegionRuntime(t *testing.T) {
	for _, counted := range []bool{false, true} {
		options := native.Options{ProgramRegion: true, Sanitize: true, Count: counted}
		binary := filepath.Join(t.TempDir(), "region")
		if err := native.Build(programRegionHarness, binary, options); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(binary)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
		output, err := command.CombinedOutput()
		if err != nil || !strings.Contains(string(output), "safe\n") {
			t.Fatalf("%v %s", err, output)
		}
		if counted && !strings.Contains(string(output), "allocations 3 frees 2") {
			t.Fatal(string(output))
		}
		report, err := leakcheck.Check(leakcheck.Program{
			C: programRegionHarness, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
			BuildCounted: func(code, output string) error {
				return native.Build(code, output, native.Options{ProgramRegion: true, Count: true})
			},
			Execute: func(environment []string, name string, args ...string) leakcheck.Run {
				cmd := exec.Command(name, args...)
				cmd.Env = append(os.Environ(), environment...)
				out, err := cmd.CombinedOutput()
				status := 0
				if err != nil {
					if e, ok := err.(*exec.ExitError); ok {
						status = e.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				return leakcheck.Run{Stderr: out, ExitCode: status}
			},
		})
		if err != nil || report != "" {
			t.Fatalf("%v %s", err, report)
		}
		shared := exec.Command(binary, "shared")
		output, err = shared.CombinedOutput()
		if err == nil || !strings.Contains(string(output), "a Program region member cannot cross into parallel work") {
			t.Fatalf("sharing: %v %s", err, output)
		}
		t.Logf("counted=%t: header, counted union, weak invalidation, teardown and sharing backstop passed", counted)
	}
}

func TestProgramRegionRequiresOneShotCLI(t *testing.T) {
	t.Setenv("ADAMIC_PROGRAM_REGION", "")
	for _, options := range []native.Options{{ProgramRegion: true, Target: "wasm32-wasi"}, {ProgramRegion: true, Request: true}} {
		if err := native.ValidateOptions(options); err == nil || err.Error() != "native: Program regions require a one-shot native CLI" {
			t.Fatalf("refusal: %v", err)
		}
	}
	t.Setenv("ADAMIC_PROGRAM_REGION", "1")
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err == nil {
		t.Fatal("environment flag bypassed CLI guard")
	}
}
