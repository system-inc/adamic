package program

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// signatureTree is a fixture built twice in one directory, before and after an edit, so its paths, which
// every fingerprint hashes, are the same in both graphs.
type signatureTree struct {
	t         *testing.T
	directory string
}

func newSignatureTree(t *testing.T, files map[string]string) *signatureTree {
	t.Helper()
	tree := &signatureTree{t: t, directory: t.TempDir()}
	tree.write(files)
	tree.write(map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler"},"include":["**/*.ts"]}`})
	return tree
}

func (tree *signatureTree) write(files map[string]string) {
	tree.t.Helper()
	for name, contents := range files {
		path := filepath.Join(tree.directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tree.t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			tree.t.Fatal(err)
		}
	}
}

func (tree *signatureTree) graph() *Graph {
	tree.t.Helper()
	graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: tree.directory})
	if err != nil {
		tree.t.Fatal(err)
	}
	return graph
}

// state is one build's signatures and both fingerprints, keyed by base name.
type signatureState struct {
	signatures map[string]SignatureEntry
	computed   int
	emits      int
	bySignature,
	byContent map[string]string
}

func (tree *signatureTree) state(previous map[string]SignatureEntry) signatureState {
	tree.t.Helper()
	graph := tree.graph()
	// A first state has its signatures computed, as a project with build info does: these tests are about
	// what a computed signature holds. A first run with nothing recorded is its own test below.
	if previous == nil {
		previous = RecordedRealSignatures(graph)
	}
	signatures, computed := graph.Signatures(context.Background(), previous)
	state := signatureState{signatures: signatures, computed: computed, emits: graph.SignatureEmits(), bySignature: map[string]string{}, byContent: map[string]string{}}
	signatureFingerprints := graph.SignatureFingerprints(signatures)
	contentFingerprints := graph.TypeFingerprints()
	for _, sourceFile := range graph.ProjectFiles() {
		name := filepath.Base(sourceFile.FileName().AsString())
		signatureSum := signatureFingerprints[sourceFile.PathKey()]
		contentSum := contentFingerprints[sourceFile.PathKey()]
		state.bySignature[name] = string(signatureSum[:])
		state.byContent[name] = string(contentSum[:])
	}
	return state
}

func (state signatureState) signature(t *testing.T, name string) string {
	t.Helper()
	for fileName, entry := range state.signatures {
		if filepath.Base(fileName) == name {
			return entry.Signature
		}
	}
	t.Fatalf("no signature for %s", name)
	return ""
}

const signatureFixtureC = "export type Shape = { width: number };\n"
const signatureFixtureB = "import type { Shape } from \"./c\";\n" +
	"export function measure(shape: Shape): number {\n  return shape.width;\n}\n" +
	"export async function load(): Promise<number> {\n  return 1;\n}\n"
const signatureFixtureA = "import { measure } from \"./b\";\nexport const width = measure({ width: 2 });\n"

func signatureFixture(t *testing.T) *signatureTree {
	return newSignatureTree(t, map[string]string{"a.ts": signatureFixtureA, "b.ts": signatureFixtureB, "c.ts": signatureFixtureC})
}

// A body edit is the case the whole table exists for: the edited file's own fingerprint moves, and its
// importer's does not, where the content fingerprint moves both.
func TestABodyEditLeavesItsImportersSignatureFingerprintAlone(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "return shape.width;", "return shape.width * 1;", 1)})
	after := tree.state(before.signatures)

	if after.signature(t, "b.ts") != before.signature(t, "b.ts") {
		t.Fatal("a body edit changed the file's signature, so the fixture is not testing a body edit")
	}
	if after.computed != 1 {
		t.Errorf("computed %d signatures, want exactly the edited file: the rest must carry forward", after.computed)
	}
	if after.bySignature["a.ts"] != before.bySignature["a.ts"] {
		t.Error("the importer's signature fingerprint moved on a body edit, so nothing was gained")
	}
	if after.bySignature["b.ts"] == before.bySignature["b.ts"] {
		t.Error("the edited file's own fingerprint did not move, so its own findings would replay stale")
	}
	if after.byContent["a.ts"] == before.byContent["a.ts"] {
		t.Error("the content fingerprint did not move either, so this fixture shows no difference between the two")
	}
}

// An edit to what a file exports must reach every importer.
func TestAnExportEditMovesItsImportersSignatureFingerprint(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "export function measure(shape: Shape): number {", "export function measure(shape: Shape): number | string {", 1)})
	after := tree.state(before.signatures)

	if after.signature(t, "b.ts") == before.signature(t, "b.ts") {
		t.Fatal("an exported return type changed and the signature did not")
	}
	if after.bySignature["a.ts"] == before.bySignature["a.ts"] {
		t.Error("an importer's fingerprint survived a change to what it imports")
	}
}

// Two levels down: c's exported type changes and b's declaration output, which only names `Shape`, does not.
// a reaches c through b, so a's fingerprint must still move.
func TestAChangeTwoLevelsDownReachesTheImporter(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"c.ts": "export type Shape = { width: string };\n"})
	after := tree.state(before.signatures)

	if after.signature(t, "b.ts") != before.signature(t, "b.ts") {
		t.Fatal("b's signature moved, so this fixture does not isolate the transitive case")
	}
	if after.bySignature["a.ts"] == before.bySignature["a.ts"] {
		t.Error("a change two imports away did not reach a's fingerprint")
	}
}

// JSDoc on an export is part of its declaration output, so a rule reading a deprecation through types sees
// the change. Pinned because no-deprecated depends on it.
func TestAJSDocEditOnAnExportChangesTheSignature(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "export function measure(", "/** @deprecated use something else */\nexport function measure(", 1)})
	after := tree.state(before.signatures)

	if after.signature(t, "b.ts") == before.signature(t, "b.ts") {
		t.Error("a @deprecated tag on an export left the signature unchanged, so a deprecation would replay stale")
	}
}

// `async` changes a declaration's syntax and not its output, which is the gap the syntax half of the shape
// closes: the signature stays, the shape moves, and the importer's fingerprint moves with it.
func TestAnAsyncOnlyEditReachesImportersThroughTheSyntaxHalf(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "export async function load(): Promise<number> {\n  return 1;", "export function load(): Promise<number> {\n  return Promise.resolve(1);", 1)})
	after := tree.state(before.signatures)

	if after.signature(t, "b.ts") != before.signature(t, "b.ts") {
		t.Fatal("the declaration output saw an async-only edit, so this fixture no longer isolates the syntax half")
	}
	if after.bySignature["a.ts"] == before.bySignature["a.ts"] {
		t.Error("an async-only edit did not reach the importer, so a rule reading `async` off an imported declaration would replay stale")
	}
}

// What stays invisible is a body's own text: an edit inside one moves neither half.
func TestAnEditInsideABodyMovesNeitherHalf(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "  return 1;", "  return 1 + 0;", 1)})
	after := tree.state(before.signatures)

	for name, entry := range after.signatures {
		if filepath.Base(name) == "b.ts" && entry.Syntax != before.signatures[name].Syntax {
			t.Error("an edit inside a body changed the syntax half")
		}
	}
	if after.bySignature["a.ts"] != before.bySignature["a.ts"] {
		t.Error("an edit inside a body reached the importer")
	}
}

// A file that declares something global changes types everywhere without being imported.
func TestAGlobalDeclarationMovesEverySignatureFingerprint(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	tree.write(map[string]string{"globals.d.ts": "declare global { interface Window { first: string } }\nexport {};\n"})
	before := tree.state(nil)
	tree.write(map[string]string{"globals.d.ts": "declare global { interface Window { second: string } }\nexport {};\n"})
	after := tree.state(before.signatures)

	for _, name := range []string{"a.ts", "b.ts", "c.ts"} {
		if after.bySignature[name] == before.bySignature[name] {
			t.Errorf("%s's fingerprint survived a change to a global declaration", name)
		}
	}
}

// Nothing changed means nothing computed and nothing moved.
func TestAnUnchangedTreeComputesNoSignatures(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	before := tree.state(nil)
	if before.computed != 3 {
		t.Errorf("a first run computed %d signatures, want all 3", before.computed)
	}
	after := tree.state(before.signatures)
	if after.computed != 0 {
		t.Errorf("an unchanged tree computed %d signatures", after.computed)
	}
	for name, sum := range before.bySignature {
		if after.bySignature[name] != sum {
			t.Errorf("%s's fingerprint moved on an unchanged tree", name)
		}
	}
}

// Seeding is only worth anything if the compiler's recorded signatures are the ones computed here, so a
// seeded entry and a computed one must be equal byte for byte. Measured on ahra as 3,741 of 3,741; pinned
// here so a compiler bump that changes the definition fails a test instead of quietly over-invalidating.
//
// The compiler records a shape only for a file it saw change, so a first build's buildinfo has none. The
// fixture builds, edits b, and builds again, which is how a real tree's buildinfo fills up.
func TestSeededSignaturesEqualComputedOnes(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	tree.write(map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler","incremental":true,"noEmit":true,"tsBuildInfoFile":".cache/tsconfig.tsbuildinfo"},"include":["**/*.ts"]}`})
	tree.graph().IncrementalDiagnostics(context.Background())
	tree.write(map[string]string{"b.ts": strings.Replace(signatureFixtureB, "return shape.width;", "return shape.width * 1;", 1)})
	tree.graph().IncrementalDiagnostics(context.Background())

	graph := tree.graph()
	seeded := graph.SeedSignatures(nil)
	computed, _ := graph.Signatures(context.Background(), RecordedRealSignatures(graph))
	if len(seeded) == 0 {
		t.Fatal("nothing was seeded from the buildinfo, so this proves nothing")
	}
	for name, entry := range seeded {
		if computed[name].Version != entry.Version || computed[name].Signature != entry.Signature {
			t.Errorf("%s: seeded %+v, computed %+v", filepath.Base(name), entry, computed[name])
		}
	}
	if _, recomputed := graph.Signatures(context.Background(), seeded); recomputed != 3-len(seeded) {
		t.Errorf("after seeding %d entries, %d were computed again", len(seeded), recomputed)
	}
}

// A first run with nothing recorded emits no declarations: every source file's version stands in for its
// signature, and the graph says how many. Declaration emit has no ceiling (an ESLint config whose export type
// is inferred from typescript-eslint ran 9 minutes and 17 GB on www-ahra-ai, #5txm9gg), so up front it runs
// only for a file whose signature is known. Counted by the emits themselves, not timed.
func TestAFirstRunEmitsNoDeclarations(t *testing.T) {
	t.Parallel()
	tree := newSignatureTree(t, map[string]string{
		"a.ts":      signatureFixtureA,
		"b.ts":      signatureFixtureB,
		"c.ts":      signatureFixtureC,
		"config.ts": "import { measure } from \"./b\";\nexport default [{ rules: { width: measure }, files: [\"**/*.ts\"] }];\n",
	})
	graph := tree.graph()
	signatures, _ := graph.Signatures(context.Background(), nil)
	if graph.SignatureEmits() != 0 {
		t.Errorf("a first run emitted declarations for %d files, want none", graph.SignatureEmits())
	}
	for fileName, entry := range signatures {
		if entry.Signature != entry.Version || entry.Syntax != entry.Version {
			t.Errorf("%s: a first run's shape is not its content (%+v)", filepath.Base(fileName), entry)
		}
	}
	if graph.ContentKeyedShapes != 4 {
		t.Errorf("the graph counts %d shapes keyed on content, want 4", graph.ContentKeyedShapes)
	}
}

// An edited file is emitted only when its recorded signature is real. One keyed on content stays keyed on
// content, and its edit still reaches its importer, since its whole text is its shape.
func TestOnlyAnEditedFileWithARealSignatureIsEmitted(t *testing.T) {
	t.Parallel()
	tree := signatureFixture(t)
	known := tree.state(nil)
	if known.emits != 3 {
		t.Fatalf("the fixture's first state emitted %d files, so its signatures are not real ones", known.emits)
	}
	bodyEdit := strings.Replace(signatureFixtureB, "return shape.width;", "return shape.width * 1;", 1)
	tree.write(map[string]string{"b.ts": bodyEdit})
	edited := tree.state(known.signatures)
	if edited.emits != 1 {
		t.Errorf("an edit to a file with a real signature emitted %d files, want the one edited", edited.emits)
	}

	tree.write(map[string]string{"b.ts": signatureFixtureB})
	graph := tree.graph()
	contentKeyed, _ := graph.Signatures(context.Background(), nil)
	before := graph.SignatureFingerprints(contentKeyed)
	tree.write(map[string]string{"b.ts": bodyEdit})
	graph = tree.graph()
	after, _ := graph.Signatures(context.Background(), contentKeyed)
	if graph.SignatureEmits() != 0 {
		t.Errorf("an edit to a file keyed on content emitted %d files, want none", graph.SignatureEmits())
	}
	moved := graph.SignatureFingerprints(after)
	for _, sourceFile := range graph.ProjectFiles() {
		if filepath.Base(sourceFile.FileName().AsString()) == "a.ts" && moved[sourceFile.PathKey()] == before[sourceFile.PathKey()] {
			t.Error("a body edit to a file keyed on content did not reach its importer, so the importer would replay stale")
		}
	}
}
