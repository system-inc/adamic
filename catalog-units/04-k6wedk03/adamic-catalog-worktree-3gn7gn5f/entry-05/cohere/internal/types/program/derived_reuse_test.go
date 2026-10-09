package program_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// markerRule is a rule that reads another file: it reports every variable declaration with the text of
// marker.ts, which it finds through the program's file list. fingerprint is its ProgramFingerprint, nil for
// none. runs counts the files it was run on, by base name.
func markerRule(name string, fingerprint func(rule.Program, any) [sha256.Size]byte, runs *sync.Map) rule.Rule {
	return rule.Rule{
		Name:               name,
		ProgramReads:       rule.ReadsOtherFiles,
		ProgramFingerprint: fingerprint,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if filepath.Base(ctx.SourceFile.FileName().AsString()) == "marker.ts" {
				return nil
			}
			count, _ := runs.LoadOrStore(filepath.Base(ctx.SourceFile.FileName().AsString()), new(int))
			*count.(*int)++
			marker := markerText(ctx.Program)
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: "marker", Description: "the marker says " + marker})
			}}
		},
	}
}

// markerText is marker.ts's text, trimmed, read through the program's file list.
func markerText(program rule.Program) string {
	for _, sourceFile := range program.SourceFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == "marker.ts" {
			return strings.TrimSpace(sourceFile.Text())
		}
	}
	return ""
}

// markerFingerprint is markerRule's complete fingerprint: marker.ts's text is all it reads beyond its file.
func markerFingerprint(program rule.Program, _ any) [sha256.Size]byte {
	return sha256.Sum256([]byte(markerText(program)))
}

// runsOf is how many files a markerRule was run on, the marker excluded.
func runsOf(runs *sync.Map) int {
	total := 0
	runs.Range(func(_, count any) bool {
		total += *count.(*int)
		return true
	})
	return total
}

// A derived rule replays on every file whose bytes are unchanged while its program fingerprint is too, runs
// again on every file once the data its fingerprint covers changes, and always agrees with a walk with no
// cache (#kdee854). The control is a fingerprint that misses the data the rule reads: it replays the old marker
// after the marker changed, so the comparison with an uncached walk is shown able to fail.
func TestADerivedRuleReplaysWhileItsProgramFingerprintHolds(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name        string
		fingerprint func(rule.Program, any) [sha256.Size]byte
		wantStale   bool
	}{
		{"complete fingerprint", markerFingerprint, false},
		{"control: a fingerprint that misses the marker", func(rule.Program, any) [sha256.Size]byte { return [sha256.Size]byte{7} }, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			root := writeProject(t, map[string]string{
				"tsconfig.json": minimalConfig,
				"marker.ts":     "// red\n",
				"a.ts":          "export const a = 1;\n",
				"b.ts":          "export const b = 2;\n",
			})
			ctx := context.Background()
			key := program.HashRuleSet([]string{"derived"})
			var recorded *program.LintCache
			walk := func(cached bool) (program.Result, *sync.Map) {
				t.Helper()
				graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
				if err != nil {
					t.Fatal(err)
				}
				runs := &sync.Map{}
				rules := []rule.Rule{markerRule("test-derived", scenario.fingerprint, runs)}
				var reuse *program.FindingsReuse
				if cached {
					reuse = program.NewFindingsReuse(key, recorded, program.PathAnchor{})
					graph.FindingsReuse = reuse
				}
				result, err := graph.Walk(ctx, graph.ProjectFiles(), rules)
				if err != nil {
					t.Fatal(err)
				}
				if cached {
					recorded = roundTripLintCache(t, reuse.Recorded())
				}
				return result, runs
			}
			agrees := func(stage string, warm program.Result) bool {
				t.Helper()
				cold, _ := walk(false)
				return reflect.DeepEqual(diagnosticKeys(warm.Diagnostics), diagnosticKeys(cold.Diagnostics))
			}

			walk(true)
			warm, runs := walk(true)
			if runsOf(runs) != 0 || warm.FilesReplayed != 3 {
				t.Errorf("nothing changed, and the rule ran on %d files with %d replayed; want 0 and all 3", runsOf(runs), warm.FilesReplayed)
			}
			if !agrees("unchanged", warm) {
				t.Errorf("an unchanged tree's replay differs from an uncached walk")
			}

			// An edit outside what the fingerprint covers: the edited file runs, the other replays.
			if err := os.WriteFile(filepath.Join(root, "a.ts"), []byte("export const a = 1;\nexport const more = 3;\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			warm, runs = walk(true)
			if runsOf(runs) != 1 {
				t.Errorf("after an edit to a.ts the rule ran on %d files, want a.ts alone", runsOf(runs))
			}
			if !agrees("a.ts edited", warm) {
				t.Errorf("after an edit to a.ts the replay differs from an uncached walk")
			}

			// The data the rule reads changes, and no file it reports in does.
			if err := os.WriteFile(filepath.Join(root, "marker.ts"), []byte("// blue\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			warm, runs = walk(true)
			stale := !agrees("marker edited", warm)
			if stale != scenario.wantStale {
				t.Errorf("after the marker changed the replay is stale: %v, want %v (the rule ran on %d files)", stale, scenario.wantStale,
					runsOf(runs))
			}
			if !scenario.wantStale && runsOf(runs) != 2 {
				t.Errorf("after the marker changed the rule ran on %d files, want both", runsOf(runs))
			}
		})
	}
}

// A rule that reads other files without a program fingerprint is walked on every file, every run, as before:
// the default is the sound one. A fingerprint that panics leaves its rule unkeyed for the run, walked
// everywhere and recorded nowhere, and the walk goes on.
func TestARuleWithNoUsableProgramFingerprintIsWalkedEveryRun(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name        string
		fingerprint func(rule.Program, any) [sha256.Size]byte
	}{
		{"none declared", nil},
		{"one that panics", func(rule.Program, any) [sha256.Size]byte { panic("a fingerprint that fails") }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			root := writeProject(t, map[string]string{
				"tsconfig.json": minimalConfig,
				"marker.ts":     "// red\n",
				"a.ts":          "export const a = 1;\n",
			})
			ctx := context.Background()
			key := program.HashRuleSet([]string{"uncacheable"})
			var recorded *program.LintCache
			for run := 1; run <= 2; run++ {
				graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
				if err != nil {
					t.Fatal(err)
				}
				runs := &sync.Map{}
				reuse := program.NewFindingsReuse(key, recorded, program.PathAnchor{})
				graph.FindingsReuse = reuse
				result, err := graph.Walk(ctx, graph.ProjectFiles(), []rule.Rule{markerRule("test-reads-other-files", scenario.fingerprint, runs)})
				if err != nil {
					t.Fatal(err)
				}
				if runsOf(runs) != 1 {
					t.Errorf("run %d: the rule ran on %d files, want a.ts every run", run, runsOf(runs))
				}
				if len(result.Diagnostics) != 1 {
					t.Errorf("run %d: %d findings, want a.ts's one", run, len(result.Diagnostics))
				}
				recorded = roundTripLintCache(t, reuse.Recorded())
			}
		})
	}
}

// optionMarkerRule is a derived rule whose option names the marker file it reads: it reports every variable
// declaration with that file's text. fingerprint is its ProgramFingerprint. runs counts the files it was run
// on, by base name, the markers excluded.
func optionMarkerRule(fingerprint func(rule.Program, any) [sha256.Size]byte, runs *sync.Map) rule.Rule {
	return rule.Rule{
		Name:               "test-derived-options",
		ProgramReads:       rule.ReadsOtherFiles,
		ProgramFingerprint: fingerprint,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if strings.HasSuffix(filepath.Base(ctx.SourceFile.FileName().AsString()), "-marker.ts") {
				return nil
			}
			count, _ := runs.LoadOrStore(filepath.Base(ctx.SourceFile.FileName().AsString()), new(int))
			*count.(*int)++
			marker := namedMarkerText(ctx.Program, options.(string))
			return rule.Listeners{ast.KindVariableDeclaration: func(node *ast.Node) {
				ctx.ReportNode(node, rule.Message{Id: "marker", Description: "the marker says " + marker})
			}}
		},
	}
}

// namedMarkerText is the text of the file with that base name, trimmed, read through the program's file list.
func namedMarkerText(program rule.Program, name string) string {
	for _, sourceFile := range program.SourceFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == name {
			return strings.TrimSpace(sourceFile.Text())
		}
	}
	return ""
}

// A derived rule whose option chooses what it reads is fingerprinted under each selection's own options
// (#s9k38p3): files configured to read red-marker.ts replay while only blue-marker.ts changes, files
// configured to read blue-marker.ts run again, and every walk agrees with one with no cache. The control
// fingerprints every selection under the top-level option, as a fingerprint with no options in hand must, and
// replays the blue files' old marker after it changed, so the comparison is shown able to fail.
func TestADerivedRuleIsFingerprintedUnderEachSelectionsOptions(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name        string
		fingerprint func(rule.Program, any) [sha256.Size]byte
		wantStale   bool
	}{
		{"the options' marker", func(program rule.Program, options any) [sha256.Size]byte {
			return sha256.Sum256([]byte(namedMarkerText(program, options.(string))))
		}, false},
		{"control: the top-level option's marker, whatever the selection", func(program rule.Program, _ any) [sha256.Size]byte {
			return sha256.Sum256([]byte(namedMarkerText(program, "red-marker.ts")))
		}, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			root := writeProject(t, map[string]string{
				"tsconfig.json": `{"compilerOptions":{"target":"ES2022","module":"esnext","moduleResolution":"bundler","strict":true,"noEmit":true},` +
					`"include":["**/*.ts"]}`,
				"red-marker.ts":  "// red\n",
				"blue-marker.ts": "// blue\n",
				"reds/a.ts":      "export const a = 1;\n",
				"blues/b.ts":     "export const b = 2;\n",
			})
			setting := func(marker string) configuration.RuleSetting {
				return configuration.RuleSetting{Severity: configuration.SeverityError, Options: []json.RawMessage{json.RawMessage(`"` + marker + `"`)}}
			}
			config := &configuration.Config{
				Root:  root,
				Rules: map[string]configuration.RuleSetting{"test-derived-options": setting("red-marker.ts")},
				Overrides: []configuration.Override{{Files: []string{"blues/**"},
					Rules: map[string]configuration.RuleSetting{"test-derived-options": setting("blue-marker.ts")}}},
			}
			registry := configuration.OptionsRegistry{"test-derived-options": {Decode: func(raw json.RawMessage) (any, error) {
				var marker string
				err := json.Unmarshal(raw, &marker)
				return marker, err
			}}}
			ctx := context.Background()
			key := program.HashRuleSet([]string{"derived-options"})
			var recorded *program.LintCache
			walk := func(cached bool) (program.Result, *sync.Map) {
				t.Helper()
				// Four checkers, so the walk runs on several workers and each takes its selections' fingerprints at
				// once, which is the case -race has to see.
				graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), Checkers: 4})
				if err != nil {
					t.Fatal(err)
				}
				if graph.Workers() < 2 {
					t.Fatalf("the walk runs on %d worker, so no two fingerprints are ever taken at once", graph.Workers())
				}
				graph.LintConfig, graph.RuleOptions = config, registry
				runs := &sync.Map{}
				var reuse *program.FindingsReuse
				if cached {
					reuse = program.NewFindingsReuse(key, recorded, program.PathAnchor{})
					graph.FindingsReuse = reuse
				}
				result, err := graph.Walk(ctx, graph.ProjectFiles(), []rule.Rule{optionMarkerRule(scenario.fingerprint, runs)})
				if err != nil {
					t.Fatal(err)
				}
				if cached {
					recorded = roundTripLintCache(t, reuse.Recorded())
				}
				return result, runs
			}
			agrees := func(warm program.Result) bool {
				t.Helper()
				cold, _ := walk(false)
				if len(cold.Diagnostics) != 2 {
					t.Fatalf("an uncached walk reports %d findings, want a.ts's and b.ts's", len(cold.Diagnostics))
				}
				return reflect.DeepEqual(diagnosticKeys(warm.Diagnostics), diagnosticKeys(cold.Diagnostics))
			}
			ran := func(runs *sync.Map, name string) bool {
				_, found := runs.Load(name)
				return found
			}

			walk(true)
			warm, runs := walk(true)
			if runsOf(runs) != 0 || warm.FilesReplayed != 4 {
				t.Errorf("nothing changed, and the rule ran on %d files with %d replayed; want 0 and all 4", runsOf(runs), warm.FilesReplayed)
			}

			// Only the blue files' data changes: they run again, and the red files replay.
			if err := os.WriteFile(filepath.Join(root, "blue-marker.ts"), []byte("// navy\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			warm, runs = walk(true)
			stale := !agrees(warm)
			if stale != scenario.wantStale {
				t.Errorf("after the blue marker changed the replay is stale: %v, want %v (the rule ran on %d files)", stale, scenario.wantStale,
					runsOf(runs))
			}
			if !scenario.wantStale && (!ran(runs, "b.ts") || ran(runs, "a.ts")) {
				t.Errorf("after the blue marker changed the rule ran on b.ts: %v and a.ts: %v, want b.ts alone", ran(runs, "b.ts"), ran(runs, "a.ts"))
			}

			// And the red files' data: they run again, and the blue files replay.
			if err := os.WriteFile(filepath.Join(root, "red-marker.ts"), []byte("// crimson\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			warm, runs = walk(true)
			if !scenario.wantStale && (!agrees(warm) || !ran(runs, "a.ts") || ran(runs, "b.ts")) {
				t.Errorf("after the red marker changed the rule ran on a.ts: %v and b.ts: %v, want a.ts alone, agreeing with an uncached walk",
					ran(runs, "a.ts"), ran(runs, "b.ts"))
			}
		})
	}
}

// A walk computes each derived rule's program fingerprint once for each set of options it gives the rule,
// however many selections and workers ask (#s9k38p3). Six derived rules, no options, over three resolutions on
// several workers: six computations. One per selection on each worker was 330 on an ahra edit run instead of
// 6, about 1.85 GB of allocation (#f96cnry's profile).
func TestAWalkComputesEachProgramFingerprintOnce(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"tsconfig.json": `{"compilerOptions":{"target":"ES2022","module":"esnext","moduleResolution":"bundler","strict":true,"noEmit":true},` +
			`"include":["**/*.ts"]}`,
	}
	for _, directory := range []string{"a", "b", "c"} {
		for _, name := range []string{"one", "two", "three"} {
			files[directory+"/"+name+".ts"] = "export const " + name + " = 1;\n"
		}
	}
	root := writeProject(t, files)
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), Checkers: 4})
	if err != nil {
		t.Fatal(err)
	}
	if graph.Workers() < 2 {
		t.Fatalf("the walk runs on %d worker, so no two workers ever ask for one fingerprint", graph.Workers())
	}

	var computations atomic.Int32
	var rules []rule.Rule
	settings := map[string]configuration.RuleSetting{}
	for index := range 6 {
		name := "test-derived-" + string(rune('a'+index))
		rules = append(rules, rule.Rule{
			Name:         name,
			ProgramReads: rule.ReadsOtherFiles,
			ProgramFingerprint: func(rule.Program, any) [sha256.Size]byte {
				computations.Add(1)
				return sha256.Sum256([]byte(name))
			},
			Run: func(ctx rule.Context, options any) rule.Listeners { return nil },
		})
		settings[name] = configuration.RuleSetting{Severity: configuration.SeverityError}
	}
	// Two overrides that change nothing but which resolution a file lands in, so the walk makes three
	// selections on each worker that sees all three directories.
	graph.LintConfig = &configuration.Config{Root: root, Rules: settings, Overrides: []configuration.Override{
		{Files: []string{"b/**"}, Rules: map[string]configuration.RuleSetting{"test-derived-a": {Severity: configuration.SeverityError}}},
		{Files: []string{"c/**"}, Rules: map[string]configuration.RuleSetting{"test-derived-b": {Severity: configuration.SeverityError}}},
	}}
	graph.FindingsReuse = program.NewFindingsReuse(program.HashRuleSet([]string{"computed-once"}), nil, program.PathAnchor{})
	if _, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules); err != nil {
		t.Fatal(err)
	}
	if computations.Load() != 6 {
		t.Fatalf("a walk of six derived rules over three resolutions on %d workers computed %d program fingerprints, want 6",
			graph.Workers(), computations.Load())
	}
}
