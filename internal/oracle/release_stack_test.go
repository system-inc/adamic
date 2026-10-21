package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestStackReservationMeasuresArgv(t *testing.T)        { t.Parallel(); reservationProbe(t, true) }
func TestStackReservationMeasuresEnvironment(t *testing.T) { t.Parallel(); reservationProbe(t, false) }

// Linux cannot exec macOS's above-quarter argv layouts. A deterministic constructor probe
// exercises that reservation with real strings and a controlled initial frame and soft limit.
func reservationProbe(t *testing.T, argv bool) {
	t.Helper()
	runtimeDirectory, err := filepath.Abs(filepath.Join(repository, "internal/native/runtime"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(runtimeDirectory, "stack.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []bool{false, true} {
		body := string(source)
		if mutant {
			body = strings.Replace(body, "arguments = measured;", "arguments = size / 4;", 1)
		}
		list := "NULL, values"
		if argv {
			list = "values, NULL"
		}
		harness := `#include <sys/resource.h>
#include <stdint.h>
#include <stdlib.h>
static char storage[2 << 20];
static int stack_rlimit(int resource, struct rlimit *limit) { (void)resource; limit->rlim_cur = 1 << 20; return 0; }
#define getrlimit stack_rlimit
#define __builtin_frame_address(depth) ((void *)storage)
#define __attribute__(attributes)
` + body + fmt.Sprintf(`
_Noreturn void adamic_panic(const char *message, size_t length) { (void)message; (void)length; exit(70); }
int main(void) {
 char *values[] = {storage + (600 << 10), NULL};
 find_stack_limit(1, %s);
 uintptr_t expected = (uintptr_t)storage - (1 << 20) + (600 << 10) + 1 + ABOVE_STRINGS + (128 << 10);
 return adamic_stack_limit == expected ? 0 : 1;
}
`, list)
		directory := t.TempDir()
		path := filepath.Join(directory, "probe.c")
		binary := filepath.Join(directory, "probe")
		if err := os.WriteFile(path, []byte(harness), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, command := bounded(t, "clang", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-I", runtimeDirectory, path, "-o", binary)
		command.Env = toolchain.Env()
		if stdout, stderr, err := childguard.Capture(ctx, command); err != nil {
			t.Fatalf("compile mutant=%v: %v\n%s%s", mutant, err, stdout, stderr)
		}
		got := execute(t, binary)
		if !mutant && got.exitCode != 0 {
			t.Fatalf("measured reservation missing: %+v", got)
		}
		if mutant && got.exitCode != 1 {
			t.Fatalf("quarter-only mutant escaped: %+v", got)
		}
		if mutant {
			t.Log("quarter-only reservation mutant caught: exit 1")
		}
	}
}
