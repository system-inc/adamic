package lower

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// These artifacts must compile, exit cleanly under ASan/UBSan/LeakSanitizer,
// and differ only when an independent Node execution supplies the expected output.
func TestLibraryArrayNewFamilyMutants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, fixture string
		mutate        func(*ir.Program) bool
		mutateC       func(string) string
	}{
		{name: "generic forward boundary", fixture: "library_array_search.a", mutate: func(program *ir.Program) bool {
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != "array_generic_indexOf" {
					continue
				}
				for index, statement := range function.Body {
					branch, ok := statement.(ir.If)
					if !ok {
						continue
					}
					eligible, ok := branch.Condition.(ir.Binary)
					if !ok || eligible.Operator != ir.And {
						continue
					}
					bound := eligible.Right.(ir.Binary)
					bound.Operator = ir.LessOrEqual
					eligible.Right = bound
					branch.Condition = eligible
					function.Body[index] = branch
					changed = true
				}
			}
			return changed
		}},
		{name: "generic backward negative sign", fixture: "library_array_search.a", mutate: func(program *ir.Program) bool {
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != "array_generic_lastIndexOf" {
					continue
				}
				for index, statement := range function.Body {
					declaration, ok := statement.(ir.Declare)
					if !ok || program.Locals[declaration.Local].Name != "from" {
						continue
					}
					value := declaration.Value.(ir.Conditional)
					negative := value.WhenTrue.(ir.Binary)
					negative.Operator = ir.Subtract
					value.WhenTrue = negative
					declaration.Value = value
					function.Body[index] = declaration
					changed = true
				}
			}
			return changed
		}},
		{name: "isArray classification", fixture: "library_array_metadata.a", mutate: func(program *ir.Program) bool {
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != "array_is_array" {
					continue
				}
				last := len(function.Body) - 1
				returned := function.Body[last].(ir.Return)
				value := returned.Value.(ir.BooleanConstant)
				value.Value = !value.Value
				returned.Value = value
				function.Body[last] = returned
				changed = true
			}
			return changed
		}},
		{name: "copyWithin undefined end", fixture: "library_array_copy_within.a", mutateC: func(source string) string {
			return strings.ReplaceAll(source, "HUGE_VAL", "0.0")
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(filepath.Join("../oracle/testdata", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if test.name == "copyWithin undefined end" {
				// Exclude pre-existing infinity cases: only the new undefined-end
				// behavior may kill this mutant.
				start := strings.Index(string(data), "console.log([1, 2, 3, 4].copyWithin(0, 2, undefined)")
				if start < 0 {
					t.Fatal("undefined-end fixture segment missing")
				}
				data = data[start:]
			}
			program, err := lowerSource(t, string(data))
			if err != nil {
				t.Fatal(err)
			}
			if test.mutate != nil && !test.mutate(program) {
				t.Fatal("mutant changed nothing")
			}
			source := native.C(program)
			if test.mutateC != nil {
				mutant := test.mutateC(source)
				if mutant == source {
					t.Fatal("C mutant changed nothing")
				}
				source = mutant
			}
			directory := t.TempDir()
			binary := filepath.Join(directory, "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			module := filepath.Join(directory, "truth.mts")
			if err := os.WriteFile(module, data, 0644); err != nil {
				t.Fatal(err)
			}
			run := func(name string, arguments ...string) []byte {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, name, arguments...)
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				if err := command.Run(); err != nil || stderr.Len() != 0 {
					t.Fatalf("failed outside Node comparison: %s: %v, stderr %s", name, err, stderr.Bytes())
				}
				return stdout.Bytes()
			}
			got := run(binary)
			truth := run("node", "--disable-warning=ExperimentalWarning", module)
			if bytes.Equal(got, truth) {
				t.Fatal("Node stdout comparison did not catch the mutant")
			}
			t.Log("caught only by stdout comparison with Node; exit 0, no sanitizer or leak report")
		})
	}
}
