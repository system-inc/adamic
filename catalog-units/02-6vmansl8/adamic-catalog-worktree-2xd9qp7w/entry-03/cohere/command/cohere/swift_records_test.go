package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// contractFixture reads one of the fixtures swift/Contract.md defines, from the same files the Swift
// encoder's tests read, so neither side can drift without the other's test failing.
func contractFixture(t *testing.T, name string) []string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "..", "swift", "Contract", name))
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{}
	for _, line := range strings.Split(strings.TrimRight(string(contents), "\n"), "\n") {
		lines = append(lines, line)
	}
	if len(lines) < 2 {
		t.Fatalf("%s holds %d records, which cannot be a contract fixture", name, len(lines))
	}
	return lines
}

// renderRecords feeds records through a fresh renderer and ends the run the way the engine would.
func renderRecords(t *testing.T, mode swiftMode, lines []string, engineExit int) (string, int, error) {
	t.Helper()
	var out bytes.Buffer
	run := newSwiftRun(&out, mode, "", time.Now())
	for index, line := range lines {
		if err := run.accept([]byte(line)); err != nil {
			run.writeUnfinished("the Swift engine broke its contract")
			return out.String(), 1, &recordError{index: index, err: err}
		}
	}
	exitCode, err := run.finish(engineExit, "exited "+string(rune('0'+engineExit)))
	return out.String(), exitCode, err
}

// recordError says which record a refusal came at, so a failing case names the line that tripped.
type recordError struct {
	index int
	err   error
}

func (e *recordError) Error() string { return e.err.Error() }

func requireLines(t *testing.T, output string, wanted ...string) {
	t.Helper()
	for _, line := range wanted {
		if !strings.Contains(output, line) {
			t.Errorf("missing %q from:\n%s", line, output)
		}
	}
}

func forbidLines(t *testing.T, output string, unwanted ...string) {
	t.Helper()
	for _, line := range unwanted {
		if strings.Contains(output, line) {
			t.Errorf("printed %q, which it must not:\n%s", line, output)
		}
	}
}

// Not parallel: its subtests render check runs through swiftRun, whose record acceptors set the package-level
// activeSummary
func TestSwiftRunRendersEachContractFixture(t *testing.T) {
	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("Clean", func(t *testing.T) {
		output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Clean.jsonl"), 0)
		if err != nil || exitCode != 0 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"package described in 412ms — 3 Swift files in the package, 2 of them ours\n",
			"fix: 0 of 2 files rewritten, 0 fixes applied, 0 refused\n",
			"format scope: changed files (working tree against HEAD): 0\n",
			"types: 0 diagnostics over 2 files in 820ms\n",
			"lint: 0 findings — 4 rules over 2 files, 512 nodes visited, walked by the fix phase (nothing was rewritten, so its findings still hold)\n",
			"coverage: 4 rules = 4 ran and found nothing\n",
			"  0 files crashed · details: cohere --coverage\n",
			"  config: no CohereSettings.json beside Package.swift, so every house rule ran at error",
			"  not checked: /project/Sources/Example/Generated.swift (marked // @generated)\n",
			"phases: fix ran in 31ms · types ran in 820ms · lint reused the fix phase's walk (nothing was rewritten) · unused skipped",
			"— package 412ms",
		)
		forbidLines(t, output, "did not check everything", "modified tree", "in scope (")
	})

	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("Findings", func(t *testing.T) {
		output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Findings.jsonl"), 1)
		if err != nil || exitCode != 1 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"fix: 1 of 2 files rewritten, 1 fixes applied, 0 refused, 1 reformatted\n",
			"/project/Sources/Example/Warning.swift:2:9 - warning: variable 'neverMutated' was never mutated; consider changing to 'let' constant [#VariableNeverMutated]\n",
			// The message's newline is collapsed, so the rule tag stays on the line with the position.
			"/project/Sources/Example/ByteRing.swift:159:61 - A force unwrap crashes the process when the value is nil. Say what happens on nil with guard let or if let. [cohere-swift/force-unwrapping/forceUnwrap]\n",
			"coverage: 4 rules = 1 found something + 2 ran and found nothing + 1 listened to no files\n",
			"  this binary was built from a modified tree, so no commit reproduces these findings\n",
		)
		// The silent rule is counted by default and named behind --coverage, as a TypeScript run does.
		forbidLines(t, output, "did not check everything", "cohere-swift/force-cast")
	})

	// --coverage names each rule the record names exactly once, under its category, and counts the
	// ones the record only counts.
	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("FindingsWithCoverage", func(t *testing.T) {
		var out bytes.Buffer
		run := newSwiftRun(&out, swiftModeCheck, "", time.Now())
		run.details = true
		for _, line := range contractFixture(t, "Findings.jsonl") {
			if err := run.accept([]byte(line)); err != nil {
				t.Fatal(err)
			}
		}
		output := out.String()
		requireLines(t, output,
			"  listened to no files (1): nothing gave it a file, or it declined every one",
			"    cohere-swift/force-cast\n",
			"  ran and found nothing (2): ",
			"    2 the engine's record counts and does not name\n",
		)
		if count := strings.Count(output, "cohere-swift/force-cast\n"); count != 1 {
			t.Errorf("the silent rule is named %d times, want 1:\n%s", count, output)
		}
		forbidLines(t, output, "details: cohere --coverage")
	})

	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("TypesBail", func(t *testing.T) {
		output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "TypesBail.jsonl"), 1)
		if err != nil || exitCode != 1 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"/project/Sources/Example/Broken.swift:3:23 - error: cannot convert value of type 'String' to specified type 'Int'\n",
			"fix skipped (--no-fix)",
			"lint did not run (types bailed: 1 type errors — lint findings against wrong semantics are noise)",
			"  this run did not check everything — the phases above say what was not checked\n",
		)
		// A compiler finding with no group has no bracket.
		forbidLines(t, output, "Int' [#")
	})

	// Contract 2: the file the engine could not read is named with the excluded files, and the run is
	// incomplete with exit 1 though every phase ran and nothing was found.
	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("Unreadable", func(t *testing.T) {
		output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Unreadable.jsonl"), 1)
		if err != nil || exitCode != 1 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"  not checked: /project/Sources/Example/Latin1.swift (could not be read: it is not valid UTF-8 or could not be opened (Foundation could not decode it as UTF-8))\n",
			"  not checked: /project/Sources/Example/Generated.swift (marked // @generated)\n",
			"  this run did not check everything — the notes above name the files nothing checked\n",
		)
	})

	// The --unused report renders after lint and is not a gate: two items, and the summary still counts no
	// findings and exits 0. Were the renderer to count them, the summary check would refuse the stream.
	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("Unused", func(t *testing.T) {
		output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Unused.jsonl"), 0)
		if err != nil || exitCode != 0 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"\nunused: a report, not a gate — nothing here fails a build\n",
			"\n  imports nothing uses — 3\n",
			// Two items on one line arrive out of column order and print in it.
			"    /project/Sources/Example/Main.swift:1:1 — import Foundation\n    /project/Sources/Example/Main.swift:1:20 — import Darwin\n",
			"    /project/Sources/Example/Tally.swift:2:1 — import Combine\n",
			"  looked at 2 files and 6 imports (cohere-swift/correctness-no-unused-import)\n",
			"  not checked for unused imports: 1 files (it has #if, and the index describes only the configuration the build compiled)\n",
			"  not checked for unused imports: 1 files (the build has not compiled it as it stands)\n",
			"  never reported: 1 imports (re-exported with @_exported, which is API)\n",
			// The second rule's section, from a real run on ahraos-presence: its own heading and count, its
			// items, and its own coverage lines, every skip reason by name.
			"\n  declarations nothing uses — 2\n",
			"    /project/Sources/AhraOSPresence/Stage/StagePane+HandTouchProbe.swift:42:13 — var deepestAt\n",
			"    /project/Sources/AhraOSPresence/Studio/Panels/VariantsMenuProbe.swift:349:25 — func after(_:_:)\n",
			"  looked at 1092 files and 4311 declarations (cohere-swift/correctness-no-unused-declaration)\n",
			"  not checked for unused declarations: 156 files (it has #if, and the index describes only the configuration the build compiled)\n",
			"  never reported: 358 declarations (a stored property of a type whose conformances may read every stored property (Codable))\n",
			"  never reported: 175 declarations (an override or a protocol witness, reached through what it overrides)\n",
			"  never reported: 23 declarations (reached by the Objective-C runtime (@objc, @IBAction, @IBOutlet, @NSManaged, dynamic))\n",
		)
		forbidLines(t, output, "did not check everything", "Main.swift:1:1 - ", "deepestAt - ")
		if strings.Index(output, "imports nothing uses") < strings.Index(output, "lint: ") {
			t.Errorf("the unused report printed before the lint line:\n%s", output)
		}
		// Each rule's section in the order the engine wrote them, and every skip reason printed: ten for
		// declarations, counted rather than listed so a dropped one cannot hide among the pinned three.
		if strings.Index(output, "declarations nothing uses") < strings.Index(output, "never reported: 1 imports") {
			t.Errorf("the declarations section printed before the imports section ended:\n%s", output)
		}
		if got := strings.Count(output, "never reported: "); got != 11 {
			t.Errorf("%d never-reported lines, the fixture holds 1 for imports and 10 for declarations:\n%s", got, output)
		}
	})

	// A rule the front door has no words for prints under its own id and counts items, never under
	// another rule's heading.
	// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
	t.Run("UnusedUnknownRule", func(t *testing.T) {
		lines := []string{}
		for _, line := range contractFixture(t, "Unused.jsonl") {
			lines = append(lines, strings.ReplaceAll(line, "cohere-swift/correctness-no-unused-import", "cohere-swift/unused-parameter"))
		}
		output, exitCode, err := renderRecords(t, swiftModeCheck, lines, 0)
		if err != nil || exitCode != 0 {
			t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
		}
		requireLines(t, output,
			"\n  cohere-swift/unused-parameter — 3\n",
			"  looked at 2 files and 6 items (cohere-swift/unused-parameter)\n",
			"  not checked for unused items: 1 files (the build has not compiled it as it stands)\n",
			// The known rule beside it keeps its own words.
			"\n  declarations nothing uses — 2\n",
		)
		forbidLines(t, output, "imports nothing uses")
	})

	// The contract's crash: the stream stops after the fix phase. Exit 0 or 1 without a summary is a
	// crash whatever the code says, so both are asserted.
	for _, engineExit := range []int{0, 1} {
		// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
		t.Run("CrashWithoutSummary", func(t *testing.T) {
			output, exitCode, err := renderRecords(t, swiftModeCheck, contractFixture(t, "CrashWithoutSummary.jsonl"), engineExit)
			if err == nil || exitCode != 1 {
				t.Fatalf("a stream with no summary exited %d with err %v\n%s", exitCode, err, output)
			}
			if !strings.Contains(err.Error(), "without a summary, so nothing was checked") {
				t.Errorf("the crash did not say nothing was checked: %v", err)
			}
			requireLines(t, output,
				"types did not run (the Swift engine exited",
				"  this run did not check everything",
			)
		})
	}
}

// TestSwiftRunRefusesBrokenStreams holds every refusal the contract asks for, each as a mutation of a
// real fixture. Every one must exit 1 with an error, and none may print a clean bill of health.
// Not parallel: its subtests render check runs through swiftRun, whose record acceptors set the package-level
// activeSummary
func TestSwiftRunRefusesBrokenStreams(t *testing.T) {
	findings := contractFixture(t, "Findings.jsonl")
	clean := contractFixture(t, "Clean.jsonl")
	bail := contractFixture(t, "TypesBail.jsonl")
	unreadable := contractFixture(t, "Unreadable.jsonl")
	unused := contractFixture(t, "Unused.jsonl")
	last := len(findings) - 1

	replace := func(lines []string, index int, line string) []string {
		mutated := append([]string(nil), lines...)
		mutated[index] = line
		return mutated
	}
	without := func(lines []string, index int) []string {
		mutated := append([]string(nil), lines[:index]...)
		return append(mutated, lines[index+1:]...)
	}

	cases := []struct {
		name       string
		lines      []string
		engineExit int
		reason     string
	}{
		{"a line that is not a record", replace(findings, 4, "warning: the build printed to stdout"), 1, "not a record"},
		{"a record of unknown kind", replace(findings, 4, `{"kind":"telemetry"}`), 1, "unknown kind"},
		{"provenance not first", without(findings, 0), 1, "provenance is always first"},
		{"a contract the front door does not speak", replace(findings, 0, strings.Replace(findings[0], `"contract":3`, `"contract":2`, 1)), 1, "contract 2"},
		{"a provenance with no contract", replace(findings, 0, strings.Replace(findings[0], `"contract":3,`, "", 1)), 1, "contract none"},
		{"a finding the summary does not count", replace(findings, last, strings.Replace(findings[last], `"findings":2`, `"findings":1`, 1)), 1, "counts 1 findings and 2"},
		{"a types record that disagrees with its findings", replace(findings, 5, strings.Replace(findings[5], `"diagnostics":1`, `"diagnostics":0`, 1)), 1, "types record counts 0"},
		{"a lint record that disagrees with its findings", replace(findings, 8, strings.Replace(findings[8], `"findings":1`, `"findings":3`, 1)), 1, "lint record counts 3"},
		{"a missing phase record", without(findings, 9), 1, "expects lint"},
		{"phases out of order", replace(clean, 3, `{"kind":"phase","name":"lint","outcome":"ran","elapsedMilliseconds":1,"detail":""}`), 0, "expects fix"},
		{"an outcome the pipeline does not have", replace(clean, 3, strings.Replace(clean[3], `"ran"`, `"partial"`, 1)), 0, "not one of"},
		{"a record after the summary", append(append([]string(nil), clean...), clean[1]), 0, "after the summary"},
		// The dangerous direction: lint was cut off and the summary claims the run was complete.
		{"a summary claiming complete over a bailed lint", replace(bail, len(bail)-1, strings.Replace(bail[len(bail)-1], `"complete":false`, `"complete":true`, 1)), 1, "calls the run complete"},
		{"a summary whose exit code disagrees with its own findings", replace(findings, last, strings.Replace(findings[last], `"exitCode":1`, `"exitCode":0`, 1)), 0, "summary says exit 0"},
		{"an engine whose exit disagrees with its summary", findings, 0, "said it would exit 1"},
		{"a summary missing its fields", replace(findings, last, `{"kind":"summary"}`), 1, "missing findings"},
		// changedFiles, additive within contract 3: "N cohered" is the list, so the list must be the count.
		{"a changed-file list shorter than its count", replace(findings, 2, strings.Replace(findings[2], `"filesRewritten":1`, `"filesRewritten":2`, 1)), 1, "counts 2 files rewritten and lists 1"},
		{"a file listed twice", replace(findings, 2, strings.Replace(strings.Replace(findings[2], `"filesRewritten":1`, `"filesRewritten":2`, 1), `"formatted":true}]`, `"formatted":true},{"file":"/project/Sources/Example/ByteRing.swift","fixedBy":{},"formatted":true}]`, 1)), 1, "lists /project/Sources/Example/ByteRing.swift twice"},
		{"a file nothing changed", replace(findings, 2, strings.Replace(strings.Replace(findings[2], `{"cohere-swift/legacy-constructors":1}`, `{}`, 1), `"formatted":true`, `"formatted":false`, 1)), 1, "rewritten by no fix and not by the formatter"},
		// Contract 2's unreadable record: a summary may not call the run complete over one, and it has one place.
		{"a summary claiming complete over an unreadable file", replace(unreadable, len(unreadable)-1, `{"kind":"summary","findings":0,"complete":true,"exitCode":0}`), 0, "calls the run complete"},
		{"an unreadable record after a phase", append(append(append([]string(nil), unreadable[:5]...), unreadable[2]), unreadable[5:]...), 1, "after phase fix"},
		{"an unreadable record before the project", append([]string{unreadable[0], unreadable[2]}, unreadable[1:]...), 1, "before the project record"},
		// The --unused report: well-formed, in its place, and counted by its coverage record.
		{"an unused record that does not decode", replace(unused, 8, `{"kind":"unused","line":"one"}`), 0, "does not decode"},
		{"an unused record with no subject", replace(unused, 8, strings.Replace(unused[8], `"subject":`, `"unnamed":`, 1)), 0, "missing its file, position, rule or subject"},
		{"an unused record before the lint phase", append(append(append([]string(nil), unused[:7]...), unused[8], unused[7]), unused[9:]...), 0, "outside the unused phase"},
		{"an unused coverage count that disagrees", replace(unused, 11, strings.Replace(unused[11], `"found":3`, `"found":4`, 1)), 0, "counts 4 found and 3"},
		// The fixture holds two rules: unused-import's records at 8-10 and its coverage at 11, then
		// unused-declaration's records at 12-13 and its coverage at 14.
		{"unused records with no coverage record", without(unused, 14), 0, "with no unusedCoverage record"},
		{"one rule's records interleaved before another's coverage", without(unused, 11), 0, "of another rule arrived before the coverage record"},
		{"an unused phase that ran with no coverage", without(without(without(without(without(without(without(unused, 14), 13), 12), 11), 10), 9), 8), 0, "no unusedCoverage record said"},
	}

	for _, testCase := range cases {
		// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
		t.Run(testCase.name, func(t *testing.T) {
			output, exitCode, err := renderRecords(t, swiftModeCheck, testCase.lines, testCase.engineExit)
			if err == nil || exitCode != 1 {
				t.Fatalf("accepted a broken stream: exit %d, err %v\n%s", exitCode, err, output)
			}
			if !strings.Contains(err.Error(), testCase.reason) {
				t.Errorf("refused for the wrong reason: %v, want one naming %q", err, testCase.reason)
			}
		})
	}
}

// TestSwiftRunGroupsExcludedFiles holds both sides of the limit: a reason with many files is one
// counted line, and a reason with few still names each file.
// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
func TestSwiftRunGroupsExcludedFiles(t *testing.T) {
	clean := contractFixture(t, "Clean.jsonl")
	vendored := []string{}
	for index := range excludedNamedLimit + 1 {
		vendored = append(vendored, `{"file":"/project/Vendor/F`+string(rune('a'+index))+`.swift","reason":"vendored under Vendor/"}`)
	}
	clean[1] = strings.Replace(clean[1], `"excluded":[`, `"excluded":[`+strings.Join(vendored, ",")+",", 1)

	output, exitCode, err := renderRecords(t, swiftModeCheck, clean, 0)
	if err != nil || exitCode != 0 {
		t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
	}
	requireLines(t, output,
		"  not checked: 6 files (vendored under Vendor/)\n",
		"  not checked: /project/Sources/Example/Generated.swift (marked // @generated)\n",
	)
	forbidLines(t, output, "/project/Vendor/Fa.swift")
}

// TestSwiftRunBelievesAnEngineThatFellShort holds the allowed direction of the completeness check: an
// engine that says it did not check everything is believed even when its records show no gap, because
// it can see gaps no record carries, such as a file it could not read.
// Not parallel: it renders a check run through swiftRun, whose record acceptors set the package-level activeSummary
func TestSwiftRunBelievesAnEngineThatFellShort(t *testing.T) {
	clean := contractFixture(t, "Clean.jsonl")
	last := len(clean) - 1
	clean[last] = strings.Replace(strings.Replace(clean[last], `"complete":true`, `"complete":false`, 1), `"exitCode":0`, `"exitCode":1`, 1)

	output, exitCode, err := renderRecords(t, swiftModeCheck, clean, 1)
	if err != nil || exitCode != 1 {
		t.Fatalf("exit %d, err %v\n%s", exitCode, err, output)
	}
	requireLines(t, output, "  this run did not check everything — the Swift engine reported that it fell short without a record saying where\n")
}

// TestSwiftRunListingModes holds the two runs that end without a summary on purpose.
func TestSwiftRunListingModes(t *testing.T) {
	t.Parallel()
	provenance := contractFixture(t, "Clean.jsonl")[0]

	t.Run("version", func(t *testing.T) {
		t.Parallel()
		output, exitCode, err := renderRecords(t, swiftModeVersion, []string{provenance}, 0)
		if err != nil || exitCode != 0 {
			t.Fatalf("exit %d, err %v", exitCode, err)
		}
		requireLines(t, output, "swift engine: cohere-swift 0.1.0 (commit dev, swiftlang-6.4.0.34.1, swift-syntax 604.0.0, swift-format 604.0.0)\n")
	})

	t.Run("rules, sorted", func(t *testing.T) {
		t.Parallel()
		lines := []string{
			provenance,
			`{"kind":"rule","name":"cohere-swift/no-try-bang","severity":"error"}`,
			`{"kind":"rule","name":"cohere-swift/force-unwrapping","severity":"warning"}`,
		}
		output, exitCode, err := renderRecords(t, swiftModeRules, lines, 0)
		if err != nil || exitCode != 0 {
			t.Fatalf("exit %d, err %v", exitCode, err)
		}
		if output != "cohere-swift/force-unwrapping\ncohere-swift/no-try-bang\n" {
			t.Errorf("rules printed as %q", output)
		}
		enabled, _, err := renderRecords(t, swiftModeRulesEnabled, lines, 0)
		if err != nil || enabled != "cohere-swift/force-unwrapping\twarning\ncohere-swift/no-try-bang\terror\n" {
			t.Errorf("rules-enabled printed as %q (err %v)", enabled, err)
		}
	})

	t.Run("a finding in a listing is refused", func(t *testing.T) {
		t.Parallel()
		findings := contractFixture(t, "Findings.jsonl")
		_, exitCode, err := renderRecords(t, swiftModeRules, []string{provenance, findings[4]}, 0)
		if err == nil || exitCode != 1 {
			t.Fatalf("a finding in a rules listing was accepted")
		}
	})

	t.Run("an engine that could not run", func(t *testing.T) {
		t.Parallel()
		_, exitCode, err := renderRecords(t, swiftModeCheck, []string{provenance}, 2)
		if err == nil || exitCode != 1 || !strings.Contains(err.Error(), "did not run, so nothing was checked") {
			t.Fatalf("exit 2 was rendered as exit %d, err %v", exitCode, err)
		}
	})
}

func TestSwiftEngineArguments(t *testing.T) {
	t.Parallel()
	location := projectLocation{
		Root:               "/work/macos",
		Engine:             engineSwift,
		LintConfigFileName: "/work/macos/sub/Local.json",
		ArgumentBase:       "/work/macos/sub",
	}
	values := map[string]string{"no-fix": "true", "single-threaded": "false", "fix-passes": "4", "rules": "true"}
	value := func(name string) string { return values[name] }

	arguments, mode, err := swiftEngineArguments(location,
		map[string]bool{"no-fix": true, "single-threaded": true, "fix-passes": true, "lint-config": true},
		value, []string{"Thing.swift", "/abs/Other.swift"})
	if err != nil {
		t.Fatal(err)
	}
	want := "--contract 3 --root /work/macos --no-fix --lint-config /work/macos/sub/Local.json --fix-passes 4 /work/macos/sub/Thing.swift /abs/Other.swift"
	if strings.Join(arguments, " ") != want || mode != swiftModeCheck {
		// `--single-threaded=false` was typed and is false, so it is not forwarded as a switch.
		t.Errorf("arguments %q mode %s, want %q", strings.Join(arguments, " "), mode, want)
	}

	_, mode, _ = swiftEngineArguments(location, map[string]bool{"rules": true}, value, nil)
	if mode != swiftModeRules {
		t.Errorf("--rules ran in mode %s", mode)
	}

	if _, _, err := swiftEngineArguments(location, map[string]bool{"explain": true}, value, nil); err == nil || !strings.Contains(err.Error(), "--explain") {
		t.Errorf("--explain was not refused by name: %v", err)
	}
}

// A flag the engine does not read and the front door does not answer is refused by name, never dropped: a
// dropped --format-only made a format gate type-check and lint the whole package (#nqb3mjv). A switch typed
// as false asks for nothing and passes, and the flags the front door answers itself pass too.
func TestSwiftEngineArgumentsRefuseWhatNothingHonors(t *testing.T) {
	t.Parallel()
	location := projectLocation{Root: "/work/macos", Engine: engineSwift, ArgumentBase: "/work/macos"}
	values := map[string]string{"no-format": "false", "verbose": "true", "phases": "true", "coverage": "true"}
	value := func(name string) string {
		if set, found := values[name]; found {
			return set
		}
		return "true"
	}

	for flag, says := range map[string]string{
		"format-only":    "--format-only is not implemented for Swift yet: the engine has no format-only mode",
		"no-cache":       "--no-cache is not implemented for Swift yet",
		"print-config":   "--print-config is not implemented for Swift yet",
		"stdin-filepath": "--stdin-filepath is not implemented for Swift yet",
		"cache-dump":     "--cache-dump is not implemented for Swift yet",
	} {
		if _, _, err := swiftEngineArguments(location, map[string]bool{"no-fix": true, flag: true}, value, nil); err == nil ||
			!strings.Contains(err.Error(), says) || !strings.Contains(err.Error(), "nothing was checked") {
			t.Errorf("--%s was not refused by name: %v", flag, err)
		}
	}

	values["no-format"] = "true"
	if _, _, err := swiftEngineArguments(location, map[string]bool{"no-format": true}, value, nil); err == nil ||
		!strings.Contains(err.Error(), "formats on every run") {
		t.Errorf("--no-format was not refused by name: %v", err)
	}
	values["no-format"] = "false"

	given := map[string]bool{"no-fix": true, "no-format": true, "verbose": true, "phases": true, "coverage": true, "directory": true}
	arguments, _, err := swiftEngineArguments(location, given, value, nil)
	if err != nil || strings.Join(arguments, " ") != "--contract 3 --root /work/macos --no-fix" {
		t.Errorf("the flags the front door answers were not passed over: %q (%v)", strings.Join(arguments, " "), err)
	}
}

// TestSwiftRunListsTheFilesItRewrote: a fix record with its list fills the run summary the footer counts as
// cohered, each path relative to the package, with the fixes by rule. A record from an engine before the list
// has none, and the summary is left unlisted, never set to no files, so the footer prints no cohered count.
// Not parallel: it assigns the package-level activeSummary and reads back what the Swift record acceptors set
// there
func TestSwiftRunListsTheFilesItRewrote(t *testing.T) {
	activeSummary = runSummary{}
	if _, _, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Findings.jsonl"), 1); err != nil {
		t.Fatal(err)
	}
	want := []changedFile{{Path: "Sources/Example/ByteRing.swift", Fixed: true, Formatted: true, FixedBy: map[string]int{"cohere-swift/legacy-constructors": 1}}}
	if !reflect.DeepEqual(activeSummary.Changed, want) {
		t.Errorf("the summary lists %+v, expected %+v", activeSummary.Changed, want)
	}
	if activeSummary.cohered() != 1 {
		t.Errorf("cohered %d, expected the one file listed", activeSummary.cohered())
	}

	activeSummary = runSummary{}
	output, _, err := renderRecords(t, swiftModeCheck, contractFixture(t, "Clean.jsonl"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if activeSummary.Changed != nil {
		t.Errorf("a stream with no list set the summary to %+v; it is not listed, not empty", activeSummary.Changed)
	}
	forbidLines(t, output, "cohered")
}
