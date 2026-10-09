package program_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/incremental"
	"github.com/system-inc/cohere/internal/types/program"
)

// incrementalConfig turns on the two options the incremental machinery needs. Everything else
// matches minimalConfig, so a difference in findings is a difference the machinery caused
// rather than a difference in how the program was configured.
const incrementalConfig = `{
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "strict": true,
        "noEmit": true,
        "incremental": true,
        "tsBuildInfoFile": "./tsconfig.tsbuildinfo"
    },
    "include": ["./*.ts"]
}`

// messagesOf renders diagnostics into a comparable form. Comparing pointers would compare
// identity rather than content, and two runs never share pointers.
//
// The identity is file, code, and position rather than message text: MessageText is empty for
// a compiler diagnostic (the text is assembled at render time from a key and args), so keying
// on it would make every diagnostic compare equal and every comparison here vacuous.
func messagesOf(diagnostics []*ast.Diagnostic) []string {
	rendered := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		name := ""
		if file := diagnostic.File(); file != nil {
			name = filepath.Base(file.FileName().AsString())
		}
		rendered = append(rendered, fmt.Sprintf("%s:%d:TS%d", name, diagnostic.Pos(), diagnostic.Code()))
	}
	return rendered
}

// TestIncrementalProgramSkipsUnchangedFiles is the control, and it is deliberately not a test
// of the incremental path. TestIncrementalDiagnosticsSurvivesAChange below is that.
//
// This one never writes a build info, so nothing here is served from a cache: every leg runs
// the full pass, and leg one's assertion that ReadBuildInfoProgram returns nil holds on every
// run rather than only the first. That is the point. It establishes what the plant-and-heal
// harness reports with no cache involved, which is the baseline the cached path has to
// reproduce exactly.
//
// Keeping a no-cache control matters more here than it usually would: the failure this domain
// exists to prevent is a cache that reports a clean tree forever, and a suite where every test
// runs through the cache cannot tell that from a genuinely clean fixture.
func TestIncrementalProgramSkipsUnchangedFiles(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": incrementalConfig,
		"clean.ts":      "export const value: number = 1;\n",
	})

	build := func() *program.Graph {
		t.Helper()
		graph, err := program.Build(program.Options{
			ConfigFileName:   "tsconfig.json",
			CurrentDirectory: directory,
		})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		return graph
	}

	// The host has to be the one the program was built with. The build info's paths resolve
	// relative to it, so a fresh host would resolve them elsewhere and report everything as
	// changed, which looks like a cold cache rather than like a bug.
	readBuildInfo := func(graph *program.Graph) *incremental.Program {
		if graph.CompilerHost == nil {
			t.Fatal("Graph.CompilerHost is nil; the incremental path cannot be reached without it")
		}
		reader := incremental.NewBuildInfoReader(graph.CompilerHost)
		return incremental.ReadBuildInfoProgram(graph.Config, reader, graph.CompilerHost)
	}

	// Leg one: cold. No build info exists, so ReadBuildInfoProgram must return nil, and that
	// nil means rebuild-everything rather than an error.
	//
	// Nothing in this test writes one, so this also holds on every later leg, which is what
	// makes the rest a no-cache baseline.
	cold := build()
	if previous := readBuildInfo(cold); previous != nil {
		t.Fatal("a build info was found before any run wrote one; the temporary project is dirty " +
			"and every later leg would be measured against the wrong state")
	}
	coldFindings := messagesOf(cold.AllDiagnostics(context.Background()))

	// The clean fixture must produce nothing, or the later legs cannot tell a real finding from
	// pre-existing noise.
	if len(coldFindings) != 0 {
		t.Fatalf("the clean fixture produced %d findings, so this test cannot distinguish a "+
			"planted error from existing ones: %v", len(coldFindings), coldFindings)
	}

	// Leg two: plant an error and confirm a fresh build finds it. This proves the fixture and
	// the harness can produce a finding at all, which is what makes leg three's result mean
	// something. Without it, a silent leg three is indistinguishable from a broken detector.
	plantedPath := filepath.Join(directory, "planted.ts")
	if err := os.WriteFile(plantedPath, []byte("export const broken: number = \"a string\";\n"), 0o644); err != nil {
		t.Fatalf("planting: %v", err)
	}

	planted := build()
	plantedFindings := messagesOf(planted.AllDiagnostics(context.Background()))
	if len(plantedFindings) == 0 {
		t.Fatal("a deliberate type error produced no findings; this harness cannot detect a " +
			"regression and a clean result from it would be vacuous")
	}
	t.Logf("control: planted error found, %d finding(s): %v", len(plantedFindings), plantedFindings)

	// Leg three: remove the error and confirm the findings clear. A cache that never
	// invalidates would keep reporting the planted error here, and a cache that reports
	// nothing regardless would have failed leg two.
	if err := os.Remove(plantedPath); err != nil {
		t.Fatalf("removing the planted file: %v", err)
	}

	healed := build()
	healedFindings := messagesOf(healed.AllDiagnostics(context.Background()))
	if len(healedFindings) != 0 {
		t.Fatalf("after removing the planted error the findings did not clear, which is a stale "+
			"result rather than a clean tree: %v", healedFindings)
	}
}

// TestBuildInfoReaderReturnsNilBeforeAnyBuild pins the contract a caller is most likely to get
// wrong: nil is "no previous state", not "an error occurred". A caller that treats it as a
// failure reports an error on every first run.
func TestBuildInfoReaderReturnsNilBeforeAnyBuild(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": incrementalConfig,
		"main.ts":       "export const value: number = 1;\n",
	})

	graph, err := program.Build(program.Options{
		ConfigFileName:   "tsconfig.json",
		CurrentDirectory: directory,
	})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	if graph.CompilerHost == nil {
		t.Fatal("Graph.CompilerHost is nil; the incremental path cannot be reached without it")
	}

	reader := incremental.NewBuildInfoReader(graph.CompilerHost)
	if previous := incremental.ReadBuildInfoProgram(graph.Config, reader, graph.CompilerHost); previous != nil {
		t.Fatal("ReadBuildInfoProgram returned a program with no build info on disk")
	}
}

// TestIncrementalDiagnosticsSurvivesAChange is the real test the control above is a control
// for, and the only one that can tell a working cache from a permanently-stale one.
//
// Four legs, and leg three is the one that matters:
//
//	cold          no build info, full pass, clean tree
//	write         a build info lands on disk
//	warm dirty    a violation planted AFTER the write is still found
//	warm clean    removing it clears the findings again
//
// A cache that never invalidates passes cold, write, and warm-clean, and fails warm-dirty.
// That is precisely why cold-equals-warm is not the shape used: a stale cache reporting
// yesterday's clean tree is the fastest run we would ever measure.
func TestIncrementalDiagnosticsSurvivesAChange(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": incrementalConfig,
		"clean.ts":      "export const value: number = 1;\n",
	})

	build := func() *program.Graph {
		t.Helper()
		graph, err := program.Build(program.Options{
			ConfigFileName:   "tsconfig.json",
			CurrentDirectory: directory,
		})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		return graph
	}

	ctx := context.Background()

	// Leg one: cold. No build info yet, so this is the full pass.
	cold := build()
	if findings := messagesOf(cold.IncrementalDiagnostics(ctx)); len(findings) != 0 {
		t.Fatalf("the clean fixture produced %d findings, so no later leg can distinguish a "+
			"planted error from existing noise: %v", len(findings), findings)
	}

	// Leg two: write the build info, and confirm it actually reached disk. Skipping this
	// check would let every later leg run cold while appearing to test a cache.
	coldSession := cold.NewIncrementalSession()
	if coldSession == nil {
		t.Fatal("NewIncrementalSession returned nil on an incremental config")
	}
	coldSession.Diagnostics(ctx)
	if diagnostics := coldSession.Write(ctx); len(diagnostics) != 0 {
		t.Fatalf("writing the build info reported %d diagnostics: %v",
			len(diagnostics), messagesOf(diagnostics))
	}
	buildInfoPath := filepath.Join(directory, "tsconfig.tsbuildinfo")
	written, err := os.Stat(buildInfoPath)
	if err != nil {
		t.Fatalf("no build info on disk after WriteBuildInfo, so every warm leg below would "+
			"silently run cold: %v", err)
	}
	if written.Size() == 0 {
		t.Fatal("the build info is empty, which reads as a successful write and behaves as no cache")
	}
	t.Logf("build info written, %d bytes", written.Size())

	// Leg three: plant a violation AFTER the build info was written, then run warm. This is
	// the case a stale cache fails. The file is new, so the cache has never seen it, and a
	// cache that trusts its snapshot over the filesystem reports nothing here.
	plantedPath := filepath.Join(directory, "planted.ts")
	if err := os.WriteFile(plantedPath, []byte("export const broken: number = \"a string\";\n"), 0o644); err != nil {
		t.Fatalf("planting: %v", err)
	}

	warmDirty := build()
	dirtyFindings := messagesOf(warmDirty.IncrementalDiagnostics(ctx))
	if len(dirtyFindings) == 0 {
		t.Fatal("a violation planted after the build info was written was NOT found on a warm " +
			"run. That is a stale cache reporting a clean tree, which is the exact failure this " +
			"whole domain exists to prevent")
	}
	t.Logf("warm run found the planted violation: %v", dirtyFindings)

	// Leg four: remove it and confirm the findings clear rather than persisting.
	if err := os.Remove(plantedPath); err != nil {
		t.Fatalf("removing the planted file: %v", err)
	}
	rewriteSession := warmDirty.NewIncrementalSession()
	if rewriteSession == nil {
		t.Fatal("NewIncrementalSession returned nil on the rewrite")
	}
	rewriteSession.Diagnostics(ctx)
	if diagnostics := rewriteSession.Write(ctx); len(diagnostics) != 0 {
		t.Fatalf("rewriting the build info reported diagnostics: %v", messagesOf(diagnostics))
	}

	warmClean := build()
	if findings := messagesOf(warmClean.IncrementalDiagnostics(ctx)); len(findings) != 0 {
		t.Fatalf("after removing the planted error the warm run still reports findings, which "+
			"is a stale result rather than a clean tree: %v", findings)
	}
}

// TestIncrementalMatchesFullPass pins the equivalence the saving is only worth having if it
// holds: the incremental path must produce the same findings as the full one, not merely
// fewer of them faster.
func TestIncrementalMatchesFullPass(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{
		"tsconfig.json": incrementalConfig,
		"clean.ts":      "export const value: number = 1;\n",
		"broken.ts":     "export const wrong: number = \"text\";\nexport const alsoWrong: string = 5;\n",
	})

	build := func() *program.Graph {
		t.Helper()
		graph, err := program.Build(program.Options{
			ConfigFileName:   "tsconfig.json",
			CurrentDirectory: directory,
		})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		return graph
	}

	ctx := context.Background()

	first := build()
	full := messagesOf(first.AllDiagnostics(ctx))
	if len(full) < 2 {
		t.Fatalf("the fixture was meant to produce at least two findings and produced %d; "+
			"an equivalence test over an empty set proves nothing: %v", len(full), full)
	}
	firstSession := first.NewIncrementalSession()
	if firstSession == nil {
		t.Fatal("NewIncrementalSession returned nil on an incremental config")
	}
	firstSession.Diagnostics(ctx)
	if diagnostics := firstSession.Write(ctx); len(diagnostics) != 0 {
		t.Fatalf("writing the build info reported diagnostics: %v", messagesOf(diagnostics))
	}

	second := build()
	warm := messagesOf(second.IncrementalDiagnostics(ctx))

	slices.Sort(full)
	slices.Sort(warm)
	if !slices.Equal(full, warm) {
		t.Fatalf("the incremental path disagrees with the full pass.\n full: %v\n warm: %v", full, warm)
	}
	t.Logf("equivalence holds across %d findings", len(full))
}
