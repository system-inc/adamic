package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/enums.a",
		"internal/oracle/testdata/enums_const.a",
		"internal/oracle/testdata/enums_modules/main.a",
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
func TestEnumSemanticMutants(test *testing.T) {
	test.Parallel()
	for _, family := range []string{
		"numeric member",
		"string member",
		"reverse alias",
		"string reverse map",
		"const member",
		"key effects",
	} {
		test.Run(family, func(test *testing.T) {
			test.Parallel()
			fixture := "enums.a"
			if family == "const member" {
				fixture = "enums_const.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
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
				test.Fatal("mutant changed no IR")
			}
			binary := filepath.Join(test.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			result := executeWith(
				test,
				[]string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"},
				binary,
			)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("mutant must finish cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
				test.Fatalf("mutant not caught by Node stdout: %q", difference)
			}
			test.Log("caught by Node stdout; clean exit 0 and sanitizers")
		})
	}
}

func TestEnumCleanupMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(program)
	release := ""
	for _, local := range program.Locals {
		if local.Global && local.Name == "Kind" {
			release = "adamic_release(" + emittedLocalSymbol(test, code, "adamic_global_", "Kind") + ");"
		}
	}
	if release == "" || strings.Count(code, release) != 1 {
		test.Fatal("enum cleanup site not unique")
	}
	code = strings.Replace(code, release, "", 1)
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	// Dead compiler temporaries may still resemble roots in registers or the returned main stack.
	// Keep global roots enabled: the emitted global clearing is still required.
	report, err := leakcheck.Check(leakcheck.Program{
		C: code, Sanitized: binary, Counted: filepath.Join(test.TempDir(), "counted"),
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			environment = append(
				environment,
				"LSAN_OPTIONS=use_stacks=0:use_registers=0",
				"UBSAN_OPTIONS=halt_on_error=1",
			)
			return leakRun(executeWith(test, environment, name, arguments...))
		},
	})
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(report, "LeakSanitizer: detected memory leaks") &&
		!strings.Contains(report, "heap values leaked:") {
		test.Fatalf("enum cleanup mutant not caught: %s", report)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != string(onNode(test, path).stdout) {
		test.Fatal("cleanup mutant must finish cleanly and preserve Node stdout")
	}
	test.Logf("skipped enum cleanup caught only by leak check; Node stdout unchanged: %s", report)
}
