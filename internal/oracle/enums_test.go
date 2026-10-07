package oracle

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/enums.a",
		"internal/oracle/testdata/enums_const.a",
		"internal/oracle/testdata/enums_modules/main.a",
		"internal/oracle/testdata/enums_values.a",
		"internal/oracle/testdata/enums_collections.a",
		"internal/oracle/testdata/enums_switch.a",
		"internal/oracle/testdata/enums_linked/main.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// These mutants must compile and finish without a sanitizer finding. Only independent Node
// output comparison can kill them; a clang diagnostic is not evidence for enum semantics.
func TestEnumSemanticMutants(t *testing.T) {
	for _, family := range []string{"numeric member", "string member", "reverse alias", "string reverse map", "const member", "key effects"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "enums.a"
			if family == "const member" {
				fixture = "enums_const.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for position, statement := range program.Main {
				declaration, ok := statement.(ir.Declare)
				if !ok {
					continue
				}
				literal, ok := declaration.Value.(ir.ObjectLiteral)
				if ok {
					name := program.Locals[declaration.Local].Name
					for index, field := range literal.Fields {
						if family == "numeric member" && name == "Kind" && field.Name == "First" {
							literal.Fields[index].Value = ir.NumberConstant{Value: 99}
							changed = true
						}
						if family == "string member" && name == "Text" && field.Name == "Second" {
							for _, replacement := range literal.Fields {
								if replacement.Name == "First" {
									literal.Fields[index].Value = replacement.Value
									changed = true
								}
							}
						}
						if family == "reverse alias" && name == "Kind" && field.Name == "4" {
							for index, text := range program.Strings {
								if text == "Fourth" {
									literal.Fields[index].Value = ir.StringConstant{Index: index}
									changed = true
								}
							}
						}
					}
					if family == "string reverse map" && name == "Text" {
						literal.Fields = append(literal.Fields, ir.Field{Name: "first", Value: literal.Fields[0].Value})
						changed = true
					}
					declaration.Value = literal
					program.Main[position] = declaration
				}
				if family == "const member" && program.Locals[declaration.Local].Name == "values" {
					array := declaration.Value.(ir.ArrayLiteral)
					array.Elements[0] = ir.NumberConstant{Value: 1}
					declaration.Value = array
					program.Main[position] = declaration
					changed = true
				}
			}
			if family == "key effects" {
				for index := range program.Functions {
					if program.Functions[index].Name == "index" {
						program.Functions[index].Body = program.Functions[index].Body[1:]
						changed = true
					}
				}
			}
			if !changed {
				t.Fatal("mutant changed no IR")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("mutant not caught by Node stdout: %q", difference)
			}
			t.Log("caught by Node stdout; clean exit 0 and sanitizers")
		})
	}
}

func TestEnumCleanupMutant(t *testing.T) {
	t.Parallel()
	// The oracle uses LeakSanitizer on Linux and the leaks tool on macOS.
	if runtime.GOOS != "linux" {
		t.Skip("LeakSanitizer detect_leaks is not supported on this platform")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	release := ""
	for index, local := range program.Locals {
		if local.Global && local.Name == "Kind" {
			release = fmt.Sprintf("adamic_release(adamic_global_%d_Kind);", index)
		}
	}
	if release == "" || strings.Count(code, release) != 1 {
		t.Fatal("enum cleanup site not unique")
	}
	code = strings.Replace(code, release, "", 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	// Dead compiler temporaries may still resemble roots in registers or the returned main stack.
	// Global roots remain enabled, so the emitted global clearing is still required.
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if !strings.Contains(string(result.stderr), "LeakSanitizer: detected memory leaks") {
		t.Fatalf("enum cleanup mutant not caught: exit %d, stderr %s", result.exitCode, result.stderr)
	}
	if string(result.stdout) != string(onNode(t, path).stdout) {
		t.Fatal("cleanup mutant must preserve Node stdout")
	}
	t.Log("skipped enum cleanup caught only by LeakSanitizer; Node stdout unchanged")
}
