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

// A program with generators may have a Program region (#8xcz03h): a generator's call is an ir.Call, which
// programAllocation never stamps, and internal/native/generator.go adopts nothing, so the frame stays a
// counted allocation while the Tree literals it walks are members. The fixture yields each member to a
// for-of over the call and through an Iterable<T> parameter, and Node holds the answer.
const programGeneratorsFixture = "internal/oracle/testdata/program_region/program_region_generators.a"

// programGeneratorConstructor is a generator's constructor, which the emitter writes from the frame's
// allocation to its return (generator.go): the statement right after the allocation, and the whole body.
var programGeneratorConstructor = regexp.MustCompile(`(?s)adamic_generator \*frame = adamic_generator_new\([^;]*\);\s*([^;]*;)(.*?)\n return frame;\n\}`)

// programGenerators is the fixture's Node run, and its C with a check, right before each constructor
// returns its frame, that the frame is no Program member, so no run can pass with one adopted.
func programGenerators(t *testing.T) (run, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, programGeneratorsFixture))
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	want := "branch2\nroot1\nbranch2,root1,\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		t.Fatalf("Node: %d %q %s", oracle.exitCode, oracle.stdout, oracle.stderr)
	}
	program := programRegionLowered(t, path, true)
	if diff := disagreement(oracle, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	if !strings.Contains(code, "adamic_program_adopt") || !strings.Contains(code, "adamic_generator_new(") {
		t.Fatal("the fixture must adopt Program members and make a generator frame")
	}
	matches := programGeneratorConstructor.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 || len(matches) != strings.Count(code, "adamic_generator_new(") {
		t.Fatalf("expected exactly one generator constructor, got %d of %d frame allocations", len(matches), strings.Count(code, "adamic_generator_new("))
	}
	m := matches[0]
	if next := code[m[2]:m[3]]; strings.Contains(next, "adopt") {
		t.Fatalf("the statement after the frame's allocation adopts: %s", next)
	}
	if body := code[m[2]:m[5]]; strings.Contains(body, "adopt") {
		t.Fatalf("the generator's constructor adopts: %s", body)
	}
	message := "generator frame is a Program member"
	assertion := fmt.Sprintf("\n if (adamic_program_is(frame)) adamic_panic(\"%s\", sizeof \"%s\" - 1);", message, message)
	return oracle, code[:m[5]] + assertion + code[m[5]:]
}

func TestProgramRegionGeneratorsNotAdopted(t *testing.T) {
	t.Parallel()
	oracle, checked := programGenerators(t)
	binary := filepath.Join(t.TempDir(), "generators")
	if err := programRegionBuild(t, checked, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		t.Fatalf("generators over members: %s %s", difference, result.stderr)
	}
	report, err := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(t, code, output, native.Options{ProgramRegion: true, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, args...))
		},
	})
	if err != nil || report != "" {
		t.Fatalf("leak check: %v %s", err, report)
	}
	// LeakSanitizer is the Linux leak proof; counters independently account for every member and frame.
	countedBinary := filepath.Join(t.TempDir(), "generators-counts")
	if err := programRegionBuild(t, checked, countedBinary, native.Options{ProgramRegion: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	command, args := pinnedStack(countedBinary)
	counted := execute(t, command, args...)
	observed := counted
	observed.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(oracle, observed); difference != "" {
		t.Fatalf("counted generators over members: %s %s", difference, counted.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		t.Fatal(report)
	}
	t.Logf("generators over Program members agree with Node under ASan/UBSan and the leak check: %s", counted.stderr)
}
