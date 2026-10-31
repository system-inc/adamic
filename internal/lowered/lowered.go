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
//   - the program's sources, by content, laid out as they sit beside the entry;
//   - lower's options and load's (TSGo), every environment variable the closure reads that can shape the IR
//     (keyedEnvironment, held to the closure's source by TestEveryEnvironmentReadOfTheLoweringIsKeyed), and the working
//     directory load resolves names against (load.go's os.Getwd);
//   - the Go release and the system it targets, and GOFLAGS, whose tags choose the closure's files.
//
// The product holds the IR as gob, encoded once, and its sha256 beside it. Every read checks the sum and decodes, and a
// sum that doesn't match or a decode that fails is a miss, rebuilt, never an IR served. Each read decodes afresh, since
// emit writes into the IR it's given.
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
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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

// productName is the product's name. A change to what the product holds, or to how it's read, is a new name.
const productName = "lowered-ir-v1"

// The product's two files: the IR as gob, and the hex sha256 of exactly those bytes.
const (
	gobFile = "program.gob"
	sumFile = "program.gob.sha256"
)

// keyedEnvironment is every environment variable the lowering's closure reads that can change the IR it makes, each
// keyed by its value: lower.Lower's two (lower.go). Lowered lowers with the Options it's given, which lower.Lower and
// cmd/adamic take from these, so the key holds both.
var keyedEnvironment = []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_OWNERSHIP_QUERY"}

// unkeyedEnvironment is every variable the closure reads that can't change the IR, with why. A read the guard finds in
// neither list fails it.
var unkeyedEnvironment = map[string]string{
	"ADAMIC_BUILD_LOG": "buildlog.Phase appends a phase's timing to the log the variable names; nothing it writes is an input to the program (buildlog.go)",
}

// Lowered is the program lowered from paths, the entry first and then each file its imports read, with options: a
// fresh decode of the stored IR, built on a miss, and stage 1's key, the Inputs a stage 2 product extends with emit's
// closure.
func Lowered(t testing.TB, paths []string, options Options) (*ir.Program, buildcache.Inputs) {
	t.Helper()
	program, inputs, _ := stored(t, sourcesOf(t, paths), options, build)
	return program, inputs
}

// sources are a program's files as they were read for the key, by their path from the directory holding them all;
// entry is the first's.
type sources struct {
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
	// The directory holding them all, which the build lays them out under as they sit there.
	base := filepath.Dir(absolute[0])
	for _, path := range absolute[1:] {
		for base != filepath.Dir(base) && !strings.HasPrefix(path, base+string(filepath.Separator)) {
			base = filepath.Dir(base)
		}
	}
	read := sources{files: map[string][]byte{}}
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

// environment is what the key reads from the process: a variable's value, and the working directory.
type environment struct {
	getenv           func(string) string
	workingDirectory string
}

// processEnvironment is this process's.
func processEnvironment(t testing.TB) environment {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return environment{getenv: os.Getenv, workingDirectory: workingDirectory}
}

// inputsOf is stage 1's key for the program from sources with options, under environment.
func inputsOf(t testing.TB, read sources, options Options, under environment) buildcache.Inputs {
	t.Helper()
	lowering, err := loweringClosure()
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      productName,
		Files:     slices.Clone(lowering.files),
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	// Every field of both option sets by name, so a field added to either is keyed without a line here.
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("options=%+v", options), "entry="+read.entry)
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
	// GOFLAGS' tags choose the closure's files (a noembed build reads the checker's libraries from disk), and
	// buildcache drops its -p share.
	inputs.Flags = append(inputs.Flags, "GOFLAGS="+under.getenv("GOFLAGS"), "working-directory="+under.workingDirectory)
	inputs.Flags = append(inputs.Flags, lowering.external...)
	return inputs
}

// stored is the program from sources, lowered with options: the stored IR, decoded, the key it's stored under and the
// directory it was read from. A stored IR that doesn't read back is a miss: it's built again, as a product
// keyed beside the first by what it found there, so the next read of the same torn product finds the rebuilt one. builder
// is build, or a test's count of it.
func stored(t testing.TB, read sources, options Options, builder func(directory string, read sources, options Options) error) (*ir.Program, buildcache.Inputs, string) {
	t.Helper()
	inputs := inputsOf(t, read, options, processEnvironment(t))
	directory := buildcache.Product(t, inputs, func(directory string) error { return builder(directory, read, options) })
	program, err := readStored(directory)
	if err == nil {
		return program, inputs, directory
	}
	t.Logf("lowered: the IR stored in %s doesn't read back, a miss: %v", directory, err)
	rebuilt := inputs
	rebuilt.Flags = append(slices.Clone(inputs.Flags), "replaces="+storedDigest(directory))
	directory = buildcache.Product(t, rebuilt, func(directory string) error { return builder(directory, read, options) })
	program, err = readStored(directory)
	if err != nil {
		t.Fatalf("lowered: the IR rebuilt in %s doesn't read back either: %v", directory, err)
	}
	return program, inputs, directory
}

// storedDigest names what a product directory holds, whatever it is: the sha256 of its gob and its sum together.
func storedDigest(directory string) string {
	hash := sha256.New()
	for _, file := range []string{gobFile, sumFile} {
		data, _ := os.ReadFile(filepath.Join(directory, file))
		fmt.Fprintf(hash, "%s %d\n", file, len(data))
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// readStored decodes the IR stored in directory, once its bytes are the ones its sum names.
func readStored(directory string) (*ir.Program, error) {
	data, err := os.ReadFile(filepath.Join(directory, gobFile))
	if err != nil {
		return nil, err
	}
	sum, err := os.ReadFile(filepath.Join(directory, sumFile))
	if err != nil {
		return nil, err
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(data)), strings.TrimSpace(string(sum)); got != want {
		return nil, fmt.Errorf("%s's sha256 is %s, and its sum says %s", gobFile, got, want)
	}
	ir.RegisterGob()
	var program ir.Program
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&program); err != nil {
		return nil, fmt.Errorf("decode %s: %w", gobFile, err)
	}
	return &program, nil
}

// build lowers the program from sources into directory: the sources laid out under source/ as they sat, loaded from
// there, lowered with options, encoded once, and the encoding's sum written beside it.
func build(directory string, read sources, options Options) error {
	program, err := loaded(directory, read, options)
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
	if err := os.WriteFile(filepath.Join(directory, gobFile), encoded.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, sumFile), fmt.Appendf(nil, "%x\n", sha256.Sum256(encoded.Bytes())), 0o644); err != nil {
		return err
	}
	// The sources' imports may name this checkout; the product names it <repository> (#tqrqx60).
	return buildcache.RelativeFiles(filepath.Join(directory, "source"))
}

// loaded is the program from sources, laid out under directory's source/ and loaded from there, refused when the
// checker read a file from disk that isn't one of them: the key names the sources alone, so a program reading
// another would be served when that file changed. The checker's libraries, Node's declarations and the prelude are
// in the binary (TestTheLibrariesAndThePreludeAreInTheBinary), at paths no disk has.
func loaded(directory string, read sources, options Options) (*load.Program, error) {
	source := filepath.Join(directory, "source")
	for file, data := range read.files {
		path := filepath.Join(source, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, err
		}
	}
	// Named placelessly: the IR's positions would otherwise carry this build directory, random per build (#tqrqx60).
	program, err := load.LoadDisplayed([]string{filepath.Join(source, filepath.FromSlash(read.entry))}, buildcache.Relative)
	if err != nil {
		return nil, err
	}
	// The checker may name the directory as its links resolve (macOS's /var is /private/var).
	places := []string{source + "/"}
	if resolved, err := filepath.EvalSymlinks(source); err == nil {
		places = append(places, resolved+"/")
	}
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		path := file.FileName().AsString()
		if !inBinary(file) && !slices.ContainsFunc(places, func(place string) bool { return strings.HasPrefix(path, place) }) {
			return nil, fmt.Errorf("lowered: the program reads %s, which isn't one of its sources; name it among Lowered's paths", path)
		}
	}
	if options.TSGo {
		program.EnableTSGo()
	}
	return program, nil
}

// bundledScheme is the prefix of every path the checker serves from the binary: its libraries (bundled:///libs) and
// Node's declarations (bundled:///node).
const bundledScheme = "bundled:///"

// inBinary is whether the checker read file from the binary rather than from disk: the prelude, or a bundled file.
func inBinary(file *ast.SourceFile) bool {
	return load.IsPrelude(file) || strings.HasPrefix(file.FileName().AsString(), bundledScheme)
}
