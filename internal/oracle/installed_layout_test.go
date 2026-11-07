package oracle

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// installed-layout's answer is measured, not remembered: tsc 6.0.3's require("source-map-support"),
// from lib/_tsc.js in a typescript package with nothing installed beside it, run on the Node the
// oracle runs. The error's name, code and first line are the constants native's rows are pinned from
// (backend_divergences_test.go), so this fails if Node's answer ever changes. The rest is the
// declared difference: a Require stack tail and a requireStack naming the requiring file.
func TestInstalledLayoutIsNodes(t *testing.T) {
	t.Parallel()
	library := filepath.Join(t.TempDir(), "typescript", "lib")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}
	// The package's own package.json, without a type, as npm installs it: _tsc.js is CommonJS whatever
	// a directory above the temporary one says.
	if err := os.WriteFile(filepath.Join(library, "..", "package.json"), []byte(`{"name": "typescript", "version": "6.0.3"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tsc := filepath.Join(library, "_tsc.js")
	script := `try { require("source-map-support"); console.log("installed"); } catch (e) { const lines = e.message.split("\n"); console.log([e.name, e.code, lines[0], lines[1], Object.keys(e).join(",")].join("\n")); }` + "\n"
	if err := os.WriteFile(tsc, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	// NODE_PATH would add a directory to where Node looks, which the installed layout doesn't have.
	measured := executeWith(t, []string{"NODE_PATH="}, "node", tsc)
	if measured.exitCode != 0 || len(measured.stderr) != 0 {
		t.Fatalf("node: exit %d, stdout %q, stderr %q", measured.exitCode, measured.stdout, measured.stderr)
	}
	lines := strings.Split(strings.TrimSuffix(string(measured.stdout), "\n"), "\n")
	want := []string{installedLayoutName, installedLayoutCode, installedLayoutMessage, "Require stack:", "code,requireStack"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Node's answer for an installed tsc: got %q, want %q", lines, want)
	}
}

// installedLayoutReadsFixture prints what tsc derives from __filename and __dirname, a line per read, and neither
// raw (Adamic and Platforms on #0ta3dk9).
const installedLayoutReadsFixture = "testdata/installed_layout/tsc_reads.ts"

// installedLayoutFiles are the files tsc_reads.ts reads beside its entry, by their path under lib, with what each
// holds: the default lib and the es2020 lib it steps to, and the messages --locale ja loads.
var installedLayoutFiles = map[string]string{
	"lib.d.ts":                             "/// <reference no-default-lib=\"true\"/>\n",
	"lib.es2020.d.ts":                      "/// <reference no-default-lib=\"true\"/>\n",
	"ja/diagnosticMessages.generated.json": `{"Version_0": "バージョン {0}"}`,
}

// installedLayoutReads are the lines Node prints for tsc_reads.ts in the layout, the case probe's aside: it answers
// the file system's case sensitivity, which differs between the gate's Linux and a Mac, and both legs share it.
var installedLayoutReads = map[int]string{
	0: "directory lib",
	1: "lib.d.ts true",
	3: "locale " + installedLayoutFiles["ja/diagnosticMessages.generated.json"],
	4: "lib.es2020.d.ts true",
}

// layInstalledTypeScript lays out typescript/lib under root as npm installs it, with installedLayoutFiles in lib and
// the package's own package.json, without a type, so a .js entry there is CommonJS whatever a directory above says.
// It returns lib, where the caller puts the entry, by its real path: every leg's answer for its own path is then lib's
// own spelling, the one installedLayoutShape takes out (a Mac's temporary directory is under a link, /var).
func layInstalledTypeScript(t *testing.T, root string) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	library := filepath.Join(root, "typescript", "lib")
	if err := os.MkdirAll(filepath.Join(library, "ja"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(library, "..", "package.json"), []byte(`{"name": "typescript", "version": "6.0.3"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, contents := range installedLayoutFiles {
		if err := os.WriteFile(filepath.Join(library, filepath.FromSlash(name)), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return library
}

// installedLayoutMissing is each file of installedLayoutFiles that isn't in library, sorted.
func installedLayoutMissing(library string) []string {
	var missing []string
	for name := range installedLayoutFiles {
		if information, err := os.Stat(filepath.Join(library, filepath.FromSlash(name))); err != nil || !information.Mode().IsRegular() {
			missing = append(missing, name)
		}
	}
	slices.Sort(missing)
	return missing
}

// runInstalledLayout checks the layout around entry, then runs name with arguments, so a file missing from the
// layout fails as the layout, never as a read that comes out wrong. It runs from this package's directory, outside
// lib, as a user runs tsc from a project.
func runInstalledLayout(t *testing.T, entry string, environment []string, name string, arguments ...string) run {
	t.Helper()
	if missing := installedLayoutMissing(filepath.Dir(entry)); len(missing) != 0 {
		t.Fatalf("the installed layout beside %s lacks %v", entry, missing)
	}
	return executeWith(t, environment, name, arguments...)
}

// installedLayoutNode runs fixture, a path from this package, on Node as tsc ships: transpiled to CommonJS by
// TypeScript 6.0.3's own transpileModule (programNodeAPI), as lib/_tsc.js. It returns the run and lib.
func installedLayoutNode(t *testing.T, fixture string) (run, string) {
	t.Helper()
	api := filepath.Join(programNodeAPI(t), "lib/typescript.js")
	transpiled := filepath.Join(t.TempDir(), "_tsc.js")
	script := `const ts = require(process.argv[1]); const fs = require("node:fs"); ` +
		`const output = ts.transpileModule(fs.readFileSync(process.argv[2], "utf8"), {fileName: "_tsc.ts", compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022}}); ` +
		`fs.writeFileSync(process.argv[3], output.outputText);`
	if result := execute(t, "node", "-e", script, api, fixture, transpiled); result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("transpiling %s to CommonJS: exit %d, stderr %q", fixture, result.exitCode, result.stderr)
	}
	program, err := os.ReadFile(transpiled)
	if err != nil {
		t.Fatal(err)
	}
	library := layInstalledTypeScript(t, t.TempDir())
	tsc := filepath.Join(library, "_tsc.js")
	if err := os.WriteFile(tsc, program, 0o644); err != nil {
		t.Fatal(err)
	}
	// NODE_PATH would add a directory to where Node looks, which the installed layout doesn't have.
	return runInstalledLayout(t, tsc, []string{"NODE_PATH="}, "node", tsc), library
}

// installedLayoutLowered is fixture, a path from this package, lowered as a build that declares CommonJS lowers it
// (lower.ModuleFormatCommonJS).
func installedLayoutLowered(t *testing.T, fixture string) *ir.Program {
	t.Helper()
	path, err := filepath.Abs(fixture)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := lower.LowerWithOptions(context.Background(), loaded, lower.Options{ModuleFormat: lower.ModuleFormatCommonJS})
	if err != nil {
		t.Fatal(err)
	}
	return program
}

// installedLayoutNative runs fixture built as a user's build is, declared CommonJS, as lib/_tsc. mutate rewrites the
// program's C before it's built, or is nil. It returns the run and lib.
func installedLayoutNative(t *testing.T, fixture string, mutate func(t *testing.T, code string) string) (run, string) {
	t.Helper()
	code := native.C(installedLayoutLowered(t, fixture))
	if mutate != nil {
		code = mutate(t, code)
	}
	library := layInstalledTypeScript(t, t.TempDir())
	tsc := filepath.Join(library, "_tsc")
	if err := native.Build(code, tsc, native.Options{}); err != nil {
		t.Fatal(err)
	}
	return runInstalledLayout(t, tsc, nil, tsc), library
}

// installedLayoutJavaScript runs fixture through the JavaScript backend, declared CommonJS, as lib/_tsc.mjs: the
// emitted program is an ES module, and the package's own package.json has no type, so only .mjs runs it as one. It
// runs under oracle/node.mjs, as onJavaScriptBackend does, so its import of 'adamic' finds the runtime. It returns the
// run and lib.
func installedLayoutJavaScript(t *testing.T, fixture string) (run, string) {
	t.Helper()
	program := installedLayoutLowered(t, fixture)
	library := layInstalledTypeScript(t, t.TempDir())
	module := filepath.Join(library, "_tsc.mjs")
	if err := os.WriteFile(module, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return runInstalledLayout(t, module, nil, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), module), library
}

// installedLayoutShape is result with every spelling of library in what it wrote made <lib>, so a row can pin a path
// a leg prints by its shape: library is its leg's own, and real, so this is an exact substitution, never a pattern.
func installedLayoutShape(result run, library string) run {
	result.stdout = []byte(strings.ReplaceAll(string(result.stdout), library, "<lib>"))
	result.stderr = []byte(strings.ReplaceAll(string(result.stderr), library, "<lib>"))
	return result
}

// installedLayoutLines is a run's stdout, a line each.
func installedLayoutLines(result run) []string {
	return strings.Split(strings.TrimSuffix(string(result.stdout), "\n"), "\n")
}

// __filename and __dirname are held to Node in tsc's installed layout (#redxes2, Adamic and Platforms on #0ta3dk9) by
// what tsc derives from them, never by the raw path, which differs by the program file's own name: the directory it
// runs from, the default lib beside it, the case probe, the locale's messages and the es2020 lib. Node's lines are
// pinned too, so the two legs can't agree on a layout neither reads. Mutants: tsc_reads.mutant.md.
func TestInstalledLayoutReadsAreNodes(t *testing.T) {
	t.Parallel()
	node, _ := installedLayoutNode(t, installedLayoutReadsFixture)
	if node.exitCode != 0 || len(node.stderr) != 0 {
		t.Fatalf("node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	lines := installedLayoutLines(node)
	if len(lines) != 5 {
		t.Fatalf("node printed %q, want a line for each of five reads", lines)
	}
	for index, want := range installedLayoutReads {
		if lines[index] != want {
			t.Errorf("node's read %d: got %q, want %q", index, lines[index], want)
		}
	}
	built, _ := installedLayoutNative(t, installedLayoutReadsFixture, nil)
	if difference := disagreement(node, built); difference != "" {
		t.Fatalf("%s\nnode: exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q", difference, node.exitCode, node.stdout, node.stderr, built.exitCode, built.stdout, built.stderr)
	}
}

// The layout is checked before anything runs: a lib without ja/ is named as the layout's, so a missing locale
// directory never shows as a wrong locale line. Mutant: ja/'s file taken out of installedLayoutFiles, and this goes
// red, while TestInstalledLayoutReadsAreNodes's pin of Node's locale line goes red as a wrong read.
func TestInstalledLayoutNamesWhatItLacks(t *testing.T) {
	t.Parallel()
	if missing := installedLayoutMissing(layInstalledTypeScript(t, t.TempDir())); len(missing) != 0 {
		t.Fatalf("a fresh layout lacks %v", missing)
	}
	library := layInstalledTypeScript(t, t.TempDir())
	if err := os.Remove(filepath.Join(library, "ja", "diagnosticMessages.generated.json")); err != nil {
		t.Fatal(err)
	}
	if missing := installedLayoutMissing(library); !slices.Equal(missing, []string{"ja/diagnosticMessages.generated.json"}) {
		t.Fatalf("a layout without ja/: got %v missing, want ja/diagnosticMessages.generated.json", missing)
	}
}

// installedLayoutCwdMutant is the runtime answering __filename as _tsc in the working directory and __dirname as the
// working directory, in place of the running image's path and its directory. It swaps the program's two calls of
// node_process.c's functions for functions of its own, defined after the C's includes, so the lowering and the
// runtime archive are unchanged (the generated-C mutant of TestWASIHostBuildBoundaryNodeMutant).
func installedLayoutCwdMutant(t *testing.T, code string) string {
	t.Helper()
	const includes = "#include <stdbool.h>\n"
	if strings.Count(code, includes) != 1 || !strings.Contains(code, "adamic_node_filename()") || !strings.Contains(code, "adamic_node_dirname()") {
		t.Fatal("the cwd mutant's anchors changed")
	}
	const answers = `static adamic_string *mutant_dirname(void) {
    adamic_string *directory = adamic_host_cwd();
    if (directory == NULL) { adamic_panic("mutant: no working directory", sizeof "mutant: no working directory" - 1); }
    return directory;
}
static adamic_string *mutant_filename(void) {
    adamic_string *directory = mutant_dirname();
    adamic_string entry = ADAMIC_STRING("/_tsc");
    adamic_string *path = adamic_string_concat(2, (adamic_string *const[]){directory, &entry});
    adamic_release(directory);
    return path;
}
`
	code = strings.Replace(code, includes, includes+answers, 1)
	code = strings.ReplaceAll(code, "adamic_node_filename()", "mutant_filename()")
	return strings.ReplaceAll(code, "adamic_node_dirname()", "mutant_dirname()")
}

// The cwd mutant, run from this package's directory, outside lib, moves every read made through __dirname: the
// directory, both libs and the locale come out as the working directory's, so TestInstalledLayoutReadsAreNodes's
// comparison goes red on them. The case probe isn't held: on the gate's case-sensitive file system it agrees (neither
// the swapped image path nor _tsc in the working directory exists), and on a Mac's it doesn't. Recorded in
// tsc_reads.mutant.md.
func TestInstalledLayoutCwdMutantMovesTheReads(t *testing.T) {
	t.Parallel()
	reads, _ := installedLayoutNode(t, installedLayoutReadsFixture)
	node := installedLayoutLines(reads)
	result, _ := installedLayoutNative(t, installedLayoutReadsFixture, installedLayoutCwdMutant)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	mutant := installedLayoutLines(result)
	if len(node) != 5 || len(mutant) != 5 {
		t.Fatalf("node printed %q and the mutant %q, want five reads each", node, mutant)
	}
	for _, index := range []int{0, 1, 3, 4} {
		if mutant[index] == node[index] {
			t.Errorf("the cwd mutant's read %d agrees with Node's: %q", index, node[index])
		}
	}
	t.Logf("the cwd mutant prints %q where Node prints %q: TestInstalledLayoutReadsAreNodes goes red on reads 0, 1, 3 and 4", mutant, node)
}

// The ruled path divergences (backend_divergences_test.go's rows declared CommonJS): their fixtures print __filename
// and __dirname raw, so they run here, in the installed layout, on Node, natively and through the JavaScript backend,
// never in the per-file loop, which lowers every fixture as an ES module and runs its source as one. Each leg's lib is
// made <lib>, and the unchanged sourceDisagreement and backendDisagreement hold each to the row's pinned run.
func TestInstalledLayoutPathDivergences(t *testing.T) {
	t.Parallel()
	ruled := 0
	for _, entry := range ruledBackendDivergences {
		if entry.format != lower.ModuleFormatCommonJS {
			continue
		}
		ruled++
		for _, fixture := range fixtures {
			if fixture.path == entry.fixture {
				t.Errorf("%s is among the oracle's fixtures, which lower it as an ES module", entry.fixture)
			}
		}
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			path := strings.TrimPrefix(entry.fixture, "internal/oracle/")
			node := installedLayoutShape(installedLayoutNode(t, path))
			built := installedLayoutShape(installedLayoutNative(t, path, nil))
			backend := installedLayoutShape(installedLayoutJavaScript(t, path))
			for _, difference := range []string{
				sourceDisagreement(entry.fixture, node, built, false),
				sourceDisagreement(entry.fixture, node, backend, true),
				backendDisagreement(entry.fixture, backend, built),
			} {
				if difference != "" {
					t.Errorf("%s\nnode: %q %q\nnative: %q %q\njavascript: %q %q", difference, node.stdout, node.stderr, built.stdout, built.stderr, backend.stdout, backend.stderr)
				}
			}
		})
	}
	if ruled != 2 {
		t.Fatalf("want the two path divergences declared CommonJS, found %d", ruled)
	}
}
