package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// p2b predates lowering symlinkSync. Exercise its shared syscall adapter and
// existing error formatter directly until the combined host landing arrives.
// The source fixture remains Node's independent authority.
func TestWASIEmptySymlinkAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal/oracle/testdata/wasi/symlink_empty.a")
	expected := onNode(t, path)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		t.Fatalf("Node witness: %+v", expected)
	}
	source := fmt.Sprintf(`#define _XOPEN_SOURCE 700
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <errno.h>
#include <dirent.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>
#include %q
#ifdef SYMLINK_MUTANT
#undef symlink
#endif
int main(void) {
    adamic_start(0, NULL);
    if (symlink("", "adamic-empty-symlink-link") == 0) { return 23; }
    int error = errno;
    adamic_string target = ADAMIC_STRING("");
    adamic_node_fs_raise(&target, error, "symlink");
    adamic_object *thrown = adamic_thrown;
    adamic_string *code = thrown->slots[2].reference;
    adamic_string *message = thrown->slots[1].reference;
    // WASI libc errno constants differ from Linux. Translate the two tested
    // symbolic errors to Node/libuv's Linux error numbers for comparison.
    int node_errno = error == ENOENT ? -2 : error == EINVAL ? -22 : 0;
    printf("%%.*s\n%%d\n%%.*s -> 'adamic-empty-symlink-link'\n",
           (int)code->length, code->bytes, node_errno,
           (int)message->length, message->bytes);
    adamic_thrown = NULL;
    adamic_release(thrown);
    return 0;
}
`, filepath.Join(root, "internal/native/runtime/node_fs_wasi.h"))
	for _, target := range []string{"", "wasm32-wasi"} {
		t.Run("control/"+target, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(source, binary, native.Options{Target: target}); err != nil {
				t.Fatal(err)
			}
			var actual run
			if target == "" {
				actual = execute(t, binary)
			} else {
				actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
			}
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("%s: Node %+v, runtime %+v", difference, expected, actual)
			}
			t.Logf("Node and runtime: stdout %q", actual.stdout)
		})
	}
	mutant := onWASI(t, "#define SYMLINK_MUTANT 1\n"+source)
	if mutant.exitCode != 0 || len(mutant.stderr) != 0 {
		t.Fatalf("mutant must compile and execute cleanly: %+v", mutant)
	}
	if difference := disagreement(expected, mutant); difference != "stdout differs" {
		t.Fatalf("removed-adapter mutant: %q", difference)
	}
	t.Logf("removed-adapter mutant caught by Node comparison: stdout %q", mutant.stdout)
}
