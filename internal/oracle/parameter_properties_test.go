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
		"internal/oracle/testdata/parameter_properties.a",
		"internal/oracle/testdata/parameter_properties_ownership.a",
		"internal/oracle/testdata/parameter_properties_levels.a",
		"internal/oracle/testdata/parameter_properties_kinds.a",
		"internal/oracle/testdata/params_namespaces_order.a",
		"internal/oracle/testdata/params_namespaces_values.a",
		"internal/oracle/testdata/params_namespaces_callbacks.a",
		"internal/oracle/testdata/params_namespaces_dotted.a",
		"internal/oracle/testdata/params_namespaces_types.a",
		"internal/oracle/testdata/params_namespaces_overrides.a",
		"internal/oracle/testdata/params_namespaces_shared/main.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestParameterPropertyMutants(test *testing.T) {
	test.Parallel()
	for _, family := range []string{"missing store", "late store"} {
		test.Run(family, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/parameter_properties.a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != "Plain_new" {
					continue
				}
				for position, statement := range function.Body {
					store, ok := statement.(ir.SetProperty)
					if !ok || store.Name != "value" {
						continue
					}
					function.Body = append(function.Body[:position], function.Body[position+1:]...)
					if family == "late store" {
						last := len(function.Body) - 1
						function.Body = append(function.Body[:last], store, function.Body[last])
					}
					changed = true
					break
				}
			}
			if !changed {
				test.Fatal("mutant changed no parameter-property store")
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			result := executeWith(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"},
				binary,
			)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("mutant must finish cleanly: %+v", result)
			}
			if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
				test.Fatalf("mutant not caught by Node: %q", difference)
			}
			test.Log("caught by Node stdout; exit 0 with clean sanitizers")
		})
	}
}

func TestParameterPropertyOwnershipMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/parameter_properties_ownership.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(program)
	retain := ""
	for _, local := range program.Locals {
		if local.Name == "render" && program.Functions[local.Function].Name == "Child_new" {
			retain = "adamic_retain(" + emittedLocalSymbol(test, code, "adamic_local_", "render") + ")"
		}
	}
	if retain == "" || !strings.Contains(code, retain) {
		test.Fatal("no callback field retain")
	}
	code = strings.ReplaceAll(code, retain, strings.TrimSuffix(strings.TrimPrefix(retain, "adamic_retain("), ")"))
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		test.Fatalf("ownership mutant not caught by ASan: exit %d, stderr %s", result.exitCode, result.stderr)
	}
	test.Log("missing callback field retain caught by ASan heap-use-after-free")
}
