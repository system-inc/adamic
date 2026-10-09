package unused_exports

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// fixturePath resolves a fixture under testdata to an absolute path.
//
// Absolute rather than relative because `program.Build` resolves a tsconfig against a current
// directory, and a test binary's working directory is its own package directory. Hardcoding a
// developer's scratch path here would pass on one machine and fail everywhere else, which is the
// shape of green that means nothing.
func fixturePath(t *testing.T, name string) string {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("resolving the %s fixture: %v", name, err)
	}
	return absolute
}

// The island fixture holds an abandoned export, the private world only it reaches, a live export
// reached from a rooted file, and a function referenced only by itself. Four shapes, one program, so
// the firing half and the declining half are answered by the same run.
func runFixture(t *testing.T, deep bool) *Report {
	t.Helper()
	islandFixture := fixturePath(t, "island")
	graph, err := program.Build(program.Options{
		ConfigFileName:   islandFixture + "/tsconfig.json",
		CurrentDirectory: islandFixture,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the fixture program: %v", err)
	}
	report, err := Run(context.Background(), graph, graph.ProjectFiles(), deep)
	if err != nil {
		t.Fatalf("running the report: %v", err)
	}
	return report
}

// describeIslands renders the clusters so a failure says what was found rather than only that a
// count was wrong.
func describeIslands(report *Report) string {
	var lines []string
	for _, island := range report.Islands {
		var names []string
		for _, member := range island.Members {
			names = append(names, member.Name)
		}
		sort.Strings(names)
		lines = append(lines, fmt.Sprintf("[%s]", strings.Join(names, " ")))
	}
	return strings.Join(lines, " ")
}

// TestClosureFindsTheIsland is the firing half. The planted cluster is one abandoned export and the
// two private declarations that existed only to serve it, and the whole point of the closure is
// that it reports them as ONE cluster rather than as an export the flat view can see and two
// private declarations the flat view cannot see at all.
func TestClosureFindsTheIsland(t *testing.T) {
	t.Parallel()
	report := runFixture(t, true)

	var cluster *Island
	for index := range report.Islands {
		for _, member := range report.Islands[index].Members {
			if member.Name == "ScoreWindow" {
				cluster = &report.Islands[index]
			}
		}
	}
	if cluster == nil {
		t.Fatalf("ScoreWindow was not reported dead, so the closure found nothing: %s", describeIslands(report))
	}

	want := map[string]bool{"ScoreWindow": true, "normalizeDecay": true, "DecayTable": true}
	got := map[string]bool{}
	for _, member := range cluster.Members {
		got[member.Name] = true
	}
	for name := range want {
		if !got[name] {
			t.Fatalf("%s was not in the cluster, so the private world did not travel with the export: %s",
				name, describeIslands(report))
		}
	}
	if len(cluster.Members) != 3 {
		t.Fatalf("the cluster held %d members, want 3: %s", len(cluster.Members), describeIslands(report))
	}
	// The export leads, because it is the name the reader recognizes.
	if !cluster.Members[0].Exported || cluster.Members[0].Name != "ScoreWindow" {
		t.Fatalf("the cluster did not lead with its export: %+v", cluster.Members[0])
	}
	if cluster.Exported != 1 {
		t.Fatalf("the cluster counted %d exports, want 1", cluster.Exported)
	}
}

// TestClosureSparesLiveCode is the declining half. LiveHelper is reached only from a Next.js
// `page.tsx`, which nothing imports and the root set spares, so a closure that does not seed the
// rooted files' declarations as roots reports it dead. That is the cascade this whole feature has
// to not do.
func TestClosureSparesLiveCode(t *testing.T) {
	t.Parallel()
	report := runFixture(t, true)
	for _, island := range report.Islands {
		for _, member := range island.Members {
			if member.Name == "LiveHelper" {
				t.Fatalf("LiveHelper was reported dead, so the root set did not seed the closure: %s",
					describeIslands(report))
			}
			if member.Name == "Page" {
				t.Fatalf("the rooted page itself was reported dead: %s", describeIslands(report))
			}
		}
	}
}

// TestClosureReportsSelfReferenceAsDead pins the judgment call. A recursive function nothing else
// calls references only itself, so the flat set counts that self-reference as a use and calls it
// alive. The closure does not, because nothing marks it first. This is the clearest case where the
// two answers differ, and it is written down as a test so a future no-unused-vars port that
// disagrees disagrees loudly.
func TestClosureReportsSelfReferenceAsDead(t *testing.T) {
	t.Parallel()
	flat := runFixture(t, false)
	for _, finding := range flat.Unreferenced {
		if finding.Name == "recursiveOrphan" {
			t.Fatalf("the flat view reported recursiveOrphan, so the self-reference no longer counts as a use there")
		}
	}

	deep := runFixture(t, true)
	found := false
	for _, island := range deep.Islands {
		for _, member := range island.Members {
			if member.Name == "recursiveOrphan" {
				found = true
				if len(island.Members) != 1 {
					t.Fatalf("recursiveOrphan formed a %d-member island; a self-edge must not make a cluster: %s",
						len(island.Members), describeIslands(deep))
				}
			}
		}
	}
	if !found {
		t.Fatalf("the closure did not report recursiveOrphan, so a function only calling itself reads as alive: %s",
			describeIslands(deep))
	}
}

// TestClosureIsOptIn proves the flat report is unchanged when the closure is not asked for, so a
// reader comparing the two numbers is comparing the same flat analysis against itself.
func TestClosureIsOptIn(t *testing.T) {
	t.Parallel()
	flat := runFixture(t, false)
	if len(flat.Islands) != 0 {
		t.Fatalf("islands were computed without being asked for: %d", len(flat.Islands))
	}
	if flat.DeclarationsAnalyzed != 0 {
		t.Fatalf("the closure inventory was built without being asked for: %d", flat.DeclarationsAnalyzed)
	}

	deep := runFixture(t, true)
	if len(deep.Unreferenced) != len(flat.Unreferenced) || len(deep.Files) != len(flat.Files) {
		t.Fatalf("asking for the closure changed the flat report: flat %d/%d, deep %d/%d",
			len(flat.Unreferenced), len(flat.Files), len(deep.Unreferenced), len(deep.Files))
	}
	if deep.DeclarationsAnalyzed == 0 {
		t.Fatalf("the closure ran with an empty inventory, which reports everything dead for free")
	}
}

// TestReExportIsAPassThrough pins the judgment call that was assumed wrong before it was measured.
//
// A re-export does NOT mark its target referenced, so an export reachable only through a barrel
// nobody imports is reported dead. The mechanism is that `isDeclarationName` classifies an
// ExportSpecifier's identifier as a declaring name, so the walk never records it as a use.
//
// This is pinned rather than left to the code because it is the single answer most likely to be
// changed by someone who reads it as a bug, and changing it makes every barrel an unconditional
// root — which protects everything behind it and makes the report go quiet. On the real tree this
// rule is what correctly reports four abandoned payment icons whose only mention anywhere is a
// barrel nothing imports.
func TestReExportIsAPassThrough(t *testing.T) {
	t.Parallel()
	// One export reachable only through a re-export, and a barrel nothing imports.
	barrelFixture := fixturePath(t, "barrel")
	graph, err := program.Build(program.Options{
		ConfigFileName:   barrelFixture + "/tsconfig.json",
		CurrentDirectory: barrelFixture,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the barrel fixture: %v", err)
	}
	report, err := Run(context.Background(), graph, graph.ProjectFiles(), true)
	if err != nil {
		t.Fatalf("running the report: %v", err)
	}

	found := false
	for _, file := range report.Files {
		if strings.HasSuffix(file.FileName, "thing.ts") {
			found = true
		}
	}
	if !found {
		t.Fatalf("BarrelOnlyThing was spared, so a re-export now counts as a use and every unimported barrel is a root")
	}

	// The barrel itself contributes no exports, which is the gap this documents rather than fixes.
	if report.ExportsAnalyzed != 1 {
		t.Fatalf("analyzed %d exports, want 1: a barrel's `export ... from` is not a declaration statement",
			report.ExportsAnalyzed)
	}
}

// TestDeadCodeReferenceStillCountsAsAUse contradicts a JUDGMENT claim in `roots.go`.
//
// The claim reads: "A reference from inside code that is ITSELF dead does not count, and this is
// what the closure buys. In the flat set it counts, which is why the flat set cannot see an island."
//
// The closure does not do this. It has no reachability input at all -- `closure.go` contains no
// mention of reachability, and `Unreachable` is a report field that is collected and printed and
// never read back into the reference walk. `roots.go`'s own header says the two questions are
// "separable on purpose", which is exactly right about the design and is why the JUDGMENT entry
// describing the closure as consuming one of them cannot be true.
//
// The fixture is built so reachability is the ONLY difference. Both constants are imported by the
// same file, each referenced exactly once, from two exported functions that a framework-spared page
// calls. If the claim held, `ReachedOnlyAfterReturn` would be reported and `ReachedNormally` spared.
// Both are spared.
//
// Written as a test asserting what the code DOES rather than what the comment says it does, and the
// comment is corrected alongside it. A doc comment making a falsifiable claim needs a test or it
// needs deleting -- this one survived being labelled considered judgment for exactly as long as
// nothing checked it.
func TestDeadCodeReferenceStillCountsAsAUse(t *testing.T) {
	t.Parallel()
	deadrefFixture := fixturePath(t, "deadref")
	graph, err := program.Build(program.Options{
		ConfigFileName:   deadrefFixture + "/tsconfig.json",
		CurrentDirectory: deadrefFixture,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the deadref fixture: %v", err)
	}
	report, err := Run(context.Background(), graph, graph.ProjectFiles(), true)
	if err != nil {
		t.Fatalf("running the report: %v", err)
	}

	reported := map[string]bool{}
	for _, island := range report.Islands {
		for _, declaration := range island.Members {
			reported[declaration.Name] = true
		}
	}
	for _, file := range report.Files {
		if strings.HasSuffix(file.FileName, "icons.ts") {
			t.Error("icons.ts was reported as wholly unreferenced; both of its exports are " +
				"referenced once, so this fixture is not testing what it believes it is")
		}
	}

	if reported["ReachedNormally"] {
		t.Error("ReachedNormally was reported dead; it is referenced from a live path reached by a " +
			"spared root, so the fixture's control is broken rather than the claim being confirmed")
	}
	if reported["ReachedOnlyAfterReturn"] {
		t.Error("ReachedOnlyAfterReturn was reported dead, so the closure now discounts references " +
			"from unreachable code. That would make the JUDGMENT entry in roots.go true and this " +
			"test is what should be updated -- but note that discounting is a strictly stronger " +
			"claim than reference analysis, and a wrong answer cascades through the whole closure")
	}

	// The unreachable statement is still reported, by the separate analysis that owns that question.
	// Asserted here so the two cannot be conflated: the reference survives AND the statement is
	// named, which is the design `roots.go`'s header describes.
	unreachableInUses := 0
	for _, unreachable := range report.Unreachable {
		if strings.HasSuffix(unreachable.FileName, "uses.ts") {
			unreachableInUses++
		}
	}
	if unreachableInUses != 1 {
		t.Errorf("unreachable statements reported in uses.ts = %d, want 1; the reachability half "+
			"of this fixture is what makes the reference half meaningful", unreachableInUses)
	}
}
