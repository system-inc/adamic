package lowered

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
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

// freshC is the C of the program at paths' entry, loaded from disk where it sits and lowered with options in this
// process, as a caller lowers without the cache.
func freshC(t *testing.T, entry string, options Options) string {
	t.Helper()
	program, err := load.Load([]string{entry})
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

// keyAt is the key inputs address on the tree at root.
func keyAt(t *testing.T, root string, inputs buildcache.Inputs) string {
	t.Helper()
	key, err := buildcache.Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// here is this process's environment with no variable set.
func here(t *testing.T) environment {
	t.Helper()
	under := processEnvironment(t)
	under.getenv = func(string) string { return "" }
	return under
}

// The same sources and options under the same environment key the same, and each input the lowering reads moves the
// key alone: a byte of a source, where the sources sit, each of lower's options and TSGo, each keyed variable, GOFLAGS
// and the working directory.
func TestTheKeyMovesWithEachInputTheLoweringReads(t *testing.T) {
	t.Parallel()
	quiet := here(t)
	read := sourcesOf(t, fromHere(corpus[0]))
	before := keyAt(t, quiet.root, inputsOf(t, read, Options{}, quiet))
	if again := keyAt(t, quiet.root, inputsOf(t, sourcesOf(t, fromHere(corpus[0])), Options{}, quiet)); again != before {
		t.Fatalf("one program, one set of options, two keys: %s and %s", before[:12], again[:12])
	}
	setting := func(variable string) environment {
		under := quiet
		under.getenv = func(name string) string {
			if name == variable {
				return "1"
			}
			return ""
		}
		return under
	}
	elsewhere := quiet
	elsewhere.workingDirectory = filepath.Dir(quiet.workingDirectory)
	edited := sources{base: read.base, entry: read.entry, files: map[string][]byte{read.entry: slices.Clone(read.files[read.entry])}}
	edited.files[read.entry][len(edited.files[read.entry])-2] ^= 1
	moved := read
	moved.base = filepath.Join(read.base, "enums_modules")
	for _, change := range []struct {
		what        string
		read        sources
		options     Options
		environment environment
	}{
		{"a byte of the source", edited, Options{}, quiet},
		{"where the sources sit", moved, Options{}, quiet},
		{"Debug", read, Options{Lower: lower.Options{Debug: true}}, quiet},
		{"ProgramRegion", read, Options{Lower: lower.Options{ProgramRegion: true}}, quiet},
		{"OwnershipQuery", read, Options{Lower: lower.Options{OwnershipQuery: true}}, quiet},
		{"TSGo", read, Options{TSGo: true}, quiet},
		{"ADAMIC_PROGRAM_REGION", read, Options{}, setting("ADAMIC_PROGRAM_REGION")},
		{"ADAMIC_OWNERSHIP_QUERY", read, Options{}, setting("ADAMIC_OWNERSHIP_QUERY")},
		{"TS_GO_DEBUG_STACK_LIMIT", read, Options{}, setting("TS_GO_DEBUG_STACK_LIMIT")},
		{"GOFLAGS", read, Options{}, setting("GOFLAGS")},
		{"the working directory", read, Options{}, elsewhere},
	} {
		if keyAt(t, quiet.root, inputsOf(t, change.read, change.options, change.environment)) == before {
			t.Errorf("a change to %s kept the key", change.what)
		}
	}
}

// optionLeaves are the paths of every field Options holds, a struct's own fields standing for it, with each one's
// value in options.
func optionLeaves(value reflect.Value, prefix string) map[string]reflect.Value {
	leaves := map[string]reflect.Value{}
	for index := range value.NumField() {
		field, name := value.Field(index), prefix+value.Type().Field(index).Name
		if field.Kind() == reflect.Struct {
			for leaf, held := range optionLeaves(field, name+".") {
				leaves[leaf] = held
			}
			continue
		}
		leaves[name] = field
	}
	return leaves
}

// Every field of Options and of lower.Options is keyed by its value, by a line of its own: the key's lines name exactly
// the fields the structs have, and setting any one field alone changes its line. A field added to either struct fails
// here until optionLines keys it. Mutant: a field's line dropped.
func TestEveryOptionIsKeyedByValue(t *testing.T) {
	t.Parallel()
	var zero Options
	lines := optionLines(zero)
	named := make([]string, 0, len(lines))
	for _, line := range lines {
		name, _, _ := strings.Cut(line, "=")
		named = append(named, name)
	}
	leaves := optionLeaves(reflect.ValueOf(&zero).Elem(), "")
	fields := make([]string, 0, len(leaves))
	for field := range leaves {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	sort.Strings(named)
	if !slices.Equal(fields, named) {
		t.Fatalf("Options' fields are %v, and the key's lines name %v", fields, named)
	}
	for _, field := range fields {
		var set Options
		leaf := optionLeaves(reflect.ValueOf(&set).Elem(), "")[field]
		switch leaf.Kind() {
		case reflect.Bool:
			leaf.SetBool(true)
		default:
			t.Fatalf("Options.%s is a %s: give this test a value to set it to", field, leaf.Kind())
		}
		if slices.Equal(optionLines(set), lines) {
			t.Errorf("setting Options.%s alone kept the key's lines", field)
		}
	}
}

// Two Options built apart with equal fields key alike.
func TestTwoEqualOptionsKeyAlike(t *testing.T) {
	t.Parallel()
	quiet := here(t)
	read := sourcesOf(t, fromHere(corpus[0]))
	first := Options{Lower: lower.Options{Debug: true, OwnershipQuery: true}}
	var second Options
	second.TSGo = false
	second.Lower.OwnershipQuery = true
	second.Lower.Debug = true
	if keyAt(t, quiet.root, inputsOf(t, read, first, quiet)) != keyAt(t, quiet.root, inputsOf(t, read, second, quiet)) {
		t.Fatal("two equal Options keyed apart")
	}
}

// copyInto copies root's file at name into another root.
func copyInto(t *testing.T, root, other, name string) {
	t.Helper()
	from, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	defer from.Close()
	path := filepath.Join(other, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	to, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer to.Close()
	if _, err := io.Copy(to, from); err != nil {
		t.Fatal(err)
	}
}

// Two checkouts of one tree key alike: every place in the key is named from the checkout's root, the sources' and the
// working directory's both. Each stand-in checkout holds the key's files and the program, copied from this one, and the
// key's environment names its own root. Mutant: the working directory keyed as it is.
func TestTwoCheckoutsOfOneTreeKeyAlike(t *testing.T) {
	t.Parallel()
	this := here(t)
	lowering, err := loweringClosure()
	if err != nil {
		t.Fatal(err)
	}
	workingDirectory, inside := inRepository(this.root, this.workingDirectory)
	if !inside {
		t.Fatalf("the test runs outside the repository, in %s", this.workingDirectory)
	}
	var keys []string
	for range 2 {
		checkout := t.TempDir()
		for _, file := range append(slices.Clone(lowering.files), corpus[3]...) {
			copyInto(t, this.root, checkout, file)
		}
		under := environment{root: checkout, getenv: this.getenv, workingDirectory: filepath.Join(checkout, filepath.FromSlash(workingDirectory))}
		paths := make([]string, len(corpus[3]))
		for index, path := range corpus[3] {
			paths[index] = filepath.Join(checkout, filepath.FromSlash(path))
		}
		keys = append(keys, keyAt(t, checkout, inputsOf(t, sourcesOf(t, paths), Options{}, under)))
	}
	if keys[0] != keys[1] {
		t.Fatalf("two checkouts of one tree keyed apart: %s and %s", keys[0][:12], keys[1][:12])
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
	inputs := inputsOf(t, sourcesOf(t, fromHere(corpus[0])), Options{}, here(t))
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

// A stored IR read back emits the C a fresh lowering from disk emits, byte for byte, the build's own decode and a hit's
// alike, and the second read is a hit: the program built once.
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
			root, err := repositoryRoot()
			if err != nil {
				t.Fatal(err)
			}
			if keyAt(t, root, missedInputs) != keyAt(t, root, hitInputs) {
				t.Fatal("two reads of one program keyed it apart")
			}
			want := freshC(t, fromHere(paths)[0], options)
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

// replaced writes data as the stored gob, with sum as its sum file, then reads the program again: a miss, so it built
// again into another directory, and the IR it serves emits the C a fresh lowering does; a read after that finds the
// rebuilt product.
func replaced(t *testing.T, first string, read sources, builds *atomic.Int32, data, sum []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(first, gobFile), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(first, sumFile), sum, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readStored(first); err == nil {
		t.Fatal("the stored IR reads back")
	} else {
		t.Logf("the stored IR doesn't read back: %v", err)
	}
	program, _, second := stored(t, read, Options{}, counted(builds))
	if second == first || builds.Load() != 2 {
		t.Fatalf("served from %s, built %d times: the stored IR was served", second, builds.Load())
	}
	sameC(t, "the rebuilt IR", freshC(t, fromHere(tornProgram)[0], Options{}), native.C(program))
	if _, _, third := stored(t, read, Options{}, counted(builds)); third != second || builds.Load() != 2 {
		t.Fatalf("a read after the rebuild built again (%d builds), from %s", builds.Load(), third)
	}
}

// storedFiles are a product's gob and trace, as its sum names them.
func storedFiles(t *testing.T, directory string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	for _, file := range []string{gobFile, resolutionFile} {
		data, err := os.ReadFile(filepath.Join(directory, file))
		if err != nil {
			t.Fatal(err)
		}
		files[file] = data
	}
	return files
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
	files := storedFiles(t, first)
	at := bytes.Index(files[gobFile], []byte("outside"))
	if at < 0 {
		t.Fatalf("%s holds no 'outside' to tear", tornProgram[0])
	}
	torn := slices.Clone(files[gobFile])
	copy(torn[at:], "OUTSIDE")
	replaced(t, first, read, &builds, torn, sums(files))
}

// A stored IR gob can't decode is a miss and rebuilt, its sum right for the bytes it holds.
//
// Not parallel: privateCache sets the build cache's variables for the process.
func TestAnUndecodableStoredIRIsAMissAndRebuilt(t *testing.T) {
	privateCache(t)
	read := sourcesOf(t, fromHere(tornProgram))
	var builds atomic.Int32
	_, _, first := stored(t, read, Options{}, counted(&builds))
	files := storedFiles(t, first)
	files[gobFile] = []byte("not a gob")
	replaced(t, first, read, &builds, files[gobFile], sums(files))
}

// writeProgram writes files, by path from directory, and returns the entry's path, the first's.
func writeProgram(t *testing.T, directory string, files ...[2]string) string {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(directory, filepath.FromSlash(file[0]))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(file[1]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, filepath.FromSlash(files[0][0]))
}

// A package.json beside a program, which resolution reads for the module kind of every file in its scope, refuses it:
// when it's there as the program's lowering is built, and when it's added after, as the stored lowering is read. The
// program's trace names it. Mutant: the absence check dropped.
func TestASiblingPackageJSONRefusesTheProgram(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := writeProgram(t, directory, [2]string{"main.ts", "console.log('beside');\n"})
	read := sourcesOf(t, []string{entry})
	product := t.TempDir()
	if err := build(product, read, Options{}); err != nil {
		t.Fatal(err)
	}
	_, trace, err := readStored(product)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(trace, "at package.json") || !slices.Contains(trace, "above package.json") {
		t.Fatalf("the trace %v doesn't name the package.json beside the program and above it", trace)
	}
	writeProgram(t, directory, [2]string{"package.json", `{"type": "commonjs"}` + "\n"})
	beside := filepath.Join(directory, "package.json")
	if err := absent(read.base, trace); err == nil || !strings.Contains(err.Error(), beside) {
		t.Errorf("the stored lowering read with %s added: %v, want it refused", beside, err)
	}
	if _, _, err := loaded(read, Options{}); err == nil || !strings.Contains(err.Error(), beside) {
		t.Errorf("the program built beside %s: %v, want it refused", beside, err)
	}
}

// A node_modules package a program's bare import resolves to refuses it, though a fresh load from disk checks it: the
// sealed load finds no package, and its trace names where it looked. Mutant: the absence check dropped.
func TestANodeModulesPackageAProgramImportsRefusesIt(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := writeProgram(t, directory,
		[2]string{"main.ts", "import { value } from 'helper';\nconsole.log(value);\n"},
		[2]string{"node_modules/helper/package.json", `{"name": "helper", "types": "./index.ts"}` + "\n"},
		[2]string{"node_modules/helper/index.ts", "export const value = 'helped';\n"})
	if _, err := load.Load([]string{entry}); err != nil {
		t.Logf("a fresh load from disk doesn't check either: %v", err)
	}
	_, _, err := loaded(sourcesOf(t, []string{entry}), Options{})
	if err == nil || !strings.Contains(err.Error(), filepath.Join(directory, "node_modules")) || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("got %v, want the program refused for the node_modules it resolves through", err)
	}
	t.Log(err)
}

// A node_modules directory beside a program whose imports are all relative changes nothing it lowers to: resolution
// never looks there (its trace names no node_modules), and the sealed lowering emits the C a fresh lowering from disk
// emits with the directory there.
func TestANodeModulesDirectoryBesideAProgramThatNeverLooksChangesNothing(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := writeProgram(t, directory,
		[2]string{"main.ts", "import { value } from './values.ts';\nconsole.log(value);\n"},
		[2]string{"values.ts", "export const value = 'relative';\n"},
		[2]string{"node_modules/values/package.json", `{"name": "values", "types": "./index.ts"}` + "\n"},
		[2]string{"node_modules/values/index.ts", "export const value = 'bare';\n"})
	program, trace, err := loaded(sourcesOf(t, []string{entry, filepath.Join(directory, "values.ts")}), Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range trace {
		if strings.Contains(line, "node_modules") {
			t.Errorf("the trace names %s", line)
		}
	}
	lowering, err := lower.LowerWithOptions(context.Background(), program, lower.Options{})
	if err != nil {
		t.Fatal(err)
	}
	sameC(t, "the sealed lowering", freshC(t, entry, Options{}), native.C(lowering))
}

// The checker asks after no path under the working directory: given absolute sources, load names nothing from it, so
// the key holds it from the repository's root and never as the machine spells it.
func TestTheTraceNamesNothingOfTheWorkingDirectory(t *testing.T) {
	t.Parallel()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, paths := range corpus {
		read := sourcesOf(t, fromHere(paths))
		held := map[string]string{}
		for file, data := range read.files {
			held[filepath.ToSlash(filepath.Join(read.base, file))] = string(data)
		}
		disk := sealedOver(held)
		if _, err := load.LoadFrom([]string{filepath.Join(read.base, read.entry)}, buildcache.Relative, disk); err != nil {
			t.Fatal(err)
		}
		for _, name := range disk.trace() {
			if strings.HasPrefix(name, filepath.ToSlash(workingDirectory)+"/") {
				t.Errorf("%s: the checker asked after %s, under the working directory", paths[0], name)
			}
		}
	}
}

// An environment read is a variable some file of the closure reads, and where.
type environmentRead struct{ variable, where string }

// environmentReads is every os.Getenv, os.LookupEnv and os.Environ (and syscall's) in a package's files: by the
// variable it names, a string literal or a constant of the package holding one, or "" for a read by any other name, or
// os.Environ, which reads them all.
func environmentReads(t *testing.T, files []string) []environmentRead {
	t.Helper()
	positions := token.NewFileSet()
	var parsed []*ast.File
	constants := map[string]string{}
	for _, file := range files {
		syntax, err := parser.ParseFile(positions, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed = append(parsed, syntax)
		for _, declaration := range syntax.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.CONST {
				continue
			}
			for _, spec := range general.Specs {
				value := spec.(*ast.ValueSpec)
				for index, name := range value.Names {
					if index < len(value.Values) {
						if literal, ok := value.Values[index].(*ast.BasicLit); ok && literal.Kind == token.STRING {
							constants[name.Name], _ = strconv.Unquote(literal.Value)
						}
					}
				}
			}
		}
	}
	var reads []environmentRead
	for _, syntax := range parsed {
		packages := map[string]bool{}
		for _, spec := range syntax.Imports {
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
		called := map[*ast.SelectorExpr]bool{}
		ast.Inspect(syntax, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selector, ok := reading(call.Fun); ok {
					called[selector] = true
					variable := ""
					if len(call.Args) > 0 && selector.Sel.Name != "Environ" {
						switch argument := call.Args[0].(type) {
						case *ast.BasicLit:
							if argument.Kind == token.STRING {
								variable, _ = strconv.Unquote(argument.Value)
							}
						case *ast.Ident:
							variable = constants[argument.Name]
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
	}
	return reads
}

// Every environment variable the lowering's closure reads, in each package inside the repository go list -deps
// ./internal/load ./internal/lower names (this module's, the checker's and its shims'), as Linux, macOS and Android
// build it, is keyed by its value or said to shape nothing the IR holds, so a new read can't leave the key silently
// short; a read by a name no source spells fails, and so does a listed variable nothing reads. Today's reads:
// ADAMIC_PROGRAM_REGION and ADAMIC_OWNERSHIP_QUERY (lower.Lower) and TS_GO_DEBUG_STACK_LIMIT (tsc's core), keyed;
// ADAMIC_BUILD_LOG (buildlog.Phase) and TERMUX_EXEC__PROC_SELF_EXE (tsc's osutil, on Android), not. Mutants: an
// os.Getenv planted in internal/lower with no key line; TS_GO_DEBUG_STACK_LIMIT's key line dropped.
func TestEveryEnvironmentReadOfTheLoweringIsKeyed(t *testing.T) {
	t.Parallel()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	byPackage := map[string]map[string]bool{}
	for _, system := range []string{"linux", "darwin", "android"} {
		packages, err := listed(root, []string{"GOOS=" + system, "CGO_ENABLED=1"}, true, loweringEntry...)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range packages {
			if _, inside := inRepository(root, item.Dir); !inside {
				continue
			}
			if byPackage[item.Dir] == nil {
				byPackage[item.Dir] = map[string]bool{}
			}
			for _, file := range append(slices.Clone(item.GoFiles), item.CgoFiles...) {
				byPackage[item.Dir][filepath.Join(item.Dir, file)] = true
			}
		}
	}
	read := map[string][]string{}
	count := 0
	for _, files := range byPackage {
		names := make([]string, 0, len(files))
		for file := range files {
			names = append(names, file)
		}
		count += len(names)
		for _, found := range environmentReads(t, names) {
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
	setting := here(t)
	setting.getenv = func(variable string) string { return "set-" + variable }
	keyed := inputsOf(t, sourcesOf(t, fromHere(corpus[0])), Options{}, setting)
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
		sort.Strings(where)
		variables = append(variables, variable+" ("+strings.Join(slices.Compact(where), ", ")+")")
	}
	sort.Strings(variables)
	t.Logf("%d files in %d packages read %s", count, len(byPackage), strings.Join(variables, ", "))
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
		program, _, err := loaded(read, Options{})
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
			case strings.HasPrefix(path, filepath.ToSlash(read.base)+"/"):
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

// A program that imports a file on disk that isn't among its sources is refused, never stored under a key that
// doesn't name the file.
func TestAProgramReadingAFileBesideItsSourcesIsRefused(t *testing.T) {
	t.Parallel()
	elsewhere := writeProgram(t, t.TempDir(), [2]string{"values.ts", "export const value = 'beside';\n"})
	entry := writeProgram(t, t.TempDir(), [2]string{"main.ts", fmt.Sprintf("import { value } from %q;\nconsole.log(value);\n", elsewhere)})
	// Resolution asks first whether the file's directory is there, and the trace names that.
	if _, _, err := loaded(sourcesOf(t, []string{entry}), Options{}); err == nil || !strings.Contains(err.Error(), filepath.Dir(elsewhere)) || !strings.Contains(err.Error(), "refused") {
		t.Fatalf("got %v, want the read of %s refused", err, elsewhere)
	} else {
		t.Log(err)
	}
}
