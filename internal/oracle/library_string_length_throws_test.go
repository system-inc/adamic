package oracle

import (
	"path/filepath"
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

// A join refuses the next piece before growing its buffer, and this holds it by the buffer's own
// largest capacity, which a counted build records (count.h's join_capacity_peak). join_memory.a joins
// three pieces of 2^28 units. The first fits, so the buffer doubles from 64 to 2^28 bytes. The second
// brings the running length to 2^29, past ADAMIC_STRING_MAX_UNITS (2^29 - 24), so the join throws
// before growing again, and the largest capacity is 2^28 exactly. Dropping array.c's running check
// grows it to 2^29 and then 2^30 before the range error, which still agrees with Node, so only this
// number tells them apart. It's the runtime's own, so a loaded box can't move it, where the resident
// memory this pin once read did: realloc copying or growing in place and huge pages decided that.
func TestStringLengthJoinMemory(t *testing.T) {
	// The capacity is the child program's own counter, so other tests running beside it don't move it.
	t.Parallel()
	holdStringLengthFixtures(t, []string{"join_memory"})
	t.Run("capacity", func(t *testing.T) {
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
		code = "#include <stdio.h>\n" + code[:at] + "\tfprintf(stderr, \"join_capacity_bytes %zu\\n\", atomic_load_explicit(&adamic_counted.join_capacity_peak, memory_order_relaxed));\n" + code[at:]
		binary := filepath.Join(t.TempDir(), "join")
		if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if got.exitCode != 0 || string(got.stdout) != "RangeError: Invalid string length\n" {
			t.Fatalf("exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
		}
		capacity := -1
		for _, line := range strings.Split(string(got.stderr), "\n") {
			if value, found := strings.CutPrefix(line, "join_capacity_bytes "); found {
				if capacity, err = strconv.Atoi(value); err != nil {
					t.Fatalf("join capacity %q: %v", value, err)
				}
			}
		}
		if capacity < 0 {
			t.Fatalf("no join capacity line; stderr %q", got.stderr)
		}
		if capacity != 1<<28 {
			t.Errorf("join buffer reached %d bytes, want %d: the running length check let it grow past the piece that fit", capacity, 1<<28)
		}
	})
}
