package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// normalize given a form that isn't one of the four, and String.fromCodePoint given a code that isn't a
// code point, throw Node's RangeError, and a catch takes it, natively and on the JavaScript backend
// (runtime/normalize.c and from_codes.c; ir/string_throws.go): caught where it's thrown, a spread's
// codes among them, through a call, a function value and a finally, and in an async function. Every
// one of these programs finishes, so the leak check holds each to letting go of everything.
func TestLibraryStringThrowsCaught(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"caught", "propagated", "async", "constant"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_string_throws/"+name+".a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatalf("a try around normalize or String.fromCodePoint must lower: %v", err)
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

// One of the four forms written out, the form left out, and code points written out as integers from
// 0 to 0x10FFFF, beside fromCharCode, which never throws: none is a call that can throw, so no function
// is marked as one that can, and the C tests the pending word nowhere.
func TestLibraryStringConstantsDoNotCheck(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_throws/constant.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
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
		test.Error("the C tests the pending word after a call given constants it can't fail on")
	}
}
