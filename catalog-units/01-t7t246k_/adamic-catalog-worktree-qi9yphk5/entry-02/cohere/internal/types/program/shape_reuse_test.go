package program_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// countingRule wraps a rule so a test can see which files it actually ran on, which is the only way to
// tell a replayed finding from a recomputed one that happens to be equal.
type countingRule struct {
	mutex sync.Mutex
	ran   map[string]int
}

func (counter *countingRule) wrap(subject rule.Rule) rule.Rule {
	run := subject.Run
	subject.Run = func(ctx rule.Context, options any) rule.Listeners {
		counter.mutex.Lock()
		counter.ran[filepath.Base(ctx.SourceFile.FileName().AsString())]++
		counter.mutex.Unlock()
		return run(ctx, options)
	}
	return subject
}

// snapshot is which files the rule ran on so far. Taken right after the walk under test, before an
// uncached walk for the truth runs every rule again and counts too.
func (counter *countingRule) snapshot() map[string]int {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()
	copied := make(map[string]int, len(counter.ran))
	for name, count := range counter.ran {
		copied[name] = count
	}
	return copied
}

// shapeWalk walks a freshly built graph with shapes computed against the previous run's, as the command
// does, and returns the result, what it recorded, and the shapes it used.
func shapeWalk(t *testing.T, root string, rules []rule.Rule, previous *program.LintCache,
	previousShapes map[string]program.SignatureEntry) (program.Result, *program.LintCache, map[string]program.SignatureEntry) {
	t.Helper()
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// A first walk has its signatures computed, as a project with build info does, so the tests below are
	// about shapes rather than about a first run's content-keyed fallback.
	if previousShapes == nil {
		previousShapes = program.RecordedRealSignatures(graph)
	}
	shapes, _ := graph.Signatures(context.Background(), previousShapes)
	graph.Shapes = shapes
	reuse := program.NewFindingsReuse(program.HashRuleSet([]string{"fixture"}), previous, program.PathAnchor{})
	graph.FindingsReuse = reuse
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return result, reuse.Recorded(), shapes
}

// shapeFixture has consumer.ts negate a value whose type comes from a function in lib.ts, so the
// type-aware rule's finding on consumer depends on lib's declaration and not on lib's body.
func shapeFixture(t *testing.T) string {
	return writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"lib.ts":        "export function value(): string {\n  return \"1\";\n}\n",
		"consumer.ts":   "import { value } from \"./lib\";\nexport const negated = -value();\n",
	})
}

// The point of the whole table: a body edit to a dependency re-runs a shape-keyed rule on the edited file
// alone, while the same rule keyed on contents re-runs on every importer too. Both must still report
// exactly what an uncached walk reports.
func TestABodyEditReplaysShapeKeyedFindingsOnImporters(t *testing.T) {
	t.Parallel()
	for name, reach := range map[string]rule.TypeReach{"shape-keyed": rule.TypeReachShapes, "content-keyed": rule.TypeReachContents} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := shapeFixture(t)
			counter := &countingRule{ran: map[string]int{}}
			typeAware := typeAwareRule(t)
			typeAware.TypeReach = reach
			rules := append(findingsReuseRules(t), counter.wrap(typeAware))

			before, recorded, shapes := shapeWalk(t, root, rules, nil, nil)
			if len(before.Diagnostics) == 0 {
				t.Fatal("consumer.ts produced no finding, so a replay of it would prove nothing")
			}

			if err := os.WriteFile(filepath.Join(root, "lib.ts"), []byte("export function value(): string {\n  return \"1\" + \"\";\n}\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			counter.ran = map[string]int{}
			after, _, _ := shapeWalk(t, root, rules, recorded, shapes)
			ran := counter.snapshot()
			truth := plainWalk(t, root, rules)

			if !reflect.DeepEqual(diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
				t.Fatalf("findings after a body edit differ from an uncached walk:\n cached %v\n truth  %v",
					diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics))
			}
			if !reflect.DeepEqual(foldedCoverage(after.Coverage), foldedCoverage(truth.Coverage)) {
				t.Errorf("coverage after a body edit differs from an uncached walk")
			}
			ranOnConsumer := ran["consumer.ts"] > 0
			if reach == rule.TypeReachShapes && ranOnConsumer {
				t.Error("a shape-keyed rule re-ran on an importer after an edit inside a body, so nothing was gained")
			}
			if reach == rule.TypeReachContents && !ranOnConsumer {
				t.Error("a contents-keyed rule replayed on an importer after its dependency's bytes changed")
			}
			if ran["lib.ts"] == 0 {
				t.Error("the edited file itself was not re-walked")
			}
			// The run says so too: consumer.ts replayed, and the count of files whose shape-keyed rules
			// ran again is what the lint line prints.
			if reach == rule.TypeReachShapes && after.ShapeKeyedRerun != 0 {
				t.Errorf("ShapeKeyedRerun = %d after a body edit, want 0", after.ShapeKeyedRerun)
			}
			if reach == rule.TypeReachContents && after.TypeAwareRerun == 0 {
				t.Error("TypeAwareRerun = 0, though the content-keyed rule ran again on the importer")
			}
		})
	}
}

// An edit to what a dependency exports reaches a shape-keyed rule's findings on every importer: here
// value() starts returning a number, and the finding on consumer goes away.
func TestAnExportEditReachesShapeKeyedFindingsOnImporters(t *testing.T) {
	t.Parallel()
	root := shapeFixture(t)
	counter := &countingRule{ran: map[string]int{}}
	typeAware := typeAwareRule(t)
	typeAware.TypeReach = rule.TypeReachShapes
	rules := append(findingsReuseRules(t), counter.wrap(typeAware))

	before, recorded, shapes := shapeWalk(t, root, rules, nil, nil)
	if len(before.Diagnostics) == 0 {
		t.Fatal("consumer.ts produced no finding before the edit")
	}
	if err := os.WriteFile(filepath.Join(root, "lib.ts"), []byte("export function value(): number {\n  return 1;\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	counter.ran = map[string]int{}
	after, refreshed, refreshedShapes := shapeWalk(t, root, rules, recorded, shapes)
	ran := counter.snapshot()
	truth := plainWalk(t, root, rules)

	if len(truth.Diagnostics) != 0 {
		t.Fatalf("the edit did not remove the finding even uncached, so the fixture proves nothing: %v", diagnosticKeys(truth.Diagnostics))
	}
	if !reflect.DeepEqual(diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
		t.Errorf("a shape-keyed finding was replayed over an export edit to its dependency:\n cached %v\n truth  %v",
			diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics))
	}
	if ran["consumer.ts"] == 0 {
		t.Error("the shape-keyed rule did not re-run on the importer of a changed export")
	}
	if after.ShapeKeyedRerun != 1 {
		t.Errorf("ShapeKeyedRerun = %d after an export edit, want 1 (consumer.ts)", after.ShapeKeyedRerun)
	}

	// The run after, with nothing changed, reads the refreshed entry: a refresh that kept the old finding
	// would bring it back here.
	again, _, _ := shapeWalk(t, root, rules, refreshed, refreshedShapes)
	if !reflect.DeepEqual(diagnosticKeys(again.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
		t.Errorf("the run after the refresh replayed a finding the edit removed: %v", diagnosticKeys(again.Diagnostics))
	}
}

// Without shapes the walk keys shape-keyed rules on the type fingerprint, which can only re-run more.
func TestShapeKeyedRulesWithoutShapesFallBackToContents(t *testing.T) {
	t.Parallel()
	root := shapeFixture(t)
	counter := &countingRule{ran: map[string]int{}}
	typeAware := typeAwareRule(t)
	typeAware.TypeReach = rule.TypeReachShapes
	rules := append(findingsReuseRules(t), counter.wrap(typeAware))

	_, recorded := walkAndRecord(t, root, rules, nil)
	if err := os.WriteFile(filepath.Join(root, "lib.ts"), []byte("export function value(): string {\n  return \"1\" + \"\";\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	counter.ran = map[string]int{}
	walkAndRecord(t, root, rules, recorded)
	if counter.ran["consumer.ts"] == 0 {
		t.Error("without shapes, a shape-keyed rule replayed on an importer after its dependency's bytes changed")
	}
}

// A comment or a blank line added to a dependency, anywhere outside a body, leaves its importers' shape-keyed
// findings replayed: no rule can read a plain comment off an imported declaration without reading the file's
// text, which keeps a rule on Contents (#zqsdzbq).
func TestACommentEditReplaysShapeKeyedFindingsOnImporters(t *testing.T) {
	t.Parallel()
	for name, text := range map[string]string{
		"appended":          "export function value(): string {\n  return \"1\";\n}\n// bench edit 1\n",
		"before the export": "// a note\n\nexport function value(): string {\n  return \"1\";\n}\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := shapeFixture(t)
			counter := &countingRule{ran: map[string]int{}}
			typeAware := typeAwareRule(t)
			typeAware.TypeReach = rule.TypeReachShapes
			rules := append(findingsReuseRules(t), counter.wrap(typeAware))

			before, recorded, shapes := shapeWalk(t, root, rules, nil, nil)
			if len(before.Diagnostics) == 0 {
				t.Fatal("consumer.ts produced no finding, so a replay of it would prove nothing")
			}
			if err := os.WriteFile(filepath.Join(root, "lib.ts"), []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			counter.ran = map[string]int{}
			after, _, _ := shapeWalk(t, root, rules, recorded, shapes)
			ran := counter.snapshot()
			truth := plainWalk(t, root, rules)
			if !reflect.DeepEqual(diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics)) {
				t.Fatalf("findings after a comment edit differ from an uncached walk:\n cached %v\n truth  %v",
					diagnosticKeys(after.Diagnostics), diagnosticKeys(truth.Diagnostics))
			}
			if ran["consumer.ts"] > 0 || after.ShapeKeyedRerun != 0 {
				t.Errorf("a comment edit to lib.ts re-ran the shape-keyed rule on its importer (ShapeKeyedRerun %d)", after.ShapeKeyedRerun)
			}
			if ran["lib.ts"] == 0 {
				t.Error("the edited file itself was not re-walked")
			}
		})
	}
}
