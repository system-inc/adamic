package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/types/program"
)

// buildNarrowFixtureGraph writes a small program to disk and builds a real type graph over it.
//
// A real graph rather than a stub, because what is under test is whether the scope, the closure and
// the file slice agree with each other. A hand-built graph would let all three be wrong together.
func buildNarrowFixtureGraph(t *testing.T, files map[string]string) (*program.Graph, string) {
	t.Helper()
	directory := t.TempDir()

	for name, contents := range files {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	configuration := `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler"},"include":["**/*.ts"]}`
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
		t.Fatal(err)
	}

	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building the fixture graph: %v", err)
	}
	return graph, directory
}

func narrowedNames(files []*ast.SourceFile) string {
	names := make([]string, 0, len(files))
	for _, sourceFile := range files {
		names = append(names, filepath.Base(sourceFile.FileName().AsString()))
	}
	return strings.Join(names, " ")
}

// TestNarrowToClosureAddsConsumersAndCountsThem holds the judgment both scoped paths share.
//
// `--changed` and a named path are different questions with the same answer, and the answer is the
// part worth getting wrong only once: the files a caller pointed at are not sufficient, because a
// change to an exported type breaks its consumers and those findings vanish rather than appear.
//
// Demonstrated on the real tree before this function existed: giving one function an `any` return
// produced seven new findings, four of them in a consumer two directories from the edit, and a run
// scoped to the edited file alone saw three.
func TestNarrowToClosureAddsConsumersAndCountsThem(t *testing.T) {
	t.Parallel()
	graph, directory := buildNarrowFixtureGraph(t, map[string]string{
		"Producer.ts":  "export function value(): number {\n  return 1;\n}\n",
		"Consumer.ts":  "import { value } from './Producer';\nexport const doubled = value() * 2;\n",
		"Unrelated.ts": "export const alone = 1;\n",
	})

	producer := filepath.Join(directory, "Producer.ts")
	scope := formatScope{
		FileNames:          []string{producer},
		index:              map[string]struct{}{producer: {}},
		RequestDescription: "Producer.ts",
	}

	narrowed, files := narrowToClosure(graph, scope, graph.ProjectFiles())

	names := narrowedNames(files)
	if !strings.Contains(names, "Consumer.ts") {
		t.Errorf("the consumer was not added to the scope: %s", names)
	}
	// The control. A closure that returned every project file would satisfy the assertion above, and
	// that is the shape the defect actually took when the reverse map had no edges.
	if strings.Contains(names, "Unrelated.ts") {
		t.Errorf("a file unrelated to the seed was in the scope: %s", names)
	}
	// One seed, one consumer, so exactly one file was added and the coverage line should say so.
	if narrowed.DependentCount != 1 {
		t.Errorf("the dependent count should be 1, got %d", narrowed.DependentCount)
	}
	if narrowed.Everything {
		t.Error("a scope well inside the limit reported itself as the whole tree")
	}
}

// TestNarrowToClosureFallsBackRatherThanTruncating holds the failure direction.
//
// Above the limit the honest answer is the whole tree. A truncated closure would be fast, silent,
// and wrong in the direction that hides findings, which is the one failure a fast path must not
// introduce, and it would look identical to a working scope from the coverage line.
func TestNarrowToClosureFallsBackRatherThanTruncating(t *testing.T) {
	t.Parallel()
	files := map[string]string{"Producer.ts": "export function value(): number {\n  return 1;\n}\n"}
	for index := 0; index <= program.DependentClosureLimit; index++ {
		name := filepath.Join("consumers", "Consumer"+itoaForTest(index)+".ts")
		files[name] = "import { value } from '../Producer';\nexport const used" +
			itoaForTest(index) + " = value();\n"
	}

	graph, directory := buildNarrowFixtureGraph(t, files)

	producer := filepath.Join(directory, "Producer.ts")
	scope := formatScope{
		FileNames:          []string{producer},
		index:              map[string]struct{}{producer: {}},
		RequestDescription: "Producer.ts",
	}

	narrowed, narrowedFiles := narrowToClosure(graph, scope, graph.ProjectFiles())

	if !narrowed.Everything {
		t.Fatalf("a closure past the limit did not fall back to the whole tree: %d files", len(narrowedFiles))
	}
	if len(narrowedFiles) != len(graph.ProjectFiles()) {
		t.Errorf("the fallback returned %d files rather than the whole program's %d",
			len(narrowedFiles), len(graph.ProjectFiles()))
	}

	// The control: a leaf in the same program narrows normally, so the fallback is a response to the
	// closure's size rather than a function that always gives up.
	leaf := filepath.Join(directory, "consumers", "Consumer0.ts")
	leafScope := formatScope{
		FileNames:          []string{leaf},
		index:              map[string]struct{}{leaf: {}},
		RequestDescription: "Consumer0.ts",
	}
	if narrowedLeaf, _ := narrowToClosure(graph, leafScope, graph.ProjectFiles()); narrowedLeaf.Everything {
		t.Error("a leaf file's closure fell back to the whole tree")
	}
}

// TestNarrowToClosureLeavesAWholeTreeScopeAlone is the case a bare run takes, and it must cost
// nothing: no closure, no filtering, the same slice back.
func TestNarrowToClosureLeavesAWholeTreeScopeAlone(t *testing.T) {
	t.Parallel()
	graph, _ := buildNarrowFixtureGraph(t, map[string]string{
		"Producer.ts": "export function value(): number {\n  return 1;\n}\n",
	})

	projectFiles := graph.ProjectFiles()
	narrowed, files := narrowToClosure(graph, formatScope{Everything: true}, projectFiles)

	if !narrowed.Everything || len(files) != len(projectFiles) {
		t.Errorf("a whole-tree scope was narrowed: everything=%v, %d files of %d",
			narrowed.Everything, len(files), len(projectFiles))
	}
	if narrowed.DependentCount != 0 {
		t.Errorf("a whole-tree scope reported %d dependents", narrowed.DependentCount)
	}
}

func itoaForTest(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
