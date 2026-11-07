package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// buildLogLine is a line buildLog writes: kind, name, key12 or -, outcome, seconds. repass is compileUnits' second
// pass after a cold check's mismatch, every unit compiled again uncached.
var buildLogLine = regexp.MustCompile(
	`^(cCompile|cArchive|cLink|cBuild) (\S+) ([0-9a-f]{12}|-) (hit|miss|off|repass) ([0-9]+\.[0-9]{3})$`,
)

// A split build writes one cCompile line per unit and one cLink line to ADAMIC_BUILD_LOG, each saying whether its
// object came from the cache and how long clang ran, and the log never reaches the binary: the same program built with
// the log and without it, cached and uncached, is the same bytes. A mutant that drops the outcome, or says hit where
// clang compiled, fails the outcomes read here; one that writes 0 seconds fails the seconds.
//
// Not parallel: test.Setenv changes ADAMIC_BUILD_LOG and ADAMIC_GATE_UNCACHED for the process.
func TestSplitBuildLogsEachUnitAndTheLink(test *testing.T) {
	// The program's names are this run's own, in units.h, which every unit includes, so every unit's preprocessed bytes
	// and so its key are new here and the first cached build compiles them all, whatever a shared cache already holds.
	nonce := strconv.FormatInt(time.Now().UnixNano(), 36)
	source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
static int adamic_state_%[1]s = 0;
static int adamic_next_%[1]s(void);
static int adamic_twice_%[1]s(void);
static int adamic_next_%[1]s(void) { return ++adamic_state_%[1]s; }
static int adamic_twice_%[1]s(void) { return adamic_next_%[1]s() + adamic_next_%[1]s(); }
int main(void) { printf("%%d\n", adamic_twice_%[1]s()); return 0; }
`, nonce)
	_, units, splitCError := splitC(source)
	if splitCError != nil {
		test.Fatal(splitCError)
	}
	directory := test.TempDir()
	build := func(name, uncached string, logged bool) ([]byte, []string) {
		test.Helper()
		log := filepath.Join(directory, name+".log")
		if logged {
			test.Setenv("ADAMIC_BUILD_LOG", log)
		} else {
			test.Setenv("ADAMIC_BUILD_LOG", "")
		}
		test.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		binary := filepath.Join(directory, name, "program")
		if createDirectoryError := os.MkdirAll(filepath.Dir(binary), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if buildError := Build(source, binary, Options{Split: true, Jobs: 2}); buildError != nil {
			test.Fatal(buildError)
		}
		contents, readError := os.ReadFile(binary)
		if readError != nil {
			test.Fatal(readError)
		}
		written, readError := os.ReadFile(log)
		if !logged {
			if !os.IsNotExist(readError) {
				test.Fatalf("%s: a build without ADAMIC_BUILD_LOG wrote %s: %v", name, log, readError)
			}
			return contents, nil
		}
		if readError != nil {
			test.Fatalf("%s: no build log: %v", name, readError)
		}
		return contents, strings.Split(strings.TrimSuffix(string(written), "\n"), "\n")
	}
	// read checks one build's lines: a cCompile per unit with outcome, each unit once, then one cLink naming the
	// binary.
	read := func(name string, lines []string, outcome string) {
		test.Helper()
		compiled := map[string]bool{}
		links := 0
		for _, line := range lines {
			match := buildLogLine.FindStringSubmatch(line)
			if match == nil {
				test.Fatalf("%s: %q isn't a build log line", name, line)
			}
			seconds, parseError := strconv.ParseFloat(match[5], 64)
			if parseError != nil {
				test.Fatal(parseError)
			}
			// The runtime's archive, looked up or compiled for the build, is its own test's
			// (TestRuntimeArchiveLogsItsCompilesAndItsHit).
			if match[1] == "cArchive" || strings.HasPrefix(match[2], "runtime/") {
				continue
			}
			switch match[1] {
			case "cCompile":
				// An uncached unit is never keyed (#m62spx0: a unit's key reads clang's facts before clang runs, which
				// an uncached build skips), so its line names none; a cached one's always does.
				if match[4] != outcome || (match[3] == "-") != (outcome == "off") {
					test.Fatalf("%s: %q, want %s, keyed unless off", name, line, outcome)
				}
				if compiled[match[2]+" "+match[3]] {
					test.Fatalf("%s: %s logged twice", name, match[2])
				}
				compiled[match[2]+" "+match[3]] = true
				// A hit's seconds are its preprocessing alone, which can round to nothing; a compile's can't.
				if outcome != "hit" && seconds <= 0 {
					test.Fatalf("%s: %q says clang took no time", name, line)
				}
			case "cLink":
				links++
				if match[2] != "program" || match[3] != "-" || match[4] != "off" || seconds <= 0 {
					test.Fatalf("%s: %q, want the program's link, unkeyed, off, with its seconds", name, line)
				}
			default:
				test.Fatalf("%s: %q from a split build", name, line)
			}
		}
		if len(compiled) != len(units) || links != 1 {
			test.Fatalf(
				"%s: %d units and %d links logged, want %d and 1:\n%s",
				name,
				len(compiled),
				links,
				len(units),
				lines,
			)
		}
	}
	quietUncached, _ := build("quiet-uncached", "1", false)
	loggedUncached, lines := build("logged-uncached", "1", true)
	read("logged-uncached", lines, "off")
	loggedMiss, lines := build("logged-miss", "0", true)
	read("logged-miss", lines, "miss")
	loggedHit, lines := build("logged-hit", "0", true)
	read("logged-hit", lines, "hit")
	quietHit, _ := build("quiet-hit", "0", false)
	for name, binary := range map[string][]byte{
		"logged-uncached": loggedUncached,
		"logged-miss":     loggedMiss,
		"logged-hit":      loggedHit,
		"quiet-hit":       quietHit,
	} {
		if !bytes.Equal(binary, quietUncached) {
			test.Fatalf("%s's binary differs from the one built uncached without a log", name)
		}
	}
	test.Logf(
		"%d units: off, miss and hit each logged once a unit with one link, and five binaries the same bytes",
		len(units),
	)
}

// A name holding whitespace is one field of one line: an output path with a tab or a newline in it can't split the line
// or forge a second. Mutant: replacing spaces alone, as record does.
//
// Not parallel: test.Setenv changes ADAMIC_BUILD_LOG for the process.
func TestBuildLogKeepsANameOnOneLine(test *testing.T) {
	log := filepath.Join(test.TempDir(), "build.log")
	test.Setenv("ADAMIC_BUILD_LOG", log)
	buildLog("cLink", "a b\tc\ncCompile forged 0123456789ab hit 9.000", "", "off", time.Second)
	written, readError := os.ReadFile(log)
	if readError != nil {
		test.Fatal(readError)
	}
	if want := "cLink a_b_c_cCompile_forged_0123456789ab_hit_9.000 - off 1.000\n"; string(written) != want {
		test.Fatalf("wrote %q, want %q", written, want)
	}
}

// A cold check's mismatch compiles every unit again uncached (compileUnits), and that second pass's lines say repass,
// never off, so a reader counting units or a hit rate leaves them out from the line alone (Performance on #m62spx0).
// The first pass's lines keep their own outcomes: alpha's unchecked hit, and beta's checked hit that mismatched, a
// miss. One repass line a unit, each naming no key, beside the mismatch's own line. Mutant: the second pass logged off,
// red.
//
// Not parallel: test.Setenv changes ADAMIC_BUILD_LOG for the process.
func TestARepassAfterAMismatchSaysRepass(test *testing.T) {
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	units := []compilationUnit{
		{"alpha.c", unitFirst + "int adamic_alpha(void) { return VALUE; }\n"},
		{"beta.c", unitFirst + "int adamic_beta(void) { return VALUE + 1; }\n"},
	}
	root := test.TempDir()
	cache := filepath.Join(root, "cache")
	directory := filepath.Join(root, "units")
	if createDirectoryError := os.MkdirAll(directory, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	never := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
	never.check = func(compilationUnit) bool { return false }
	cached, compileUnitsError := compileUnits(
		units,
		common,
		unitKeyFlags,
		compiler,
		never,
		cache,
		directory,
		2,
		nil,
		false,
	)
	if compileUnitsError != nil {
		test.Fatal(compileUnitsError)
	}
	for _, object := range cached {
		if writeError := os.WriteFile(object, []byte("planted\n"), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	log := filepath.Join(root, "build.log")
	test.Setenv("ADAMIC_BUILD_LOG", log)
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
	keying.check = func(unit compilationUnit) bool { return unit.name == "beta.c" }
	if _, compileUnitsError := compileUnits(
		units,
		common,
		unitKeyFlags,
		compiler,
		keying,
		cache,
		directory,
		2,
		nil,
		false,
	); compileUnitsError != nil {
		test.Fatalf("a mismatched build failed: %v", compileUnitsError)
	}
	written, readError := os.ReadFile(log)
	if readError != nil {
		test.Fatal(readError)
	}
	outcomes, noted := map[string][]string{}, 0
	for line := range strings.SplitSeq(strings.TrimSuffix(string(written), "\n"), "\n") {
		// The mismatch's own line, noted where Loom's builder reads it (noteUnitMismatch), is no clang step's.
		if strings.HasPrefix(line, UnitMismatchPrefix) {
			noted++
			continue
		}
		match := buildLogLine.FindStringSubmatch(line)
		if match == nil || match[1] != "cCompile" {
			test.Fatalf("%q isn't a unit's build log line", line)
		}
		if (match[3] == "-") != (match[4] == "repass") {
			test.Fatalf("%q: a repass line names no key, and every other line of this build does", line)
		}
		outcomes[match[2]] = append(outcomes[match[2]], match[4])
	}
	if noted != 1 {
		test.Fatalf("the mismatch was noted %d times, want once:\n%s", noted, written)
	}
	for name, want := range map[string][]string{"alpha.c": {"hit", "repass"}, "beta.c": {"miss", "repass"}} {
		if got := outcomes[name]; !slices.Equal(got, want) {
			test.Fatalf("%s's lines say %v, want %v:\n%s", name, got, want, written)
		}
	}
}

// phaseLine is a line buildlog.Phase writes, in buildLog's format: phase, the phase it timed, no key, off, seconds.
var phaseLine = regexp.MustCompile(`^phase (load|lower|emit) - off [0-9]+\.[0-9]{3}$`)

// A program loaded, lowered and emitted writes one phase line for each to ADAMIC_BUILD_LOG, in buildLog's format, and
// one built in units writes one emit line still, since a split divides the C emitted once (#9x03wa7). The log never
// reaches what is built: the C and the split binary are the same bytes with it set and unset. Mutants, each red: the
// emit line written a unit rather than a program; load's line dropped; the log's path written into the C.
//
// Not parallel: test.Setenv changes ADAMIC_BUILD_LOG and ADAMIC_GATE_UNCACHED for the process.
func TestEachPhaseLogsOnceAProgram(test *testing.T) {
	directory := test.TempDir()
	entry := filepath.Join(directory, "main.ts")
	source := "function twice(value: number): number {\n\treturn value * 2;\n}\nfunction greet(name: string): string {\n\treturn `hello ${name}`;\n}\nconsole.log(`${twice(21)} ${greet('phase')}`);\n"
	if writeError := os.WriteFile(entry, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv("ADAMIC_GATE_UNCACHED", "1")
	// phases loads, lowers and emits the program, with the log set when logged, and builds the C in units when split.
	// It gives the C, the binary when split, and the phases the log names, in the order it names them.
	phases := func(name string, logged, split bool) (string, []byte, []string) {
		test.Helper()
		log := filepath.Join(directory, name+".log")
		if logged {
			test.Setenv("ADAMIC_BUILD_LOG", log)
		} else {
			test.Setenv("ADAMIC_BUILD_LOG", "")
		}
		loaded, loadError := load.Load([]string{entry})
		if loadError != nil {
			test.Fatal(loadError)
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			test.Fatal(lowerError)
		}
		code := C(program)
		var binary []byte
		if split {
			output := filepath.Join(directory, name, "program")
			if createDirectoryError := os.MkdirAll(filepath.Dir(output), 0o755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			if buildError := Build(code, output, Options{Split: true, Jobs: 2}); buildError != nil {
				test.Fatal(buildError)
			}
			var readError error
			if binary, readError = os.ReadFile(output); readError != nil {
				test.Fatal(readError)
			}
		}
		if !logged {
			return code, binary, nil
		}
		written, readError := os.ReadFile(log)
		if readError != nil {
			test.Fatalf("%s: no build log: %v", name, readError)
		}
		var named []string
		for line := range strings.SplitSeq(strings.TrimSuffix(string(written), "\n"), "\n") {
			// A split build's clang lines are TestSplitBuildLogsEachUnitAndTheLink's.
			if !strings.HasPrefix(line, "phase ") {
				continue
			}
			if !phaseLine.MatchString(line) {
				test.Fatalf("%s: %q isn't a phase line", name, line)
			}
			named = append(named, strings.Fields(line)[1])
		}
		return code, binary, named
	}
	want := []string{"load", "lower", "emit"}
	quiet, _, _ := phases("quiet", false, false)
	logged, _, named := phases("logged", true, false)
	if !slices.Equal(named, want) {
		test.Fatalf("a program loaded, lowered and emitted logged the phases %v, want %v", named, want)
	}
	if logged != quiet {
		test.Fatal("the C emitted with ADAMIC_BUILD_LOG set differs from the C emitted without it")
	}
	_, units, splitCError := splitC(logged)
	if splitCError != nil {
		test.Fatal(splitCError)
	}
	if len(units) < 2 {
		test.Fatalf("the program splits into %d units, want more than one for a split to divide", len(units))
	}
	_, quietBinary, _ := phases("quiet-split", false, true)
	_, loggedBinary, named := phases("logged-split", true, true)
	if !slices.Equal(named, want) {
		test.Fatalf("a program built in %d units logged the phases %v, want %v", len(units), named, want)
	}
	if !bytes.Equal(loggedBinary, quietBinary) {
		test.Fatal("the split binary built with ADAMIC_BUILD_LOG set differs from the one built without it")
	}
	test.Logf(
		"%d units: load, lower and emit each logged once a program, the C and the binary the same bytes",
		len(units),
	)
}
