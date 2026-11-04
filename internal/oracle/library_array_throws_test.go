package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// new Array(length) given a length that isn't an integer from 0 to 2^32 - 1, and Array.from({ length })
// given one past it, throw Node's RangeError, and a catch takes it, natively and on the JavaScript
// backend (runtime/array.c and array_from.c; ir/array_length.go): caught where it's thrown, through a
// call, a function value and a finally, and in an async function. Every one of these programs
// finishes, so the leak check holds each to letting go of everything.
func TestLibraryArrayThrowsCaught(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"caught", "propagated", "async", "constant"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_throws/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a try around an array length must lower: %v", err)
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

// A constant length an array can't fail on isn't a call that can throw: no function is marked as one
// that can, and the C tests the pending word nowhere.
func TestLibraryArrayConstantsDoNotCheck(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_throws/constant.a"))
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
		t.Error("the C tests the pending word after an array made with a constant length it can't fail on")
	}
}
