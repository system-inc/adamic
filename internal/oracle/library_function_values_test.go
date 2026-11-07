package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A receiver-free library function read as a value into a slot of exactly its type is one function
// value whose body is the direct call (#4q1b2er head 2, lower/library_function_values.go): each program
// in testdata/library_function_values agrees with Node natively, sanitized and on the JavaScript
// backend, and leaks nothing. tsc's timestamp shape, Date.now chosen at run time as a ?: arm and called
// again and again; Date.now read twice and === to itself; and Date.now handed to a parameter and kept
// in a field. Each prints facts about the clock, never the time. The mutants are in
// testdata/library_function_values/mutants.md.
func TestLibraryFunctionValuesAgreeWithNode(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/library_function_values"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(test, directory+"/"+entry.Name())
		})
	}
	if programs != 3 {
		test.Fatalf(
			"%d programs in %s, want 3: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}
