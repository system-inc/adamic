package lowered

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// corpus is a handful of the oracle's lowering corpus, by repository path, each program's entry first and then the files
// its imports read: a -0 constant (enums.a), async frames, Node's declarations, and two programs of two modules.
var corpus = [][]string{
	{"internal/oracle/testdata/enums.a"},
	{"internal/oracle/testdata/async_closures.a"},
	{"internal/oracle/testdata/node_options_widening.a"},
	{"internal/oracle/testdata/enums_modules/main.a", "internal/oracle/testdata/enums_modules/values.a"},
	{"internal/oracle/testdata/namespaces_debug_modules/main.a", "internal/oracle/testdata/namespaces_debug_modules/values.a"},
}

// fromHere is repository paths as this package's tests name them, from its directory.
func fromHere(paths []string) []string {
	here := make([]string, len(paths))
	for index, path := range paths {
		here[index] = filepath.Join("..", "..", filepath.FromSlash(path))
	}
	return here
}

// privateCache gives the test a build cache of its own, as buildcache's tests take one: nothing built here reaches the
// machine's cache, a runner's read-only one never refuses it, and no store or trace sees it.
func privateCache(t *testing.T) {
	t.Helper()
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	t.Setenv("ADAMIC_BUILD_CACHE", "")
	t.Setenv(location.OutputDirectoryVariable, "")
	t.Setenv("ADAMIC_BUILD_STORE", "off")
	t.Setenv("ADAMIC_BUILD_TRACE", "")
}

// counted is build, counting each run in builds.
func counted(builds *atomic.Int32) func(string, sources, Options) error {
	return func(directory string, read sources, options Options) error {
		builds.Add(1)
		return build(directory, read, options)
	}
}

// freshC is the C of the program at entry, a repository path, loaded where it sits and lowered with options in this
// process, as a caller lowers without the cache.
func freshC(t *testing.T, entry string, options Options) string {
	t.Helper()
	program, err := load.Load(fromHere([]string{entry}))
	if err != nil {
		t.Fatal(err)
	}
	if options.TSGo {
		program.EnableTSGo()
	}
	lowering, err := lower.LowerWithOptions(context.Background(), program, options.Lower)
	if err != nil {
		t.Fatal(err)
	}
	return native.CWithOptions(lowering, native.Options{Debug: options.Lower.Debug})
}

// sameC fails t when got isn't want, naming the first line they differ at.
func sameC(t *testing.T, what, want, got string) {
	t.Helper()
	if got == want {
		return
	}
	wantLines, gotLines := strings.Split(want, "\n"), strings.Split(got, "\n")
	for index := range max(len(wantLines), len(gotLines)) {
		var wanted, had string
		if index < len(wantLines) {
			wanted = wantLines[index]
		}
		if index < len(gotLines) {
			had = gotLines[index]
		}
		if wanted != had {
			t.Fatalf("%s emits other C than a fresh lowering, from line %d:\nfresh:  %s\n%s: %s", what, index+1, wanted, what, had)
		}
	}
}

// keyOf is the key inputs address, on this tree.
func keyOf(t *testing.T, inputs buildcache.Inputs) string {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	key, err := buildcache.Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// The same sources and options under the same environment key the same, and each input the lowering reads moves the
// key alone: a byte of a source, each of lower's options and TSGo, each keyed variable, and the working directory.
func TestTheKeyMovesWithEachInputTheLoweringReads(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "enums.a")
	original, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "enums.a"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, original, 0o644); err != nil {
		t.Fatal(err)
	}
	quiet := environment{getenv: func(string) string { return "" }, workingDirectory: directory}
	before := keyOf(t, inputsOf(t, sourcesOf(t, []string{entry}), Options{}, quiet))
	if again := keyOf(t, inputsOf(t, sourcesOf(t, []string{entry}), Options{}, quiet)); again != before {
		t.Fatalf("one program, one set of options, two keys: %s and %s", before[:12], again[:12])
	}
	setting := func(variable string) environment {
		return environment{getenv: func(name string) string {
			if name == variable {
				return "1"
			}
			return ""
		}, workingDirectory: directory}
	}
	for _, moved := range []struct {
		what        string
		options     Options
		environment environment
	}{
		{"Debug", Options{Lower: lower.Options{Debug: true}}, quiet},
		{"ProgramRegion", Options{Lower: lower.Options{ProgramRegion: true}}, quiet},
		{"OwnershipQuery", Options{Lower: lower.Options{OwnershipQuery: true}}, quiet},
		{"TSGo", Options{TSGo: true}, quiet},
		{"ADAMIC_PROGRAM_REGION", Options{}, setting("ADAMIC_PROGRAM_REGION")},
		{"ADAMIC_OWNERSHIP_QUERY", Options{}, setting("ADAMIC_OWNERSHIP_QUERY")},
		{"GOFLAGS", Options{}, setting("GOFLAGS")},
		{"the working directory", Options{}, environment{getenv: quiet.getenv, workingDirectory: filepath.Dir(directory)}},
	} {
		if keyOf(t, inputsOf(t, sourcesOf(t, []string{entry}), moved.options, moved.environment)) == before {
			t.Errorf("a change to %s kept the key", moved.what)
		}
	}
	edited := slices.Clone(original)
	edited[len(edited)-2] ^= 1
	if err := os.WriteFile(entry, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	if keyOf(t, inputsOf(t, sourcesOf(t, []string{entry}), Options{}, quiet)) == before {
		t.Error("a change to a byte of the source kept the key")
	}
}

// A change confined to internal/native or internal/javascript keeps stage 1's key: neither is in the closure go list
// computes, and no file under either is among the key's. The closure holds what it must, so the absence isn't of
// everything: the loader, lowering, the IR, the prelude, the checker's libraries and this package's own build. Mutant:
// internal/native among the lowering's packages.
func TestTheKeyLeavesOutNativeAndJavaScript(t *testing.T) {
	t.Parallel()
	lowering, err := loweringClosure()
	if err != nil {
		t.Fatal(err)
	}
	inputs := inputsOf(t, sources{entry: "main.ts", files: map[string][]byte{"main.ts": nil}}, Options{}, environment{getenv: os.Getenv})
	if !slices.Equal(inputs.Files, lowering.files) {
		t.Fatalf("the key's files are not the closure's: %d and %d", len(inputs.Files), len(lowering.files))
	}
	for _, excluded := range []string{"internal/native", "internal/javascript"} {
		for _, file := range inputs.Files {
			if file == excluded || strings.HasPrefix(file, excluded+"/") {
				t.Errorf("the key names %s, under %s", file, excluded)
			}
		}
		for _, item := range lowering.packages {
			if item.ImportPath == modulePath+"/"+excluded || strings.HasPrefix(item.ImportPath, modulePath+"/"+excluded+"/") {
				t.Errorf("the lowering's closure holds %s", item.ImportPath)
			}
		}
	}
	for _, want := range []string{"go.mod", "go.work", "internal/load/load.go", "internal/lower/lower.go", "internal/ir/gob.go",
		"internal/load/prelude.d.ts", "cohere/TypeScript/tsc/internal/bundled/libs/lib.es2024.d.ts", "internal/lowered/lowered.go"} {
		if !slices.Contains(inputs.Files, want) {
			t.Errorf("the key doesn't name %s", want)
		}
	}
	if slices.Contains(inputs.Files, "internal/lowered/lowered_test.go") {
		t.Error("the key names this package's tests, which build nothing it holds")
	}
	t.Logf("%d packages, %d files in the tree and %d outside it", len(lowering.packages), len(lowering.files), len(lowering.external))
}

// A stored IR read back emits the C a fresh lowering emits, byte for byte, the build's own decode and a hit's alike,
// and the second read is a hit: the program built once.
func hitEmitsFreshC(t *testing.T, debug bool) {
	t.Helper()
	privateCache(t)
	options := Options{Lower: lower.Options{Debug: debug}}
	for _, paths := range corpus {
		t.Run(paths[0], func(t *testing.T) {
			var builds atomic.Int32
			read := sourcesOf(t, fromHere(paths))
			missed, missedInputs, first := stored(t, read, options, counted(&builds))
			hit, hitInputs, second := stored(t, read, options, counted(&builds))
			if builds.Load() != 1 || first != second {
				t.Fatalf("two reads built %d times, from %s and %s", builds.Load(), first, second)
			}
			if keyOf(t, missedInputs) != keyOf(t, hitInputs) {
				t.Fatal("two reads of one program keyed it apart")
			}
			if missed == hit {
				t.Fatal("two reads served one decode: each emit needs its own")
			}
			want := freshC(t, paths[0], options)
			sameC(t, "the build's decode", want, native.CWithOptions(missed, native.Options{Debug: debug}))
			sameC(t, "a hit", want, native.CWithOptions(hit, native.Options{Debug: debug}))
		})
	}
}

// Not parallel: privateCache sets the build cache's variables for the process.
func TestAHitEmitsTheCOfAFreshLoweringRelease(t *testing.T) {
	hitEmitsFreshC(t, false)
}

// Not parallel: privateCache sets the build cache's variables for the process.
func TestAHitEmitsTheCOfAFreshLoweringDebug(t *testing.T) {
	hitEmitsFreshC(t, true)
}

// tornProgram is the program the miss tests store and then tear: it prints 'outside', a string gob writes as it is.
var tornProgram = corpus[4]

// replaced stores data, then stored's next read of the program: a miss, so it built again into another directory, and
// the IR it serves emits the C a fresh lowering does; a read after that finds the rebuilt product.
func replaced(t *testing.T, first string, read sources, builds *atomic.Int32, data, sum []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(first, gobFile), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(first, sumFile), sum, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readStored(first); err == nil {
		t.Fatal("the stored IR reads back")
	} else {
		t.Logf("the stored IR doesn't read back: %v", err)
	}
	program, _, second := stored(t, read, Options{}, counted(builds))
	if second == first || builds.Load() != 2 {
		t.Fatalf("served from %s, built %d times: the stored IR was served", second, builds.Load())
	}
	sameC(t, "the rebuilt IR", freshC(t, tornProgram[0], Options{}), native.C(program))
	if _, _, third := stored(t, read, Options{}, counted(builds)); third != second || builds.Load() != 2 {
		t.Fatalf("a read after the rebuild built again (%d builds), from %s", builds.Load(), third)
	}
}

// A stored IR whose bytes aren't the ones its sum names is a miss and rebuilt, never served, though gob would decode
// it: one of its strings changed, the same length. Mutant: the sum check dropped.
//
// Not parallel: privateCache sets the build cache's variables for the process.
func TestATornStoredIRIsAMissAndRebuilt(t *testing.T) {
	privateCache(t)
	read := sourcesOf(t, fromHere(tornProgram))
	var builds atomic.Int32
	_, _, first := stored(t, read, Options{}, counted(&builds))
	data, err := os.ReadFile(filepath.Join(first, gobFile))
	if err != nil {
		t.Fatal(err)
	}
	sum, err := os.ReadFile(filepath.Join(first, sumFile))
	if err != nil {
		t.Fatal(err)
	}
	at := bytes.Index(data, []byte("outside"))
	if at < 0 {
		t.Fatalf("%s holds no 'outside' to tear", tornProgram[0])
	}
	torn := slices.Clone(data)
	copy(torn[at:], "OUTSIDE")
	replaced(t, first, read, &builds, torn, sum)
}

// A stored IR gob can't decode is a miss and rebuilt, its sum right for the bytes it holds.
//
// Not parallel: privateCache sets the build cache's variables for the process.
func TestAnUndecodableStoredIRIsAMissAndRebuilt(t *testing.T) {
	privateCache(t)
	read := sourcesOf(t, fromHere(tornProgram))
	var builds atomic.Int32
	_, _, first := stored(t, read, Options{}, counted(&builds))
	garbage := []byte("not a gob")
	replaced(t, first, read, &builds, garbage, fmt.Appendf(nil, "%x\n", sha256.Sum256(garbage)))
}

// An environment read is a variable some file of the closure reads, and where.
type environmentRead struct{ variable, where string }

// environmentReads is every os.Getenv, os.LookupEnv and os.Environ (and syscall's) in file: by the variable it names,
// or "" for a read whose name isn't a string literal, or os.Environ, which reads them all.
func environmentReads(t *testing.T, file string) []environmentRead {
	t.Helper()
	positions := token.NewFileSet()
	parsed, err := parser.ParseFile(positions, file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	packages := map[string]bool{}
	for _, spec := range parsed.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || (path != "os" && path != "syscall") {
			continue
		}
		local := path
		if spec.Name != nil {
			local = spec.Name.Name
		}
		packages[local] = true
	}
	reading := func(node ast.Node) (*ast.SelectorExpr, bool) {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return nil, false
		}
		identifier, ok := selector.X.(*ast.Ident)
		return selector, ok && packages[identifier.Name] && slices.Contains([]string{"Getenv", "LookupEnv", "Environ"}, selector.Sel.Name)
	}
	var reads []environmentRead
	called := map[*ast.SelectorExpr]bool{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if selector, ok := reading(call.Fun); ok {
				called[selector] = true
				variable := ""
				if len(call.Args) > 0 && selector.Sel.Name != "Environ" {
					if literal, ok := call.Args[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
						variable, _ = strconv.Unquote(literal.Value)
					}
				}
				reads = append(reads, environmentRead{variable, positions.Position(call.Pos()).String()})
			}
		}
		if selector, ok := reading(node); ok && !called[selector] {
			// A read taken as a value, or os.Environ called: no name the key could hold.
			reads = append(reads, environmentRead{"", positions.Position(selector.Pos()).String()})
		}
		return true
	})
	return reads
}

// Every environment variable the lowering's closure reads, in each package of this module go list -deps
// ./internal/load ./internal/lower names, as Linux and macOS build it, is keyed by its value or said to shape nothing the
// IR holds, so a new read can't leave the key silently short; a read by a name no source spells fails, and so does a
// listed variable nothing reads. Today's reads: ADAMIC_PROGRAM_REGION and ADAMIC_OWNERSHIP_QUERY (lower.Lower), keyed,
// and ADAMIC_BUILD_LOG (buildlog.Phase), not. Mutant: an os.Getenv planted in internal/lower with no key line.
func TestEveryEnvironmentReadOfTheLoweringIsKeyed(t *testing.T) {
	t.Parallel()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]bool{}
	for _, system := range []string{"linux", "darwin"} {
		packages, err := listed(root, []string{"GOOS=" + system, "CGO_ENABLED=1"}, true, loweringEntry...)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range packages {
			if item.Module == nil || item.Module.Path != modulePath {
				continue
			}
			for _, file := range append(slices.Clone(item.GoFiles), item.CgoFiles...) {
				files[filepath.Join(item.Dir, file)] = true
			}
		}
	}
	read := map[string][]string{}
	for file := range files {
		for _, found := range environmentReads(t, file) {
			relative, err := filepath.Rel(root, found.where)
			if err != nil {
				relative = found.where
			}
			if found.variable == "" {
				t.Errorf("%s reads the environment by a name its source doesn't spell, which no key can hold", relative)
				continue
			}
			read[found.variable] = append(read[found.variable], filepath.ToSlash(relative))
		}
	}
	keyed := inputsOf(t, sources{entry: "main.ts", files: map[string][]byte{"main.ts": nil}}, Options{}, environment{getenv: func(variable string) string { return "set-" + variable }})
	for variable, where := range read {
		_, explained := unkeyedEnvironment[variable]
		switch {
		case slices.Contains(keyedEnvironment, variable):
			if !slices.Contains(keyed.Flags, variable+"=set-"+variable) {
				t.Errorf("%s is listed as keyed and the key holds no line of its value", variable)
			}
		case !explained:
			t.Errorf("%s reads %s, which the key doesn't hold: key it (keyedEnvironment) or say why it shapes nothing the IR holds (unkeyedEnvironment)", strings.Join(where, ", "), variable)
		}
	}
	listedVariables := slices.Clone(keyedEnvironment)
	for variable := range unkeyedEnvironment {
		listedVariables = append(listedVariables, variable)
	}
	for _, variable := range listedVariables {
		if len(read[variable]) == 0 {
			t.Errorf("%s is listed, and nothing in the closure reads it", variable)
		}
	}
	variables := make([]string, 0, len(read))
	for variable, where := range read {
		variables = append(variables, variable+" ("+strings.Join(where, ", ")+")")
	}
	sort.Strings(variables)
	t.Logf("%d files read %s", len(files), strings.Join(variables, ", "))
}

// The checker's libraries, Node's declarations and the prelude are in the binary, never read from disk while a program
// loads: the checker is built embedding its libraries, every file a program loads that isn't one of its sources is at a
// bundled path or is the prelude, and the files they're embedded from are among the key's.
func TestTheLibrariesAndThePreludeAreInTheBinary(t *testing.T) {
	t.Parallel()
	if !bundled.Embedded {
		t.Fatal("the checker reads its libraries from beside the executable (a noembed build): their bytes would need a place in the key")
	}
	for _, paths := range [][]string{corpus[0], corpus[2]} {
		read := sourcesOf(t, fromHere(paths))
		directory := t.TempDir()
		program, err := loaded(directory, read, Options{})
		if err != nil {
			t.Fatal(err)
		}
		counts := map[string]int{}
		for _, file := range program.CompilerProgram().GetSourceFiles() {
			path := file.FileName().AsString()
			switch {
			case load.IsPrelude(file):
				counts["prelude"]++
			case load.IsLibrary(file):
				counts["library"]++
				if !strings.HasPrefix(path, bundledScheme) {
					t.Errorf("%s: the library %s is on disk", paths[0], path)
				}
			case strings.HasPrefix(path, bundledScheme+"node/"):
				counts["node"]++
			case strings.Contains(path, "/source/"):
				counts["source"]++
			default:
				t.Errorf("%s loads %s from disk", paths[0], path)
			}
		}
		if counts["prelude"] != 1 || counts["library"] == 0 || counts["source"] != len(paths) {
			t.Errorf("%s loads %v: want the prelude, the libraries and its own %d", paths[0], counts, len(paths))
		}
		if paths[0] == corpus[2][0] && counts["node"] == 0 {
			t.Errorf("%s, which imports node:fs, loads none of Node's declarations", paths[0])
		}
		t.Logf("%s loads %v", paths[0], counts)
	}
	lowering, err := loweringClosure()
	if err != nil {
		t.Fatal(err)
	}
	for _, embedded := range []string{"internal/load/prelude.d.ts", "internal/load/host_axioms.json",
		"internal/load/node_types/node_modules/@types/node/index.d.ts", "cohere/TypeScript/tsc/internal/bundled/libs/lib.es2024.d.ts"} {
		if !slices.Contains(lowering.files, embedded) {
			t.Errorf("%s, embedded in the lowering, isn't among the key's files", embedded)
		}
	}
}

// A program that reads a file from disk that isn't among its sources is refused, never stored under a key that
// doesn't name the file.
func TestAProgramReadingAFileBesideItsSourcesIsRefused(t *testing.T) {
	t.Parallel()
	elsewhere := filepath.Join(t.TempDir(), "values.ts")
	if err := os.WriteFile(elsewhere, []byte("export const value = 'beside';\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, fmt.Appendf(nil, "import { value } from %q;\nconsole.log(value);\n", elsewhere), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loaded(t.TempDir(), sourcesOf(t, []string{entry}), Options{}); err == nil || !strings.Contains(err.Error(), "isn't one of its sources") {
		t.Fatalf("got %v, want the read of %s refused", err, elsewhere)
	}
}
