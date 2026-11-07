package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/native-generic-empty-array.a",
		"internal/oracle/testdata/generic_empty_array.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestEmptyArrayRuntimeMutants(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/generic_empty_array.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, mutation := range []string{"empty array is falsy", "fresh number array is shared"} {
		test.Run(mutation, func(test *testing.T) {
			test.Parallel()
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			shared := -1
			for index, local := range program.Locals {
				if local.Global && local.Name == "emptyArray" {
					shared = index
				}
			}
			if shared < 0 {
				test.Fatal("shared array absent")
			}
			changed := 0
			for index := range program.Functions {
				function := &program.Functions[index]
				if !strings.HasPrefix(function.Name, "finish_") && !strings.HasPrefix(function.Name, "fresh_") {
					continue
				}
				returned := function.Body[0].(ir.Return)
				coalesce := returned.Value.(ir.Coalesce)
				if mutation == "empty array is falsy" {
					read, ok := coalesce.Value.(ir.Read)
					if !ok {
						test.Fatalf("want parameter read, got %#v", coalesce.Value)
					}
					missing := ir.Undefined{Of: ir.Array}
					coalesce.Value = ir.Conditional{
						Condition: ir.IsUndefined{Value: read}, WhenTrue: missing,
						WhenNot: ir.Conditional{
							Condition: ir.Binary{
								Operator: ir.Equal,
								Left:     ir.Length{Array: read},
								Right:    ir.NumberConstant{},
							},
							WhenTrue: missing, WhenNot: read, Of: ir.Array,
						}, Of: ir.Array,
					}
				} else {
					literal, ok := coalesce.Fallback.(ir.ArrayLiteral)
					if !ok || literal.Element != ir.Number {
						continue
					}
					coalesce.Fallback = ir.Read{Local: shared, Of: ir.Array}
				}
				returned.Value = coalesce
				function.Body[0] = returned
				changed++
			}
			if changed == 0 {
				test.Fatal("mutant target absent")
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			result := execute(test, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("mutant did not finish cleanly: %d %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
				test.Fatalf("want stdout differs, got %q", difference)
			}
			if difference := disagreement(
				onNode(test, path),
				onJavaScriptBackend(test, program),
			); difference != "stdout differs" {
				test.Fatalf("want JavaScript stdout differs, got %q", difference)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
			test.Logf("Node caught %s: %q", mutation, result.stdout)
		})
	}
}
