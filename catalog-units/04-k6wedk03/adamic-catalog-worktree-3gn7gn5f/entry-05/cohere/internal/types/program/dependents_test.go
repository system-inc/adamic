package program

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// buildFixtureGraph writes a small program to disk and builds a real type graph over it.
//
// A real graph rather than a hand-built map, because the thing under test is whether the reversal
// agrees with how the compiler actually keys its resolutions. A fixture that constructs the map
// itself would have passed while the real reversal produced almost no edges, which is the defect
// this file exists to have caught.
func buildFixtureGraph(t *testing.T, files map[string]string) *Graph {
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

	graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building the fixture graph: %v", err)
	}
	return graph
}

func fileNamed(t *testing.T, graph *Graph, suffix string) *ast.SourceFile {
	t.Helper()
	for _, sourceFile := range graph.ProjectFiles() {
		if strings.HasSuffix(sourceFile.FileName().AsString(), suffix) {
			return sourceFile
		}
	}
	t.Fatalf("no project file ends in %s", suffix)
	return nil
}

func closureNames(closure []*ast.SourceFile) []string {
	names := make([]string, 0, len(closure))
	for _, sourceFile := range closure {
		names = append(names, filepath.Base(sourceFile.FileName().AsString()))
	}
	return names
}

// TestDependentClosureReachesConsumersThroughImports is the case that motivated the whole scope.
//
// A change to Producer breaks Consumer, which the edit never touched, so a run scoped to Producer
// alone reports nothing about it. Measured on the real tree before this existed: one edited file
// produced seven new findings, four of them in a consumer two directories away.
func TestDependentClosureReachesConsumersThroughImports(t *testing.T) {
	t.Parallel()
	graph := buildFixtureGraph(t, map[string]string{
		"Producer.ts":  "export function value(): number {\n  return 1;\n}\n",
		"Consumer.ts":  "import { value } from './Producer';\nexport const doubled = value() * 2;\n",
		"Indirect.ts":  "import { doubled } from './Consumer';\nexport const tripled = doubled * 3;\n",
		"Unrelated.ts": "export const alone = 1;\n",
	})

	closure, within := graph.DependentClosure([]*ast.SourceFile{fileNamed(t, graph, "Producer.ts")})
	if !within {
		t.Fatal("a four-file program exceeded the closure limit")
	}

	names := strings.Join(closureNames(closure), " ")
	for _, want := range []string{"Producer.ts", "Consumer.ts", "Indirect.ts"} {
		if !strings.Contains(names, want) {
			t.Errorf("%s should be in the closure, got %s", want, names)
		}
	}

	// The control. Without it a closure that returned every project file would satisfy every
	// assertion above, which is the shape the defect actually took: a broken reversal that silently
	// returned too little read the same as one returning too much.
	if strings.Contains(names, "Unrelated.ts") {
		t.Errorf("a file that imports nothing in the chain was in the closure: %s", names)
	}
}

// TestDependentClosureDoesNotFollowImportsForward holds the direction.
//
// Consumer imports Producer, so editing Consumer cannot break Producer. A closure that walked the
// resolution map without reversing it would return the wrong set and still look plausible: same
// size, same shape, findings in files that were genuinely related to the edit.
func TestDependentClosureDoesNotFollowImportsForward(t *testing.T) {
	t.Parallel()
	graph := buildFixtureGraph(t, map[string]string{
		"Producer.ts": "export function value(): number {\n  return 1;\n}\n",
		"Consumer.ts": "import { value } from './Producer';\nexport const doubled = value() * 2;\n",
	})

	closure, within := graph.DependentClosure([]*ast.SourceFile{fileNamed(t, graph, "Consumer.ts")})
	if !within {
		t.Fatal("a two-file program exceeded the closure limit")
	}

	names := strings.Join(closureNames(closure), " ")
	if strings.Contains(names, "Producer.ts") {
		t.Errorf("the closure followed an import forwards: %s", names)
	}
	if !strings.Contains(names, "Consumer.ts") {
		t.Errorf("the seed itself was not in its own closure: %s", names)
	}
}

// TestDependentClosureRefusesRatherThanTruncating holds the failure direction.
//
// Above the limit the answer is `check everything`, never a partial set. A truncated closure would
// be fast, silent, and wrong in the direction that hides findings, which is the one failure mode
// scoping must not introduce.
func TestDependentClosureRefusesRatherThanTruncating(t *testing.T) {
	t.Parallel()
	files := map[string]string{"Producer.ts": "export function value(): number {\n  return 1;\n}\n"}
	for index := 0; index <= DependentClosureLimit; index++ {
		files[filepath.Join("consumers", "Consumer"+itoa(index)+".ts")] =
			"import { value } from '../Producer';\nexport const used" + itoa(index) + " = value();\n"
	}

	graph := buildFixtureGraph(t, files)

	if _, within := graph.DependentClosure([]*ast.SourceFile{fileNamed(t, graph, "Producer.ts")}); within {
		t.Fatal("a closure past the limit reported itself as usable")
	}

	// The control: the same graph, seeded at a file nothing imports, is comfortably inside the limit.
	// Without it a function that always refused would pass the assertion above.
	if _, within := graph.DependentClosure([]*ast.SourceFile{fileNamed(t, graph, "Consumer0.ts")}); !within {
		t.Error("a leaf file's closure was reported as too large")
	}
}

func itoa(value int) string {
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
