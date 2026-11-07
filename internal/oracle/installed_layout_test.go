package oracle

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
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

// installedLayoutReadsFixture is a program of two modules in two directories, main.ts and the reads it calls in
// sys/sys.ts, that prints what tsc derives from __filename and __dirname, a line per read, and neither raw (Adamic and
// Platforms on #0ta3dk9).
const installedLayoutReadsFixture = "testdata/installed_layout/tsc_reads"

// installedLayoutFiles are the files tsc_reads reads beside its entry, by their path under lib, with what each
// holds: the default lib and the es2020 lib it steps to, and the messages --locale ja loads.
var installedLayoutFiles = map[string]string{
	"lib.d.ts":                             "/// <reference no-default-lib=\"true\"/>\n",
	"lib.es2020.d.ts":                      "/// <reference no-default-lib=\"true\"/>\n",
	"ja/diagnosticMessages.generated.json": `{"Version_0": "バージョン {0}"}`,
}

// installedLayoutReads are the lines Node prints for tsc_reads in the layout, the case probe's aside: it answers
// the file system's case sensitivity, which differs between the gate's Linux and a Mac, and both legs share it.
var installedLayoutReads = map[int]string{
	0: "directory lib",
	1: "lib.d.ts true",
	3: "locale " + installedLayoutFiles["ja/diagnosticMessages.generated.json"],
	4: "lib.es2020.d.ts true",
}

// layInstalledTypeScript lays out node_modules/typescript/lib under root as npm installs it, with installedLayoutFiles
// in lib and the package's own package.json, without a type, so a .js entry there is CommonJS whatever a directory
// above says. It returns lib, where the caller puts the entry, by its real path: every leg's answer for its own path
// is then lib's own spelling, the one installedLayoutShape takes out (a Mac's temporary directory is under a link,
// /var).
func layInstalledTypeScript(t *testing.T, root string) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	library := filepath.Join(root, "node_modules", "typescript", "lib")
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

// installedLayoutEntry is how a leg runs its entry: by the entry's own path, or through the link npm makes for a
// package's command, node_modules/.bin/tsc, as a user's `tsc` is run.
type installedLayoutEntry string

const (
	installedLayoutDirect installedLayoutEntry = "Direct"
	installedLayoutLinked installedLayoutEntry = "Linked"
)

// installedLayoutInvoked is the path a leg runs entry by, in lib: entry itself, or a link to it that it makes at
// node_modules/.bin/tsc, relative as npm writes it (../typescript/lib/<entry>).
func installedLayoutInvoked(t *testing.T, entry string, how installedLayoutEntry) string {
	t.Helper()
	if how == installedLayoutDirect {
		return entry
	}
	commands := filepath.Join(filepath.Dir(entry), "..", "..", ".bin")
	if err := os.MkdirAll(commands, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(commands, "tsc")
	if err := os.Symlink(filepath.Join("..", "typescript", "lib", filepath.Base(entry)), link); err != nil {
		t.Fatal(err)
	}
	return link
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

// installedLayoutPacking is how Node runs a program's modules: Bundled, as tsc ships, every module wrapped into one
// lib/_tsc.js, so every module's __filename and __dirname are the bundle's; or PerFile, each module transpiled to a
// .js of its own beside the others, so each module's are its own.
type installedLayoutPacking string

const (
	installedLayoutBundled installedLayoutPacking = "Bundled"
	installedLayoutPerFile installedLayoutPacking = "PerFile"
)

// installedLayoutPerFileFixture is Language's fixture for the per-file difference, the one ruled row Node runs a
// file per module (Adamic's ruling on #0ta3dk9).
const installedLayoutPerFileFixture = "internal/oracle/testdata/installed_layout/per_file"

// installedLayoutModules is a fixture's modules: a directory's every .ts, its entry main.ts, or a file alone, its own
// entry. It returns the directory they're found from, the entry and every module, each a path from that directory
// with / between directories, sorted.
func installedLayoutModules(t *testing.T, fixture string) (string, string, []string) {
	t.Helper()
	information, err := os.Stat(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if !information.IsDir() {
		return filepath.Dir(fixture), filepath.Base(fixture), []string{filepath.Base(fixture)}
	}
	var modules []string
	err = filepath.WalkDir(fixture, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".ts" {
			return err
		}
		name, err := filepath.Rel(fixture, path)
		modules = append(modules, filepath.ToSlash(name))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(modules)
	return fixture, "main.ts", modules
}

// installedLayoutNode runs fixture, a path from this package, on Node as tsc ships: each module transpiled to
// CommonJS by TypeScript 6.0.3's own transpileModule (programNodeAPI) and packed as packing says, by
// testdata/installed_layout/transpile.cjs, with no bundler, its entry run as how says. Bundled, the entry is
// lib/_tsc.js; a file each, lib/main.js. It returns the run and lib.
func installedLayoutNode(t *testing.T, fixture string, how installedLayoutEntry, packing installedLayoutPacking) (run, string) {
	t.Helper()
	api := filepath.Join(programNodeAPI(t), "lib/typescript.js")
	root, entry, modules := installedLayoutModules(t, fixture)
	library := layInstalledTypeScript(t, t.TempDir())
	tsc := filepath.Join(library, "_tsc.js")
	mode, out := "bundle", tsc
	if packing == installedLayoutPerFile {
		tsc = filepath.Join(library, strings.TrimSuffix(entry, ".ts")+".js")
		mode, out = "files", library
	}
	arguments := append([]string{"testdata/installed_layout/transpile.cjs", api, mode, root, entry, out}, modules...)
	if result := execute(t, "node", arguments...); result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("transpiling %s to CommonJS: exit %d, stderr %q", fixture, result.exitCode, result.stderr)
	}
	// NODE_PATH would add a directory to where Node looks, which the installed layout doesn't have.
	return runInstalledLayout(t, tsc, []string{"NODE_PATH="}, "node", installedLayoutInvoked(t, tsc, how)), library
}

// installedLayoutLowered is fixture, a path from this package, lowered from its entry as a build that declares one
// bundled image lowers it (lower.ModuleFormatCommonJSBundle): load.Load of the entry pulls in every module it imports.
func installedLayoutLowered(t *testing.T, fixture string) *ir.Program {
	t.Helper()
	root, entry, _ := installedLayoutModules(t, fixture)
	path, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(entry)))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := lower.LowerWithOptions(context.Background(), loaded, lower.Options{ModuleFormat: lower.ModuleFormatCommonJSBundle})
	if err != nil {
		t.Fatal(err)
	}
	return program
}

// installedLayoutNative runs fixture built as a user's build is, declared one bundled image, as lib/_tsc, run as how
// says.
// mutate rewrites the program's C before it's built, or is nil. It returns the run and lib.
func installedLayoutNative(t *testing.T, fixture string, how installedLayoutEntry, mutate func(t *testing.T, code string) string) (run, string) {
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
	invoked := installedLayoutInvoked(t, tsc, how)
	return runInstalledLayout(t, tsc, nil, invoked), library
}

// installedLayoutJavaScript runs fixture through the JavaScript backend, declared one bundled image, as
// lib/_tsc.mjs: the emitted program is an ES module, and the package's own package.json has no type, so only .mjs
// runs it as one. It runs under oracle/node.mjs, as onJavaScriptBackend does, so its import of 'adamic' finds the
// runtime. It returns the run and lib.
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
	installedLayoutReadsAgree(t, installedLayoutDirect)
}

// installedLayoutReadsAgree runs tsc_reads on Node, bundled, and natively, each as how says, holds Node's lines to
// installedLayoutReads and native's output to Node's, byte for byte.
func installedLayoutReadsAgree(t *testing.T, how installedLayoutEntry) {
	t.Helper()
	node, _ := installedLayoutNode(t, installedLayoutReadsFixture, how, installedLayoutBundled)
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
	built, _ := installedLayoutNative(t, installedLayoutReadsFixture, how, nil)
	if difference := disagreement(node, built); difference != "" {
		t.Fatalf("%s\nnode: exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q", difference, node.exitCode, node.stdout, node.stderr, built.exitCode, built.stdout, built.stderr)
	}
}

// npm runs tsc through node_modules/.bin/tsc, a link to the package's file, and Node resolves the main module through
// it, so __filename is the real _tsc.js and every read is as it is run directly: Node's own lines through the link
// are held to installedLayoutReads, which proves it resolved the link. Native resolves its image as Node does
// (/proc/self/exe on Linux, realpath on darwin), so its reads through the link are held to Node's byte for byte.
// Mutant: installedLayoutInvokedMutant, recorded in tsc_reads.mutant.md.
func TestInstalledLayoutReadsThroughTheLinkAreNodes(t *testing.T) {
	t.Parallel()
	installedLayoutReadsAgree(t, installedLayoutLinked)
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

// installedLayoutPlant swaps the program's calls of node_process.c's adamic_node_filename and adamic_node_dirname for
// answers' mutant_filename and mutant_dirname, defined after the C's includes, so the lowering and the runtime archive
// are unchanged (the generated-C mutant of TestWASIHostBuildBoundaryNodeMutant).
func installedLayoutPlant(t *testing.T, code, answers string) string {
	t.Helper()
	const includes = "#include <stdbool.h>\n"
	if strings.Count(code, includes) != 1 || !strings.Contains(code, "adamic_node_filename()") || !strings.Contains(code, "adamic_node_dirname()") {
		t.Fatal("the mutant's anchors changed")
	}
	code = strings.Replace(code, includes, includes+answers, 1)
	code = strings.ReplaceAll(code, "adamic_node_filename()", "mutant_filename()")
	return strings.ReplaceAll(code, "adamic_node_dirname()", "mutant_dirname()")
}

// installedLayoutCwdMutant is the runtime answering __filename as _tsc in the working directory and __dirname as the
// working directory, in place of the running image's path and its directory.
func installedLayoutCwdMutant(t *testing.T, code string) string {
	t.Helper()
	return installedLayoutPlant(t, code, `static adamic_string *mutant_dirname(void) {
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
`)
}

// installedLayoutInvokedMutant is the runtime answering __filename as the path it was invoked by, not the image that
// path resolves to, and __dirname as that path's directory: argv[0], the first string of /proc/self/cmdline.
func installedLayoutInvokedMutant(t *testing.T, code string) string {
	t.Helper()
	return installedLayoutPlant(t, code, `static adamic_string *mutant_filename(void) {
    FILE *file = fopen("/proc/self/cmdline", "rb");
    if (file == NULL) { adamic_panic("mutant: no /proc/self/cmdline", sizeof "mutant: no /proc/self/cmdline" - 1); }
    char bytes[4096];
    size_t length = fread(bytes, 1, sizeof bytes, file);
    fclose(file);
    size_t invoked = 0;
    while (invoked < length && bytes[invoked] != '\0') { invoked++; }
    return adamic_decode_utf8((const unsigned char *)bytes, invoked);
}
static adamic_string *mutant_dirname(void) {
    adamic_string *path = mutant_filename();
    adamic_string *directory = adamic_node_path_dirname(path);
    adamic_release(path);
    return directory;
}
`)
}

// The cwd mutant, run from this package's directory, outside lib, moves every read made through __dirname: the
// directory, both libs and the locale come out as the working directory's, so TestInstalledLayoutReadsAreNodes's
// comparison goes red on them. The case probe isn't held: on the gate's case-sensitive file system it agrees (neither
// the swapped image path nor _tsc in the working directory exists), and on a Mac's it doesn't. Recorded in
// tsc_reads.mutant.md.
func TestInstalledLayoutCwdMutantMovesTheReads(t *testing.T) {
	t.Parallel()
	installedLayoutMutantMoves(t, "cwd", installedLayoutDirect, installedLayoutCwdMutant, "TestInstalledLayoutReadsAreNodes")
}

// installedLayoutMutantMoves runs tsc_reads on Node, bundled, and a mutant's build, each as how says, and holds that
// the mutant's four reads through __dirname, the directory, both libs and the locale, differ from Node's, so
// control, the test comparing the two, goes red on them.
func installedLayoutMutantMoves(t *testing.T, name string, how installedLayoutEntry, mutate func(t *testing.T, code string) string, control string) {
	t.Helper()
	reads, _ := installedLayoutNode(t, installedLayoutReadsFixture, how, installedLayoutBundled)
	node := installedLayoutLines(reads)
	result, _ := installedLayoutNative(t, installedLayoutReadsFixture, how, mutate)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	mutant := installedLayoutLines(result)
	if len(node) != 5 || len(mutant) != 5 {
		t.Fatalf("node printed %q and the mutant %q, want five reads each", node, mutant)
	}
	for _, index := range []int{0, 1, 3, 4} {
		if mutant[index] == node[index] {
			t.Errorf("the %s mutant's read %d agrees with Node's: %q", name, index, node[index])
		}
	}
	t.Logf("the %s mutant prints %q where Node prints %q: %s goes red on reads 0, 1, 3 and 4", name, mutant, node, control)
}

// The invoked-path mutant, run through node_modules/.bin/tsc by its absolute path, answers that link's path and its
// directory, .bin, where Node resolves the link to lib: the directory, both libs and the locale come out as .bin's, so
// TestInstalledLayoutReadsThroughTheLinkAreNodes goes red on them. It reads /proc, so it runs on Linux, the gate.
// Recorded in tsc_reads.mutant.md.
func TestInstalledLayoutInvokedPathMutantMovesTheReads(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("the mutant reads argv[0] from Linux's /proc/self/cmdline")
	}
	installedLayoutMutantMoves(t, "invoked-path", installedLayoutLinked, installedLayoutInvokedMutant, "TestInstalledLayoutReadsThroughTheLinkAreNodes")
}

// The ruled path divergences (backend_divergences_test.go's rows declared one bundled image): their fixtures print
// __filename and __dirname raw, so they run here, in the installed layout, on Node, natively and through the
// JavaScript backend, never in the per-file loop, which lowers every fixture as an ES module and runs its source as
// one. Node runs each bundled, as tsc ships, but for Language's per-file fixture, which it runs a file per module, the
// difference that row rules. Every leg writes its whole program under its own lib, so one substitution, of that leg's
// lib by <lib>, shapes it, and the unchanged sourceDisagreement and backendDisagreement hold each to the row's pinned
// run.
func TestInstalledLayoutPathDivergences(t *testing.T) {
	t.Parallel()
	ruled := 0
	for _, entry := range ruledBackendDivergences {
		if entry.format != lower.ModuleFormatCommonJSBundle {
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
			packing := installedLayoutBundled
			if entry.fixture == installedLayoutPerFileFixture {
				packing = installedLayoutPerFile
			}
			node := installedLayoutShape(installedLayoutNode(t, path, installedLayoutDirect, packing))
			built := installedLayoutShape(installedLayoutNative(t, path, installedLayoutDirect, nil))
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
	if ruled != 3 {
		t.Fatalf("want the three path divergences declared one bundled image, found %d", ruled)
	}
}
