package program_test

import (
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// keysOf builds the project and returns one file's type fingerprint and shape fingerprint.
func keysOf(t *testing.T, root string, name string) ([sha256.Size]byte, [sha256.Size]byte) {
	t.Helper()
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(root, "tsconfig.json")})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	shapes, _ := graph.Signatures(context.Background(), program.RecordedRealSignatures(graph))
	for _, sourceFile := range graph.ProjectFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == name {
			path := sourceFile.PathKey()
			return graph.TypeFingerprints()[path], graph.SignatureFingerprints(shapes)[path]
		}
	}
	t.Fatalf("%s is not a project file", name)
	return [sha256.Size]byte{}, [sha256.Size]byte{}
}

// a.ts exports nothing derived from v, so its own shape cannot move with v's type and only its view through
// b.ts can move its keys.
//
// When a re-export starts resolving somewhere else, the keys of the files that import the re-exporter move,
// though no byte of either file changed: what a.ts's types see through b.ts's `export *` is whatever it
// resolves to now (#9bjjk4a, a question from @system_cohere_lint before the React compiler rules key on
// shapes).
//
// Two ways a resolution moves. Into another project file, where the import edge itself moves. And between
// two installed declaration files that both stay in the program, where no project edge exists and no file
// enters or leaves the global component: a package.json's "types" changing is enough.
func TestAReExportResolvingElsewhereMovesItsImportersKeys(t *testing.T) {
	t.Parallel()
	t.Run("into another project file", func(t *testing.T) {
		t.Parallel()
		root := writeProject(t, map[string]string{
			"tsconfig.json": minimalConfig,
			"a.ts":          "import { v } from \"./b\";\nconst negated = -v;\nexport const used = typeof negated === \"number\";\n",
			"b.ts":          "export * from \"./c\";\n",
			"c.ts":          "export const v = 1;\n",
		})
		types, shapes := keysOf(t, root, "a.ts")
		// The same text at a new place: ./c now resolves to c/index.ts, and only where it lives changed.
		if err := os.Remove(filepath.Join(root, "c.ts")); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "c"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "c", "index.ts"), []byte("export const v = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		movedTypes, movedShapes := keysOf(t, root, "a.ts")
		if movedTypes == types || movedShapes == shapes {
			t.Errorf("a.ts's keys did not move when b.ts's re-export resolved to another file (type moved %v, shape moved %v)",
				movedTypes != types, movedShapes != shapes)
		}
	})

	t.Run("between two installed files", func(t *testing.T) {
		t.Parallel()
		root := writeProject(t, map[string]string{
			"tsconfig.json": minimalConfig,
			"a.ts":          "import { v } from \"./b\";\nconst negated = -v;\nexport const used = typeof negated === \"number\";\n",
			"b.ts":          "export * from \"pkg\";\n",
			// Both declarations stay in the program whichever one "types" names, so neither entering nor
			// leaving it can be what moves the key.
			"keep.ts":                       "import \"pkg/one\";\nimport \"pkg/two\";\nexport {};\n",
			"node_modules/pkg/package.json": "{\"name\": \"pkg\", \"types\": \"one.d.ts\"}\n",
			"node_modules/pkg/one.d.ts":     "export declare const v: number;\n",
			"node_modules/pkg/two.d.ts":     "export declare const v: string;\n",
		})
		types, shapes := keysOf(t, root, "a.ts")
		if err := os.WriteFile(filepath.Join(root, "node_modules/pkg/package.json"),
			[]byte("{\"name\": \"pkg\", \"types\": \"two.d.ts\"}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		movedTypes, movedShapes := keysOf(t, root, "a.ts")
		if movedTypes == types || movedShapes == shapes {
			t.Errorf("a.ts's keys did not move when b.ts's re-export resolved to another installed file, though v changed "+
				"from number to string (type moved %v, shape moved %v)", movedTypes != types, movedShapes != shapes)
		}
	})
}
