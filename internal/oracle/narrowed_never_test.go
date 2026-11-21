package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A value typed T, T's constraint holding undefined or null, narrowed past === undefined or === null, at an
// instance whose T holds nothing else (#rnbvya6): the checker narrows T by its constraint and calls the
// read a string, but here it's never, and lowering's defined (narrowed.go) holds what follows its check as
// the narrowed type. Each program in testdata/narrowed_never agrees with Node natively, sanitized and on the
// JavaScript backend, and leaks nothing (generatorAgrees). Before, return.ts's bare return handed clang the
// missing object the declaration is held as, in a function returning a string, and clang refused it. The
// mutant is in testdata/narrowed_never/mutants.md.
func TestNarrowedNeverReadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/narrowed_never"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 5 {
		t.Fatalf("%d programs in %s, want 5: a program added or removed needs this count changed with it", programs, directory)
	}
}

// What follows the check in a T made only of undefined is a loud trap, never a zero value (Compiler's should
// on #rnbvya6, rulings.md:74): return.ts's lowering holds a never_read function that panics, naming the
// read. With unreachablePlaceholder's zero value back in its place, the programs still agree with Node, so
// only this test goes red.
func TestNarrowedNeverReadIsATrap(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/narrowed_never/return.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range program.Functions {
		if function.Name != "never_read" || len(function.Body) != 1 {
			continue
		}
		if panics, ok := function.Body[0].(ir.Panic); ok {
			if message, ok := panics.Message.(ir.StringConstant); ok && strings.Contains(program.Strings[message.Index], "unreachable read of v at ") {
				return
			}
		}
	}
	t.Fatal("return.ts's never read past its check lowers to no never_read trap naming v")
}
