package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A number format given an argument out of its range throws Node's RangeError, and a catch takes it,
// natively and on the JavaScript backend (runtime/number.c, dtoa.c and radix.c; ir/number_format.go):
// caught where it's thrown, through a call, a function value and a finally, and in an async function.
// Every one of these programs finishes, so the leak check holds each to letting go of everything.
func TestLibraryNumberThrowsCaught(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"caught", "propagated", "async", "constant"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_number_throws/"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatalf("a try around a number format must lower: %v", loweredError)
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

// A constant a format can't fail on, or an argument left out, isn't a call that can throw: no function
// is marked as one that can, and the C tests the pending word nowhere.
func TestLibraryNumberConstantsDoNotCheck(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/library_number_throws/constant.a"),
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
		test.Error("the C tests the pending word after a format given a constant it can't fail on")
	}
}
