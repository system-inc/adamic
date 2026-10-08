package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// This observes the actual input/output storage addresses, rather than trusting
// the planner or counting its emitted calls. The ordinary counted control must
// take the in-place mapper path; the same allocation with Program membership
// must allocate fresh even though the planner permits the consumed source.
func TestProgramRegionMapperStorage(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/program_region_map_storage.a"))
	oracle := onNode(t, path)
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("program=%t", enabled), func(t *testing.T) {
			code := native.C(programRegionLowered(t, path, enabled))
			allocation := regexp.MustCompile(`adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_new\([^\n]+;`)
			matches := allocation.FindAllStringSubmatchIndex(code, -1)
			if len(matches) != 1 {
				t.Fatalf("expected exactly one reusable mapper allocation, got %d", len(matches))
			}
			m := matches[0]
			result, unique, source := code[m[2]:m[3]], code[m[4]:m[5]], code[m[6]:m[7]]
			condition, message := result+" != "+source, "counted mapper control did not reuse"
			if enabled {
				condition, message = result+" == "+source, "Program member array storage was reused"
			}
			assertion := fmt.Sprintf("\n if (adamic_program_is(%s) != %t) adamic_panic(\"mapper fixture has wrong membership\", sizeof \"mapper fixture has wrong membership\" - 1);\n if (%s) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n", source, enabled, condition, message, message)
			checked := code[:m[1]] + assertion + code[m[1]:]
			binary := filepath.Join(t.TempDir(), "storage")
			if err := native.Build(checked, binary, native.Options{ProgramRegion: enabled, Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			resultRun := executeWith(t, []string{leakSanitizer()}, binary)
			if difference := disagreement(oracle, resultRun); difference != "" {
				t.Fatalf("storage invariant: %s %s", difference, resultRun.stderr)
			}
			report, err := leakcheck.Check(leakcheck.Program{
				C: checked, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
				BuildCounted: func(code, output string) error {
					return native.Build(code, output, native.Options{ProgramRegion: enabled, Count: true})
				},
				Execute: func(environment []string, name string, args ...string) leakcheck.Run {
					return leakRun(executeWith(t, environment, name, args...))
				},
			})
			if err != nil || report != "" {
				t.Fatalf("leak check: %v %s", err, report)
			}
			t.Logf("flag=%t: actual storage assertion, unchanged Node output, ASan/UBSan and leak check passed", enabled)
			if !enabled {
				return
			}
			// Runtime uniqueness must reject marked storage. Admit it deliberately at
			// the selected mapper branch and require the independent address invariant
			// to fail under the same sanitizer build and unchanged Node oracle.
			guard := regexp.MustCompile(`bool ` + regexp.QuoteMeta(unique) + ` = [^\n]+;`)
			if len(guard.FindAllString(checked, -1)) != 1 {
				t.Fatal("unique guard missing/ambiguous")
			}
			mutant := guard.ReplaceAllString(checked, "bool "+unique+" = true;")
			bad := filepath.Join(t.TempDir(), "reuse-member-mutant")
			if err := native.Build(mutant, bad, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			rejected := executeWith(t, []string{leakSanitizer()}, bad)
			if rejected.exitCode != 70 || !strings.Contains(string(rejected.stderr), "Program member array storage was reused") || disagreement(oracle, rejected) == "" {
				t.Fatalf("member reuse mutant escaped: %d %s", rejected.exitCode, rejected.stderr)
			}
			t.Log("reuse-admitted mutant caught by actual storage invariant under ASan/UBSan")
		})
	}
}
