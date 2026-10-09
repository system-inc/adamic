package program

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// graduationProject builds a project of the named files, every one exporting something, and returns the graph and
// its project source files in the program's order.
func graduationProject(t *testing.T, names ...string) (*Graph, []*ast.SourceFile) {
	t.Helper()
	root := t.TempDir()
	config := `{"compilerOptions": {"target": "ES2022", "module": "esnext", "moduleResolution": "bundler", "strict": true, "noEmit": true}}`
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		text := "export function " + name + "(value: number): number {\n  return value + 1;\n}\n"
		if err := os.WriteFile(filepath.Join(root, name+".ts"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := Build(Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatal(err)
	}
	var sources []*ast.SourceFile
	for _, sourceFile := range graph.ProjectFiles() {
		if !sourceFile.IsDeclarationFile && !ast.IsJsonSourceFile(sourceFile) {
			sources = append(sources, sourceFile)
		}
	}
	if len(sources) != len(names) {
		t.Fatalf("the program holds %d source files, want %d", len(sources), len(names))
	}
	return graph, sources
}

// A first run with no build info keys every shape on content (#5txm9gg). Graduating gives each the signature the
// up-front emit computes for the same bytes, so a graduated shape is the same shape a recorded one would be
// (#9knyr86).
func TestGraduationGivesEachContentKeyedShapeTheSignatureAnEmitComputes(t *testing.T) {
	t.Parallel()
	graph, sources := graduationProject(t, "alpha", "beta")
	ctx := context.Background()
	shapes, _ := graph.Signatures(ctx, nil)
	for _, sourceFile := range sources {
		if entry := shapes[sourceFile.FileName().AsString()]; entry.Signature != entry.Version {
			t.Fatalf("%s already has a real signature before graduating, so this proves nothing", sourceFile.FileName())
		}
	}

	graduated, abandoned := graph.GraduateSignatures(shapes, time.Now().Add(time.Minute), 0)
	if graduated != len(sources) || abandoned != "" {
		t.Fatalf("graduated %d of %d, abandoned %q; want every one and none abandoned", graduated, len(sources), abandoned)
	}
	emitted, _ := graph.Signatures(ctx, RecordedRealSignatures(graph))
	for _, sourceFile := range sources {
		entry := shapes[sourceFile.FileName().AsString()]
		if entry.Signature == entry.Version || entry.Abandoned {
			t.Errorf("%s did not graduate: %+v", sourceFile.FileName(), entry)
		}
		if want := emitted[sourceFile.FileName().AsString()]; entry != want {
			t.Errorf("%s graduated to %+v, and the up-front emit computes %+v", sourceFile.FileName(), entry, want)
		}
	}
}

// The deadline holds against an emit that never ends (#5txm9gg: one ESLint config ran 9 minutes). The file in
// progress at the deadline is abandoned, the files before it graduated, the files after it are left for the next
// run untouched, and an abandoned file is never emitted again while its bytes are the same.
//
// Not parallel: it sets beforeSignatureEmit, the package's emit hook.
func TestGraduationStopsAtItsDeadlineAndNeverRetriesTheFileItAbandoned(t *testing.T) {
	graph, sources := graduationProject(t, "alpha", "beta", "gamma")
	before, planted, after := sources[0].FileName(), sources[1].FileName(), sources[2].FileName()
	shapes, _ := graph.Signatures(context.Background(), nil)

	release := make(chan struct{})
	t.Cleanup(func() {
		close(release)
		beforeSignatureEmit = nil
	})
	beforeSignatureEmit = func(fileName string) {
		if fileName == planted.AsString() {
			<-release
		}
	}
	started := time.Now()
	graduated, abandoned := graph.GraduateSignatures(shapes, started.Add(200*time.Millisecond), 0)
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("graduating returned after %s, so the deadline did not stop it", elapsed)
	}
	if abandoned != planted.AsString() || graduated != 1 {
		t.Fatalf("graduated %d and abandoned %q; want 1 and the planted %s", graduated, abandoned, planted)
	}
	if entry := shapes[before.AsString()]; entry.Signature == entry.Version {
		t.Errorf("the file before the planted one did not graduate: %+v", entry)
	}
	if entry := shapes[planted.AsString()]; !entry.Abandoned || entry.Signature != entry.Version {
		t.Errorf("the planted file is not abandoned and content-keyed: %+v", entry)
	}
	if entry := shapes[after.AsString()]; entry.Abandoned || entry.Signature != entry.Version {
		t.Errorf("the file after the planted one was touched: %+v", entry)
	}

	// The next run tries what was left, and never the abandoned file.
	beforeSignatureEmit = func(fileName string) {
		if fileName == planted.AsString() {
			t.Errorf("the abandoned %s was emitted again with its bytes unchanged", planted)
		}
	}
	if graduated, abandoned := graph.GraduateSignatures(shapes, time.Now().Add(time.Minute), 0); graduated != 1 || abandoned != "" {
		t.Errorf("the next run graduated %d and abandoned %q; want the one file left and none", graduated, abandoned)
	}

	// Carried while its bytes are the same, and dropped once they change.
	carried, _ := graph.Signatures(context.Background(), shapes)
	if !carried[planted.AsString()].Abandoned {
		t.Errorf("the abandoned mark was dropped with the bytes unchanged: %+v", carried[planted.AsString()])
	}
	edited := map[string]SignatureEntry{}
	for name, entry := range shapes {
		edited[name] = entry
	}
	edited[planted.AsString()] = SignatureEntry{Version: "other bytes", Signature: "other bytes", Abandoned: true}
	fresh, _ := graph.Signatures(context.Background(), edited)
	if fresh[planted.AsString()].Abandoned {
		t.Errorf("the abandoned mark survived an edit to the file's bytes: %+v", fresh[planted.AsString()])
	}
}

// A file the deadline catches before its emit has run for abandonAfter is innocent: it is left keyed on content
// and unmarked, for the next run to try first, never abandoned for the bad luck of starting last.
//
// Not parallel: it sets beforeSignatureEmit, the package's emit hook.
func TestGraduationLeavesAFileTheDeadlineCaughtEarlyForTheNextRun(t *testing.T) {
	graph, sources := graduationProject(t, "alpha", "beta")
	caught := sources[1].FileName()
	shapes, _ := graph.Signatures(context.Background(), nil)

	release := make(chan struct{})
	t.Cleanup(func() {
		close(release)
		beforeSignatureEmit = nil
	})
	beforeSignatureEmit = func(fileName string) {
		if fileName == caught.AsString() {
			<-release
		}
	}
	graduated, abandoned := graph.GraduateSignatures(shapes, time.Now().Add(200*time.Millisecond), time.Hour)
	if graduated != 1 || abandoned != "" {
		t.Fatalf("graduated %d and abandoned %q; want 1 and none, since the caught emit ran for less than abandonAfter", graduated, abandoned)
	}
	if entry := shapes[caught.AsString()]; entry.Abandoned || entry.Signature != entry.Version {
		t.Errorf("the file caught early was not left as it was for the next run: %+v", entry)
	}
}
