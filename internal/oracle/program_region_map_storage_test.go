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
func TestProgramRegionCountedMapperStorage(test *testing.T) {
	test.Parallel()
	programRegionMapperStorage(test, false)
}
func TestProgramRegionMemberMapperStorage(test *testing.T) {
	test.Parallel()
	programRegionMapperStorage(test, true)
}
func programRegionMapperStorage(test *testing.T, enabled bool) {
	test.Helper()
	path, _ := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/program_region/program_region_map_storage.a"),
	)
	oracle := onNode(test, path)
	program := programRegionLowered(test, path, enabled)
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	allocation := regexp.MustCompile(
		`adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_make\([^\n]+;`,
	)
	matches := allocation.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 {
		test.Fatalf("expected exactly one reusable mapper allocation, got %d", len(matches))
	}
	match := matches[0]
	result, unique, source := code[match[2]:match[3]], code[match[4]:match[5]], code[match[6]:match[7]]
	condition, message := result+" != "+source, "counted mapper control did not reuse"
	if enabled {
		condition, message = result+" == "+source, "Program member array storage was reused"
	}
	assertion := fmt.Sprintf(
		"\n if (adamic_program_is(%s) != %t) adamic_panic(\"mapper fixture has wrong membership\", sizeof \"mapper fixture has wrong membership\" - 1);\n if (%s) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n",
		source,
		enabled,
		condition,
		message,
		message,
	)
	checked := code[:match[1]] + assertion + code[match[1]:]
	binary := filepath.Join(test.TempDir(), "storage")
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		binary,
		native.Options{ProgramRegion: enabled, Sanitize: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	resultRun := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, resultRun); difference != "" {
		test.Fatalf("storage invariant: %s %s", difference, resultRun.stderr)
	}
	report, checkError := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(test.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(test, code, output, native.Options{ProgramRegion: enabled, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(test, environment, name, args...))
		},
	})
	if checkError != nil || report != "" {
		test.Fatalf("leak check: %v %s", checkError, report)
	}
	// LeakSanitizer is the Linux leak proof; counters independently account for
	// every member, and must run here too rather than only on macOS.
	countedBinary := filepath.Join(test.TempDir(), "storage-counts")
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		countedBinary,
		native.Options{ProgramRegion: enabled, Count: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	command, args := pinnedStack(countedBinary)
	counted := execute(test, command, args...)
	observed := counted
	observed.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(oracle, observed); difference != "" {
		test.Fatalf("counted storage invariant: %s %s", difference, counted.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	test.Logf("flag=%t: %s", enabled, counted.stderr)
	test.Logf("flag=%t: actual storage assertion, unchanged Node output, ASan/UBSan and leak check passed", enabled)
	if !enabled {
		return
	}
	// Runtime uniqueness must reject marked storage. Admit it deliberately at
	// the selected mapper branch and require the independent address invariant
	// to fail under the same sanitizer build and unchanged Node oracle.
	guard := regexp.MustCompile(`bool ` + regexp.QuoteMeta(unique) + ` = [^\n]+;`)
	if len(guard.FindAllString(checked, -1)) != 1 {
		test.Fatal("unique guard missing/ambiguous")
	}
	mutant := guard.ReplaceAllString(checked, "bool "+unique+" = true;")
	bad := filepath.Join(test.TempDir(), "reuse-member-mutant")
	if programRegionBuildError := programRegionBuild(
		test,
		mutant,
		bad,
		native.Options{ProgramRegion: true, Sanitize: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	rejected := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, bad)
	if rejected.exitCode != 70 ||
		!strings.Contains(string(rejected.stderr), "Program member array storage was reused") ||
		disagreement(oracle, rejected) == "" {
		test.Fatalf("member reuse mutant escaped: %d %s", rejected.exitCode, rejected.stderr)
	}
	test.Log("reuse-admitted mutant caught by actual storage invariant under ASan/UBSan")
}
