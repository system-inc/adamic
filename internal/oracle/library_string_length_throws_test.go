package oracle

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A string past V8's longest throws Node's RangeError, "Invalid string length", and a catch takes it,
// natively and on the JavaScript backend (runtime/string_build_impl.h's adamic_string_check_length;
// ir/string_throws.go): from a concatenation, a template, a join, repeat, a pad, replaceAll, normalize
// and text = text + more, caught where it's thrown, through a call, a function value, parallelMap's
// work and a finally, and in an async function. Every one of these programs finishes, so the leak
// check holds each to letting go of everything. Each makes a string of 2^28 units, so they run one at
// a time.
func TestLibraryStringLengthThrowsCaught(t *testing.T) {
	t.Parallel()
	holdStringLengthFixtures(t, []string{"caught", "propagated", "async", "constant"})
}

// One pending promise awaited before anything in its statement can run code or throw, Language's c1,
// c7, c9 and d2 (#c3bk0fp): the await consumes the promise, so each lowers, and each agrees with Node.
func TestStringLengthAwaitFirstAgrees(t *testing.T) {
	t.Parallel()
	holdStringLengthFixtures(t, []string{"await_first_callee", "await_first_concat", "await_first_binary", "await_first_declared_rejecting"})
}

// holdStringLengthFixtures holds each named program of library_string_length_throws to Node, natively
// with the leak check, in the release build and on the JavaScript backend, one at a time.
func holdStringLengthFixtures(t *testing.T, names []string) {
	t.Helper()
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_length_throws/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			sanitized, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Errorf("leaks:\n%s", leaked)
			}
			if !comparable {
				return
			}
			runs := map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)}
			for backend, got := range runs {
				if difference := disagreement(node, got); difference != "" {
					t.Errorf("%s disagrees with Node: %s\nnode: exit %d, stdout %q, stderr %q\n%s: exit %d, stdout %q, stderr %q",
						backend, difference, node.exitCode, node.stdout, node.stderr, backend, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// A slice, a trim, a number or a boolean written out, a search and a split make no string longer
// than what they're given, so none is an operation that can throw: no function is marked as one that
// can, and the C tests the pending word nowhere.
func TestLibraryStringLengthConstantsDoNotCheck(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_length_throws/constant.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.MayThrow {
			t.Errorf("%s is marked as able to throw", function.Name)
		}
	}
	if program.ClosuresMayThrow {
		t.Error("the program's function values are marked as able to throw")
	}
	if source := native.C(program); strings.Contains(source, "adamic_thrown") {
		t.Error("the C tests the pending word after an operation that can't make a longer string")
	}
}

// The string-length limit is V8's to the unit, measured on the pinned Node and never typed here:
// 'a'.repeat(n) makes a string at n and throws RangeError "Invalid string length" at n + 1, and
// ADAMIC_STRING_MAX_UNITS, as the runtime compiles it, is that n. Adamic's ruling on #c3bk0fp, that
// the limit throws a RangeError a catch takes, rests on the two being equal; if the pinned Node's limit
// moves, this is where it shows. The search makes strings of up to n one-byte units, about 512 MB.
func TestStringLengthLimitIsNodes(t *testing.T) {
	t.Parallel()
	const search = `
const fits = (units) => {
	try {
		return 'a'.repeat(units).length === units;
	} catch (error) {
		if (error instanceof RangeError && error.message === 'Invalid string length') {
			return false;
		}
		throw error;
	}
};
// fits(low) and not fits(high), from the first to the last step.
let low = 0;
let high = 2 ** 32;
if (fits(high)) {
	throw new Error('no limit below 2^32 units');
}
while (high - low > 1) {
	const middle = Math.floor((low + high) / 2);
	if (fits(middle)) {
		low = middle;
	} else {
		high = middle;
	}
}
console.log(low);
`
	measured := execute(t, "node", "--eval", search)
	if measured.exitCode != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", measured.exitCode, measured.stdout, measured.stderr)
	}
	nodeLimit, err := strconv.Atoi(strings.TrimSpace(string(measured.stdout)))
	if err != nil {
		t.Fatalf("Node's limit: %v", err)
	}
	const harness = `#include "adamic.h"
#include <stdio.h>

int main(void) {
	printf("%lld\n", (long long)ADAMIC_STRING_MAX_UNITS);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "limit")
	if err := native.Build(harness, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	compiled := execute(t, binary)
	if compiled.exitCode != 0 {
		t.Fatalf("the limit's harness: exit %d, stderr %q", compiled.exitCode, compiled.stderr)
	}
	adamicLimit, err := strconv.Atoi(strings.TrimSpace(string(compiled.stdout)))
	if err != nil {
		t.Fatalf("ADAMIC_STRING_MAX_UNITS: %v", err)
	}
	t.Logf("V8's longest string, measured: %d UTF-16 units; ADAMIC_STRING_MAX_UNITS: %d", nodeLimit, adamicLimit)
	if adamicLimit != nodeLimit {
		t.Errorf("ADAMIC_STRING_MAX_UNITS is %d, but the pinned Node makes a string of %d units and refuses one more", adamicLimit, nodeLimit)
	}
}

// A join refuses the next piece before growing its buffer. The release build's peak includes the
// original piece and the first copied piece; copying all three exceeds this bound even though the
// eventual range error still agrees with Node. Dropping array.c's running check reds peak.
func TestStringLengthJoinMemory(t *testing.T) {
	// The peak is the child program's own getrusage, so other tests running beside it don't move it.
	t.Parallel()
	holdStringLengthFixtures(t, []string{"join_memory"})
	t.Run("peak", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("peak memory measurement needs Linux")
		}
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_length_throws/join_memory.a"))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		code := native.C(program)
		at := strings.LastIndex(code, "\treturn adamic_host_exit_status();")
		if at < 0 {
			t.Fatal("main return missing")
		}
		code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage usage; getrusage(RUSAGE_SELF, &usage); fprintf(stderr, \"join_peak_kib %ld\\n\", usage.ru_maxrss);\n" + code[at:]
		binary := filepath.Join(t.TempDir(), "join")
		if err := native.Build(code, binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		var peak int64
		if _, err := fmt.Sscanf(string(got.stderr), "join_peak_kib %d", &peak); err != nil {
			t.Fatalf("peak memory: %v; stderr %q", err, got.stderr)
		}
		if got.exitCode != 0 || string(got.stdout) != "RangeError: Invalid string length\n" {
			t.Fatalf("exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
		}
		t.Logf("join peak: %d KiB", peak)
		// Two pieces need about 512 MiB; three copied pieces plus the original need about 1 GiB.
		if peak > 750*1024 {
			t.Errorf("join peak %d KiB exceeds 750 MiB", peak)
		}
	})
}
