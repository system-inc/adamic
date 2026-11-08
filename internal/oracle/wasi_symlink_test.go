package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Exercise the shared syscall adapter and error formatter directly. The Node
// witness also uses Error identity, whose library work is outside this stack.
func wasiEmptySymlinkAgreesWithNode(test *testing.T, index int) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	path := filepath.Join(root, "internal/oracle/testdata/wasi/symlink_empty.a")
	expected := onNodeWith(test, inputRun{}, path)
	if expected.exitCode != 0 || len(expected.stderr) != 0 {
		test.Fatalf("Node witness: %+v", expected)
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
	if index < 2 {
		for _, target := range []string{"", "wasm32-wasi"}[index : index+1] {
			test.Run("control/"+target, func(test *testing.T) {
				binary := filepath.Join(test.TempDir(), "program")
				if buildError := native.Build(source, binary, native.Options{Target: target}); buildError != nil {
					test.Fatal(buildError)
				}
				var actual run
				if target == "" {
					actual = execute(test, binary)
				} else {
					actual = execute(
						test,
						"node",
						"--disable-warning=ExperimentalWarning",
						filepath.Join(repository, "oracle/wasi.mjs"),
						binary,
					)
				}
				if difference := disagreement(expected, actual); difference != "" {
					test.Fatalf("%s: Node %+v, runtime %+v", difference, expected, actual)
				}
				test.Logf("Node and runtime: stdout %q", actual.stdout)
			})
		}
		return
	}
	mutant := onWasi(test, "#define SYMLINK_MUTANT 1\n"+source)
	if mutant.exitCode != 0 || len(mutant.stderr) != 0 {
		test.Fatalf("mutant must compile and execute cleanly: %+v", mutant)
	}
	if difference := disagreement(expected, mutant); difference != "stdout differs" {
		test.Fatalf("removed-adapter mutant: %q", difference)
	}
	test.Logf("removed-adapter mutant caught by Node comparison: stdout %q", mutant.stdout)
}

func TestWasiEmptySymlinkAgreesWithNode_000(test *testing.T) {
	test.Parallel()
	wasiEmptySymlinkAgreesWithNode(test, 0)
}

func TestWasiEmptySymlinkAgreesWithNode_001(test *testing.T) {
	test.Parallel()
	wasiEmptySymlinkAgreesWithNode(test, 1)
}

func TestWasiEmptySymlinkAgreesWithNode_002(test *testing.T) {
	test.Parallel()
	wasiEmptySymlinkAgreesWithNode(test, 2)
}
