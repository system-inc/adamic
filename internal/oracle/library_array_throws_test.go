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
func TestLibraryArrayThrowsCaught(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"caught", "propagated", "async", "constant"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_array_throws/"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("a try around an array length must lower: %v", loweredError)
			}
			node := onNode(test, path)
			if node.exitCode != 0 {
				test.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			sanitized, leaked, comparable := nativelyCheckingLeaks(test, program, true)
			if leaked != "" {
				test.Errorf("leaks:\n%s", leaked)
			}
			if !comparable {
				return
			}
			runs := map[string]run{
				"native":     sanitized,
				"release":    released(test, program),
				"javascript": onJavaScriptBackend(test, program),
			}
			for backend, got := range runs {
				if difference := disagreement(node, got); difference != "" {
					test.Errorf(
						"%s disagrees with Node: %s\nnode: exit %d, stdout %q, stderr %q\n%s: exit %d, stdout %q, stderr %q",
						backend,
						difference,
						node.exitCode,
						node.stdout,
						node.stderr,
						backend,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
		})
	}
}

// A constant length an array can't fail on isn't a call that can throw: no function is marked as one
// that can, and the C tests the pending word nowhere.
func TestLibraryArrayConstantsDoNotCheck(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/library_array_throws/constant.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	for _, function := range program.Functions {
		if function.MayThrow {
			test.Errorf("%s is marked as able to throw", function.Name)
		}
	}
	if program.ClosuresMayThrow {
		test.Error("the program's function values are marked as able to throw")
	}
	if source := native.C(program); strings.Contains(source, "adamic_thrown") {
		test.Error("the C tests the pending word after an array made with a constant length it can't fail on")
	}
}
