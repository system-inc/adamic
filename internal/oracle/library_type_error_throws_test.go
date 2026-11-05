package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A write to a frozen object, Object.assign into one, and matchAll or replaceAll given a RegExp
// without the g flag throw Node's TypeError, and a catch takes it, natively and on the JavaScript
// backend (runtime/library_object.c, regexp.c and regexp_replace.c; ir/library_object.go and
// regexp.go): caught where it's thrown, through a call, a function value and a finally, and in an
// async function. Every one of these programs finishes, so the leak check holds each to letting go of
// everything.
func TestLibraryTypeErrorThrowsCaught(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"library_object_throws/caught", "library_object_throws/propagated", "library_object_throws/async", "library_regexp_throws/caught", "library_regexp_throws/propagated", "library_regexp_throws/async", "library_regexp_throws/constant"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a try around a frozen write, matchAll or replaceAll must lower: %v", err)
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

// A literal with the g flag given to matchAll, and an object write in a program that never freezes
// one, isn't a call that can throw: no function is marked as one that can, and the C tests the pending
// word nowhere. replaceAll always can, past V8's longest string (ir/string_throws.go).
func TestLibraryTypeErrorConstantsDoNotCheck(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_regexp_throws/constant.a"))
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
	if program.ObjectsMayFreeze {
		t.Error("a program that never calls Object.freeze is marked as one that may freeze")
	}
	if source := native.C(program); strings.Contains(source, "adamic_thrown") {
		t.Error("the C tests the pending word after matchAll given a global literal or an object write no freeze reaches")
	}
}

// Three writes to a frozen object stop today, each at a NotYet of its own, and must throw Node's
// TypeError once lowered (Language's l03 to l05 on #c3bk0fp): an object destructuring target, an
// array destructuring target and a for-of target. Each fixture's first line names the stop it makes,
// "// a-check: refused <text>", and its second line the promise, which this test holds too. When one
// lowers, its pin goes red here, and its lowering must count FrozenWriteMayThrow and test the pending
// word after the write (checkThrown), as writeProperty does.
func TestFrozenWritePinsStopWithTheirHeaders(t *testing.T) {
	t.Parallel()
	const promise = "// stops today; when lowered, must throw Node's TypeError through FrozenWriteMayThrow."
	for _, name := range []string{"stop_destructure_object", "stop_destructure_array", "stop_for_of_target"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_object_throws", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.SplitN(string(source), "\n", 3)
			want, pinned := strings.CutPrefix(lines[0], "// a-check: refused ")
			if !pinned || want == "" {
				t.Fatalf("first line %q names no stop", lines[0])
			}
			if len(lines) < 2 || lines[1] != promise {
				t.Fatalf("second line isn't the promise %q", promise)
			}
			_, err = lowered(t, absolute)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want an error holding %q", err, want)
			}
		})
	}
}
