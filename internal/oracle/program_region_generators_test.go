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
var programGeneratorConstructor = regexp.MustCompile(
	`(?s)adamic_generator \*frame = adamic_generator_new\([^;]*\);\s*([^;]*;)(.*?)\n return frame;\n\}`,
)

// programGenerators is the fixture's Node run, and its C with a check, right before each constructor
// returns its frame, that the frame is no Program member, so no run can pass with one adopted.
func programGenerators(test *testing.T) (run, string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, programGeneratorsFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle := onNode(test, path)
	want := "branch2\nroot1\nbranch2,root1,\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		test.Fatalf("Node: %d %q %s", oracle.exitCode, oracle.stdout, oracle.stderr)
	}
	program := programRegionLowered(test, path, true)
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	if !strings.Contains(code, "adamic_program_adopt") || !strings.Contains(code, "adamic_generator_new(") {
		test.Fatal("the fixture must adopt Program members and make a generator frame")
	}
	matches := programGeneratorConstructor.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 || len(matches) != strings.Count(code, "adamic_generator_new(") {
		test.Fatalf(
			"expected exactly one generator constructor, got %d of %d frame allocations",
			len(matches),
			strings.Count(code, "adamic_generator_new("),
		)
	}
	m := matches[0]
	if next := code[m[2]:m[3]]; strings.Contains(next, "adopt") {
		test.Fatalf("the statement after the frame's allocation adopts: %s", next)
	}
	if body := code[m[2]:m[5]]; strings.Contains(body, "adopt") {
		test.Fatalf("the generator's constructor adopts: %s", body)
	}
	message := "generator frame is a Program member"
	assertion := fmt.Sprintf(
		"\n if (adamic_program_is(frame)) adamic_panic(\"%s\", sizeof \"%s\" - 1);",
		message,
		message,
	)
	return oracle, code[:m[5]] + assertion + code[m[5]:]
}

func TestProgramRegionGeneratorsNotAdopted(test *testing.T) {
	test.Parallel()
	oracle, checked := programGenerators(test)
	programGeneratorsHold(test, oracle, checked, "generators")
}

// programGeneratorsHold builds checked, a generators fixture's C with its run-time checks, with the
// Program region on, and holds it to oracle under ASan and UBSan, LeakSanitizer and the counted build.
func programGeneratorsHold(test *testing.T, oracle run, checked, name string) {
	test.Helper()
	binary := filepath.Join(test.TempDir(), name)
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		binary,
		native.Options{ProgramRegion: true, Sanitize: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		test.Fatalf("generators over members: %s %s", difference, result.stderr)
	}
	report, checkError := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(test.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(test, code, output, native.Options{ProgramRegion: true, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(test, environment, name, args...))
		},
	})
	if checkError != nil || report != "" {
		test.Fatalf("leak check: %v %s", checkError, report)
	}
	// LeakSanitizer is the Linux leak proof; counters independently account for every member and frame.
	countedBinary := filepath.Join(test.TempDir(), name+"-counts")
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		countedBinary,
		native.Options{ProgramRegion: true, Count: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	command, args := pinnedStack(countedBinary)
	counted := execute(test, command, args...)
	observed := counted
	observed.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(oracle, observed); difference != "" {
		test.Fatalf("counted generators over members: %s %s", difference, counted.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	test.Logf("generators over Program members agree with Node under ASan/UBSan and the leak check: %s", counted.stderr)
}

// programGeneratorInPlaceFixture is programGeneratorsFixture's walk written in place and called, and
// a second generator written in place into the Iterable<T> parameter (#qk8rztp S3b): each frame is
// made through a closure of its own, its carrier, which captures the members' locals.
const programGeneratorInPlaceFixture = "internal/oracle/testdata/program_region/program_region_generator_in_place.a"

// programClosureGeneratorEntry is the entry of a generator whose call is made through a closure, its
// carrier (generator.go): the frame's allocation, its retain of the carrier, and the entry's whole
// lines after it, none or more, up to its return.
var programClosureGeneratorEntry = regexp.MustCompile(
	`adamic_generator \*frame = adamic_generator_new\([^;]*\);\n frame->self = adamic_retain\(self\); \(void\)arguments;\n((?:[^\n]*\n)*?) return \(adamic_value\)\{\.reference = frame\};\n\}`,
)

// Memory on #qk8rztp S3b: with the Program region on, a generator written in place is no member, and
// neither is its carrier, the closure its call is made through, which nestedGeneratorCall makes and
// programAllocation never stamps (lower.go, the generators' invariant 2). Lowering says so of each
// function, the entries adopt nothing, and each entry checks, right before it returns, that neither
// its frame nor the carrier it retained is a member, so no run can pass with one adopted. The run is
// held to Node under the sanitizers, LeakSanitizer and the counted build.
func TestProgramRegionGeneratorInPlaceNotAdopted(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, programGeneratorInPlaceFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle := onNode(test, path)
	want := "branch2\nroot1\nbranch2,root1,\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		test.Fatalf("Node: %d %q %s", oracle.exitCode, oracle.stdout, oracle.stderr)
	}
	program := programRegionLowered(test, path, true)
	if len(program.ProgramTypes) == 0 {
		test.Fatal("no Program types: the region must engage")
	}
	written := 0
	for index, function := range program.Functions {
		if function.Generator && function.Closure && function.NestedParent == 0 {
			written++
			if function.ProgramRegion {
				test.Fatalf("function %d, a generator written in place, is a Program member", index)
			}
		}
	}
	if written != 2 {
		test.Fatalf("%d generators written in place, want 2", written)
	}
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	if !strings.Contains(code, "adamic_program_adopt") {
		test.Fatal("the fixture must adopt Program members")
	}
	matches := programClosureGeneratorEntry.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 2 || len(matches) != strings.Count(code, "adamic_generator_new(") {
		test.Fatalf(
			"expected two entries made through a carrier, got %d of %d frame allocations",
			len(matches),
			strings.Count(code, "adamic_generator_new("),
		)
	}
	message := "generator frame or carrier is a Program member"
	assertion := fmt.Sprintf(
		" if (adamic_program_is(frame) || adamic_program_is(self)) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n",
		message,
		message,
	)
	checked := code
	// From the last entry back, so each insertion leaves the earlier entries where they were found.
	for index := len(matches) - 1; index >= 0; index-- {
		m := matches[index]
		if body := code[m[2]:m[3]]; strings.Contains(body, "adopt") {
			test.Fatalf("the entry made through a carrier adopts: %s", body)
		}
		checked = checked[:m[3]] + assertion + checked[m[3]:]
	}
	programGeneratorsHold(test, oracle, checked, "generators-in-place")
}
