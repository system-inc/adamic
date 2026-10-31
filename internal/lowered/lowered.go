// Package lowered keeps a program's lowered IR as a build product, so a compiler change that touches only emit, native
// or the runtime reuses the load and the lowering and only emits C again (the IR cache's stage 1; stage 2, the C,
// is a product keyed on this one's key and emit's closure).
//
// It lives apart from both halves it joins. internal/lower can't hold it: the key would then name buildcache's
// closure, and lower is in its own key. buildcache can't hold it: every product's package would link the compiler.
// The products that keep a program.gob today (estree's, lint's, flow's) each wrote their own in a test file, which no
// other package can call; this one is a package, so each of them, and stage 2, can.
//
// Stage 1's key is what the lowering is a function of, and nothing else:
//
//   - the Go closure of the lowering, computed by go list -deps ./internal/load ./internal/lower (closure.go), with
//     this package's own files, which build the product. internal/native and internal/javascript are not in it, so a
//     change confined to emit keeps the key;
//   - the program's sources, by content, and where they sit in the repository;
//   - every field of lower's options and of load's (TSGo), by value; every environment variable the closure reads that
//     can shape the IR (keyedEnvironment, held to the closure's source by TestEveryEnvironmentReadOfTheLoweringIsKeyed);
//     the working directory, from the repository's root; and the case sensitivity the checker compares names by;
//   - the Go release and the system it targets, and GOFLAGS, whose tags choose the closure's files.
//
// The checker reads nothing from disk: it loads the sources from memory, where they sit (sealed.go), and every other
// path it asks after answers absent and is traced. That trace, what the program's resolution looked for outside its
// sources (each directory's package.json up to the root, the names an import might mean), is stored with the IR, and
// every path in it must be absent on disk, when the product is built and each time it's read (resolution.go): a
// program beside which one of them is there is refused, never lowered without it.
//
// The product holds the IR as gob, encoded once, and the trace, with the sha256 of each beside them. Every read checks
// the sums and decodes, and a sum that doesn't match or a decode that fails is a miss, rebuilt, never an IR served.
// Each read decodes afresh, since emit writes into the IR it's given.
package lowered

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Options are what a program is loaded and lowered with besides its sources: lower's options, and whether the program
// links the typescript-go checker library (load.Program.EnableTSGo).
type Options struct {
	Lower lower.Options
	TSGo  bool
}

// optionLines are every field of options by its name and value. A field added to Options or lower.Options without a
// line here fails TestEveryOptionIsKeyedByValue.
func optionLines(options Options) []string {
	return []string{
		"Lower.Debug=" + strconv.FormatBool(options.Lower.Debug),
		"Lower.ProgramRegion=" + strconv.FormatBool(options.Lower.ProgramRegion),
		"Lower.OwnershipQuery=" + strconv.FormatBool(options.Lower.OwnershipQuery),
		"TSGo=" + strconv.FormatBool(options.TSGo),
	}
}

// productName is the product's name. A change to what the product holds, or to how it's read, is a new name.
const productName = "lowered-ir-v2"

// The product's files: the IR as gob, the trace of what its resolution looked for, and the hex sha256 of each.
const (
	gobFile        = "program.gob"
	resolutionFile = "resolution"
	sumFile        = "sha256"
)

// keyedEnvironment is every environment variable the lowering's closure reads that can change the IR it makes, each
// keyed by its value: lower.Lower's two (lower.go), which lower.Lower and cmd/adamic take Options from, and the
// checker's stack limit (tsc's core.ApplyDebugStackLimit), which can change whether the checker stops.
var keyedEnvironment = []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_OWNERSHIP_QUERY"}

// unkeyedEnvironment is every variable the closure reads that can't change the IR, with why. A read the guard finds in
// neither list fails it.
var unkeyedEnvironment = map[string]string{
	"ADAMIC_BUILD_LOG":           "buildlog.Phase appends a phase's timing to the log the variable names; nothing it writes is an input to the program (buildlog.go)",
	"TERMUX_EXEC__PROC_SELF_EXE": "read on Android alone (tsc's osutil/os_android.go), naming the executable whose path osvfs probes for the file system's case sensitivity; that sensitivity is keyed itself, and the libraries a noembed build reads beside it are embedded here",
}

// Lowered is the program lowered from paths, the entry first and then each file its imports read, all inside the
// repository, with options: a fresh decode of the stored IR, built on a miss, and stage 1's key, the Inputs a stage 2
// product extends with emit's closure. A program whose resolution would find a file on disk beside its sources fails
// t: it's refused for the cache.
func Lowered(t testing.TB, paths []string, options Options) (*ir.Program, buildcache.Inputs) {
	t.Helper()
	program, inputs, _ := stored(t, sourcesOf(t, paths), options, build)
	return program, inputs
}

// sources are a program's files as they were read for the key: base, the absolute directory holding them all, and each
// file by its path from base; entry is the first's.
type sources struct {
	base  string
	entry string
	files map[string][]byte
}

// sourcesOf reads paths, the entry first, each once, so the key and the build see the same bytes.
func sourcesOf(t testing.TB, paths []string) sources {
	t.Helper()
	if len(paths) == 0 {
		t.Fatal("lowered: no sources")
	}
	absolute := make([]string, len(paths))
	for index, path := range paths {
		resolved, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		absolute[index] = resolved
	}
	base := filepath.Dir(absolute[0])
	for _, path := range absolute[1:] {
		for base != filepath.Dir(base) && !strings.HasPrefix(path, base+string(filepath.Separator)) {
			base = filepath.Dir(base)
		}
	}
	read := sources{base: base, files: map[string][]byte{}}
	for index, path := range absolute {
		relative, err := filepath.Rel(base, path)
		if err != nil {
			t.Fatal(err)
		}
		relative = filepath.ToSlash(relative)
		if index == 0 {
			read.entry = relative
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		read.files[relative] = data
	}
	return read
}

// environment is what the key reads from the process: the repository's root, a variable's value, and the working
// directory.
type environment struct {
	root             string
	getenv           func(string) string
	workingDirectory string
}

// processEnvironment is this process's.
func processEnvironment(t testing.TB) environment {
	t.Helper()
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return environment{root: root, getenv: os.Getenv, workingDirectory: workingDirectory}
}

// inRepository is path from the repository's root, when it's at or under it.
func inRepository(root, path string) (string, bool) {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(relative), true
}

// inputsOf is stage 1's key for the program from sources with options, under environment. Every place in it is named
// from the repository's root, so two checkouts of one tree key alike.
func inputsOf(t testing.TB, read sources, options Options, under environment) buildcache.Inputs {
	t.Helper()
	lowering, err := loweringClosure()
	if err != nil {
		t.Fatal(err)
	}
	location, inside := inRepository(under.root, read.base)
	if !inside {
		t.Fatalf("lowered: the sources sit in %s, outside the repository: a product names its program from the repository's root", read.base)
	}
	inputs := buildcache.Inputs{
		Name:      productName,
		Files:     slices.Clone(lowering.files),
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	inputs.Flags = append(inputs.Flags, optionLines(options)...)
	inputs.Flags = append(inputs.Flags, "location="+location, "entry="+read.entry)
	names := make([]string, 0, len(read.files))
	for file := range read.files {
		names = append(names, file)
	}
	sort.Strings(names)
	for _, file := range names {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("source %s=%x", file, sha256.Sum256(read.files[file])))
	}
	for _, variable := range keyedEnvironment {
		inputs.Flags = append(inputs.Flags, variable+"="+under.getenv(variable))
	}
	// Load reads no path from it (TestTheTraceNamesNothingOfTheWorkingDirectory), but it's the directory names resolve
	// against, so a directory of the tree is named from the root, and one outside it as it is.
	workingDirectory := under.workingDirectory
	if relative, inside := inRepository(under.root, workingDirectory); inside {
		workingDirectory = "<repository>/" + relative
	}
	// GOFLAGS' tags choose the closure's files (a noembed build reads the checker's libraries from disk), and
	// buildcache drops its -p share.
	inputs.Flags = append(inputs.Flags, "GOFLAGS="+under.getenv("GOFLAGS"), "working-directory="+workingDirectory,
		"case-sensitive="+strconv.FormatBool(osvfs.FS().CaseSensitivity() == tspath.CaseSensitive))
	inputs.Flags = append(inputs.Flags, lowering.external...)
	return inputs
}

// stored is the program from sources, lowered with options: the stored IR, decoded, the key it's stored under and the
// directory it was read from. A stored IR that doesn't read back is a miss: it's built again, as a product keyed beside
// the first by what it found there, so the next read of the same torn product finds the rebuilt one. A program whose
// traced resolution finds a file on disk now is refused, built or found. builder is build, or a test's count of it.
func stored(t testing.TB, read sources, options Options, builder func(directory string, read sources, options Options) error) (*ir.Program, buildcache.Inputs, string) {
	t.Helper()
	inputs := inputsOf(t, read, options, processEnvironment(t))
	directory := buildcache.Product(t, inputs, func(directory string) error { return builder(directory, read, options) })
	program, trace, err := readStored(directory)
	if err != nil {
		t.Logf("lowered: the IR stored in %s doesn't read back, a miss: %v", directory, err)
		rebuilt := inputs
		rebuilt.Flags = append(slices.Clone(inputs.Flags), "replaces="+storedDigest(directory))
		directory = buildcache.Product(t, rebuilt, func(directory string) error { return builder(directory, read, options) })
		program, trace, err = readStored(directory)
		if err != nil {
			t.Fatalf("lowered: the IR rebuilt in %s doesn't read back either: %v", directory, err)
		}
	}
	if err := absent(read.base, trace); err != nil {
		t.Fatal(err)
	}
	return program, inputs, directory
}

// storedDigest names what a product directory holds, whatever it is: the sha256 of its files together.
func storedDigest(directory string) string {
	hash := sha256.New()
	for _, file := range []string{gobFile, resolutionFile, sumFile} {
		data, _ := os.ReadFile(filepath.Join(directory, file))
		fmt.Fprintf(hash, "%s %d\n", file, len(data))
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// sums is the sum file for the product's files, as sha256sum writes one.
func sums(files map[string][]byte) []byte {
	var written bytes.Buffer
	for _, file := range []string{gobFile, resolutionFile} {
		fmt.Fprintf(&written, "%x  %s\n", sha256.Sum256(files[file]), file)
	}
	return written.Bytes()
}

// readStored decodes the IR stored in directory and reads its resolution's trace, once the bytes of each are the ones
// the sum file names.
func readStored(directory string) (*ir.Program, []string, error) {
	files := map[string][]byte{}
	for _, file := range []string{gobFile, resolutionFile} {
		data, err := os.ReadFile(filepath.Join(directory, file))
		if err != nil {
			return nil, nil, err
		}
		files[file] = data
	}
	sum, err := os.ReadFile(filepath.Join(directory, sumFile))
	if err != nil {
		return nil, nil, err
	}
	if want := sums(files); !bytes.Equal(sum, want) {
		return nil, nil, fmt.Errorf("the product's sha256 are\n%s and its sum file says\n%s", want, sum)
	}
	ir.RegisterGob()
	var program ir.Program
	if err := gob.NewDecoder(bytes.NewReader(files[gobFile])).Decode(&program); err != nil {
		return nil, nil, fmt.Errorf("decode %s: %w", gobFile, err)
	}
	return &program, strings.Split(strings.TrimSuffix(string(files[resolutionFile]), "\n"), "\n"), nil
}

// build lowers the program from sources into directory: loaded sealed, lowered with options, encoded once, and stored
// with its resolution's trace and the sums of both.
func build(directory string, read sources, options Options) error {
	program, trace, err := loaded(read, options)
	if err != nil {
		return err
	}
	lowering, err := lower.LowerWithOptions(context.Background(), program, options.Lower)
	if err != nil {
		return err
	}
	ir.RegisterGob()
	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(lowering); err != nil {
		return fmt.Errorf("encode the IR: %w", err)
	}
	files := map[string][]byte{gobFile: encoded.Bytes(), resolutionFile: []byte(strings.Join(trace, "\n") + "\n")}
	for file, data := range files {
		if err := os.WriteFile(filepath.Join(directory, file), data, 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(directory, sumFile), sums(files), 0o644)
}

// loaded is the program from sources, checked on a sealed disk that holds them where they sit and nothing else, and the
// trace of what its resolution looked for outside them. It's refused when one of those paths is on disk, before any
// error the check gave: the program a fresh load reads would be another one.
func loaded(read sources, options Options) (*load.Program, []string, error) {
	held := make(map[string]string, len(read.files))
	for file, data := range read.files {
		held[filepath.ToSlash(filepath.Join(read.base, filepath.FromSlash(file)))] = string(data)
	}
	disk := sealedOver(held)
	// Named placelessly: the IR's positions name the tree as <repository> (#tqrqx60).
	program, checked := load.LoadFrom([]string{filepath.Join(read.base, filepath.FromSlash(read.entry))}, buildcache.Relative, disk)
	trace, err := resolutionOf(read.base, disk.trace())
	if err != nil {
		return nil, nil, err
	}
	if err := absent(read.base, trace); err != nil {
		return nil, nil, err
	}
	if checked != nil {
		return nil, nil, checked
	}
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if !inBinary(file) && !disk.holds(file.FileName().AsString()) {
			return nil, nil, fmt.Errorf("lowered: the program reads %s, which isn't one of its sources", file.FileName().AsString())
		}
	}
	if options.TSGo {
		program.EnableTSGo()
	}
	return program, trace, nil
}

// bundledScheme is the prefix of every path the checker serves from the binary: its libraries (bundled:///libs) and
// Node's declarations (bundled:///node).
const bundledScheme = "bundled:///"

// inBinary is whether the checker read file from the binary rather than from disk: the prelude, or a bundled file.
func inBinary(file *ast.SourceFile) bool {
	return load.IsPrelude(file) || strings.HasPrefix(file.FileName().AsString(), bundledScheme)
}
