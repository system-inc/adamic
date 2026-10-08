package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
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

func TestParameterPropertyMutants(t *testing.T) {
	for _, family := range []string{"missing store", "late store"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/parameter_properties.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
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
				t.Fatal("mutant changed no parameter-property store")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %+v", result)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("mutant not caught by Node: %q", difference)
			}
			t.Log("caught by Node stdout; exit 0 with clean sanitizers")
		})
	}
}

func TestParameterPropertyOwnershipMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/parameter_properties_ownership.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	retain := ""
	for index, local := range program.Locals {
		if local.Name == "render" && program.Functions[local.Function].Name == "Child_new" {
			retain = fmt.Sprintf("adamic_retain(adamic_local_%d_render)", index)
		}
	}
	if retain == "" || !strings.Contains(code, retain) {
		t.Fatal("no callback field retain")
	}
	local := strings.TrimSuffix(strings.TrimPrefix(retain, "adamic_retain("), ")")
	// Retains can be standalone statements after emitter changes. Erase their
	// effect without letting -Werror count as an ownership mutant kill.
	code = regexp.MustCompile(`(?m)^([ \t]*)`+regexp.QuoteMeta(retain)+`;[ \t]*$`).ReplaceAllString(code, "${1}(void)"+local+";")
	code = strings.ReplaceAll(code, retain, local)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("ownership mutant not caught by ASan: exit %d, stderr %s", result.exitCode, result.stderr)
	}
	t.Log("missing callback field retain caught by ASan heap-use-after-free")
}
