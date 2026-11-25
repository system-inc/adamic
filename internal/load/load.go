// Package load turns Adamic source files into a checked program: every file parsed, bound and
// type-checked by typescript-go in this process, under the one set of options Adamic 0.1 allows.
//
// A program either loads clean or Load returns an error naming every diagnostic. There is no
// half-loaded state, because a compiler that lowers a program the checker rejected is lowering a
// guess.
package load

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/buildlog"
)

var errReadOnly = errors.New("load: the source file system is read-only")

// Program is a loaded Adamic program that the checker accepted.
type Program struct {
	// tsgo opts this compilation into the external native checker library.
	tsgo bool

	compiler *compiler.Program
	fs       *sourceFS

	// files is the program's own source, in the order Load was given it: no prelude, no lib.
	files []*ast.SourceFile
}

// CheckError is a program the checker rejected, with every diagnostic it gave.
type CheckError struct {
	Diagnostics []string
}

func (e *CheckError) Error() string {
	return strings.Join(e.Diagnostics, "\n")
}

// compilerOptions is the one configuration every Adamic 0.1 program is checked under, set here
// rather than read from a tsconfig.json, which could leave any of them out (docs/0.1.md).
func compilerOptions() *core.CompilerOptions {
	return &core.CompilerOptions{
		Strict:                     core.TSTrue,
		NoUncheckedIndexedAccess:   core.TSTrue,
		ExactOptionalPropertyTypes: core.TSTrue,
		ErasableSyntaxOnly:         core.TSFalse,
		VerbatimModuleSyntax:       core.TSTrue,
		AllowImportingTsExtensions: core.TSTrue,
		NoEmit:                     core.TSTrue,
		Module:                     core.ModuleKindESNext,
		// Every Adamic file is a module, imports or not, as the oracle runs it. Without this, a file with
		// no import is a script to the checker, and its top-level names collide with the prelude's.
		ModuleDetection:  core.ModuleDetectionKindForce,
		ModuleResolution: core.ModuleResolutionKindBundler,
		Target:           core.ScriptTargetES2024,
		Lib:              []string{"lib.es2024.d.ts"},
		Types:            []string{},
	}
}

// Load checks the program made of the given files and whatever they import.
//
// Each path is an Adamic source: a .ts file or a .a file.
func Load(paths []string) (*Program, error) {
	return load(paths, nil, nil)
}

// LoadDisplayed checks paths as Load does, naming each file by display(its name) wherever the program names it: in
// positions, diagnostics and declarations. A product that keeps a lowered program passes buildcache.Relative, so the
// program names its own build directory as <build cache>/<name key> and the tree as <repository>, never the machine or
// the directory that built it (#tqrqx60: landable-13b's estree products named a random build directory, so no two
// builds of one program were the same bytes). display must keep two names apart when it takes them in apart.
func LoadDisplayed(paths []string, display func(string) string) (*Program, error) {
	return load(paths, nil, display)
}

// LoadOverlay checks paths while replacing the named files with source held by the caller. The
// overlay is read-only and never written to disk; names are absolute or relative to the current
// directory. This is for tools which test semantics-preserving source adaptations.
func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) {
	return load(paths, overlay, nil)
}

func load(paths []string, overlay map[string]string, display func(string) string) (*Program, error) {
	// The load phase's line on $ADAMIC_BUILD_LOG, once a program, whichever of Load's forms called.
	defer buildlog.Phase("load", time.Now())
	built, err := firstProgram(paths, overlay, display)
	if err != nil {
		return nil, err
	}

	if UsesNodeModules(built.program) {
		index, err := nodeTypesIndex()
		if err != nil {
			return nil, err
		}
		// Node's console and NodeJS.Process's recordreplay in the standalone console's place (source_fs.go).
		built.roots[len(paths)] = preludeNodePath
		built.roots = append(built.roots, index)
		built.fileSystem = cachedvfs.From(&regexpLibraryFS{FS: &nodeLibraryFS{FS: bundled.WrapFS(built.fs)}})
		options := compilerOptions()
		options.TypeRoots = []tspath.RootedDirectoryPath{tspath.RootedDirectoryPathFromNormalized(nodeTypesRoot)}
		config := tsoptions.NewParsedCommandLine(options, built.roots, nil, built.currentDirectory, built.fileSystem.CaseSensitivity())
		host := compiler.NewCachedFSCompilerHost(built.fileSystem, bundled.LibPath(), nil, nil, nil)
		built.program = compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
		if built.program == nil {
			return nil, errors.New("load: the compiler built no Node program")
		}
	}

	loaded := &Program{compiler: built.program, fs: built.fs}
	if diagnostics := loaded.diagnostics(context.Background()); len(diagnostics) > 0 {
		return nil, &CheckError{Diagnostics: diagnostics}
	}

	// Every root must be in the program. One that is not would be a file silently left unchecked.
	byPath := make(map[tspath.PathKey]*ast.SourceFile)
	for _, sourceFile := range built.program.GetSourceFiles() {
		byPath[sourceFile.PathKey()] = sourceFile
	}
	for _, root := range built.roots[:len(paths)] {
		sourceFile, isLoaded := byPath[built.fileSystem.CaseSensitivity().PathKey(root.AsPath())]
		if !isLoaded {
			return nil, fmt.Errorf("load: %s was named but the compiler did not load it", built.fs.displayName(root))
		}
		loaded.files = append(loaded.files, sourceFile)
	}
	return loaded, nil
}

// programBuild is a program load built, with what it built it from: its source view, roots, directory and file system.
type programBuild struct {
	fs               *sourceFS
	roots            []tspath.RootedFilePath
	currentDirectory tspath.RootedDirectoryPath
	fileSystem       vfs.FS
	program          *compiler.Program
}

// firstProgram is the program of the given files, whatever they import, the prelude with its standalone console, and
// Apple's binding files they need, parsed and resolved but not checked. Its roots are the files in the order given,
// then the console, then the prelude, then Apple's.
func firstProgram(paths []string, overlay map[string]string, display func(string) string) (*programBuild, error) {
	if len(paths) == 0 {
		return nil, errors.New("load: no files given")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("load: resolving the current directory: %w", err)
	}
	currentDirectory := tspath.RootedDirectoryPathFromAbsolute(workingDirectory)

	normalizedOverlay := make(map[tspath.RootedFilePath]string, len(overlay))
	for name, source := range overlay {
		normalizedOverlay[currentDirectory.ResolveFile(name)] = source
	}
	fs := &sourceFS{FS: osvfs.FS(), overlay: normalizedOverlay, display: display}
	roots := make([]tspath.RootedFilePath, 0, len(paths)+2)
	for _, path := range paths {
		root, err := rootFileName(fs, currentDirectory, path)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	// The console first, then the rest of the prelude: the order their declarations had when they were one file, the
	// console its first statement. A Node program puts Node's console in the console's place.
	roots = append(roots, preludeConsolePath, preludePath)
	// Apple's binding files, for the apple/ modules the program imports (apple.go).
	apple, err := appleRoots(fs, roots[:len(paths)], normalizedOverlay)
	if err != nil {
		return nil, err
	}
	roots = append(roots, apple...)

	// bundled.WrapFS lays the embedded lib.*.d.ts files over the source view, and cachedvfs memoizes
	// the stats module resolution repeats.
	fileSystem := cachedvfs.From(&appleFS{FS: &regexpLibraryFS{FS: &nodeLibraryFS{FS: bundled.WrapFS(fs)}}})
	config := tsoptions.NewParsedCommandLine(compilerOptions(), roots, nil, currentDirectory, fileSystem.CaseSensitivity())
	host := compiler.NewCachedFSCompilerHost(fileSystem, bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})
	if program == nil {
		return nil, errors.New("load: the compiler built no program")
	}
	return &programBuild{fs: fs, roots: roots, currentDirectory: currentDirectory, fileSystem: fileSystem, program: program}, nil
}

// Unchecked is the program Load builds from paths before it asks UsesNodeModules whether Node's declarations join it:
// the files, whatever they import, the prelude with its standalone console and Apple's binding files, parsed and
// resolved but never checked, so a file the checker would refuse still has one. It is the generator's road to the root
// tsconfig.node.json's files (internal/load/projects), so that list and Load answer from one predicate.
func Unchecked(paths []string) (*compiler.Program, error) {
	first, err := firstProgram(paths, nil, nil)
	if err != nil {
		return nil, err
	}
	return first.program, nil
}

// Files is the program's own source files, in the order Load was given them: no prelude, no lib.
func (p *Program) Files() []*ast.SourceFile {
	return p.files
}

// SourceFiles is every source file the program holds but the prelude and the bundled library: its roots,
// every module they import, and every file they only import types from, declaration files among them, which
// Files and moduleOrder's runnable modules both leave out. A walk that must see every type the program
// declares reads it (#b5w3ycg head 2's kind carriers, Lower's must).
func (p *Program) SourceFiles() []*ast.SourceFile {
	files := []*ast.SourceFile{}
	for _, sourceFile := range p.compiler.GetSourceFiles() {
		if !IsPrelude(sourceFile) && !IsLibrary(sourceFile) {
			files = append(files, sourceFile)
		}
	}
	return files
}

// AppleFiles is the program's binding files for Apple's frameworks: the apple/ modules it loaded.
func (p *Program) AppleFiles() []*ast.SourceFile {
	files := []*ast.SourceFile{}
	for _, sourceFile := range p.compiler.GetSourceFiles() {
		if IsApple(sourceFile) {
			files = append(files, sourceFile)
		}
	}
	return files
}

// Checker returns the checker that owns a file, and the function that releases it.
func (p *Program) Checker(ctx context.Context, sourceFile *ast.SourceFile) (*checker.Checker, func()) {
	return p.compiler.GetTypeCheckerForFile(ctx, sourceFile)
}

// FileName is a source file's name as written: a .a file is named .a, never by the .a.ts the checker
// knows it as.
func (p *Program) FileName(sourceFile *ast.SourceFile) string {
	return p.fs.displayName(sourceFile.FileName())
}

// Where is a node's position as people write it: file:line:column, the file named as written.
func (p *Program) Where(node *ast.Node) string {
	sourceFile := ast.GetSourceFileOfNode(node)
	line, column := p.lineAndColumn(sourceFile, scanner.GetTokenPosOfNode(node, sourceFile, false))
	return fmt.Sprintf("%s:%d:%d", p.fs.displayName(sourceFile.FileName()), line, column)
}

// Line is the one-based line where a node's first token starts.
func (p *Program) Line(node *ast.Node) int {
	file := ast.GetSourceFileOfNode(node)
	line, _ := p.lineAndColumn(file, scanner.GetTokenPosOfNode(node, file, false))
	return line
}

// IsPrelude reports whether a declaration comes from Adamic's prelude rather than from the program,
// so a local named console is never mistaken for the real one. The prelude is every file its directory
// serves (source_fs.go): prelude.d.ts and the console its program has, standalone or Node's.
func IsPrelude(sourceFile *ast.SourceFile) bool {
	return sourceFile != nil && strings.HasPrefix(sourceFile.FileName().AsString(), preludeDirectory+"/")
}

// IsLibrary reports whether a declaration comes from TypeScript's bundled library (lib.es2024.d.ts and
// the files it includes), which is where Math, Array and String are declared.
func IsLibrary(sourceFile *ast.SourceFile) bool {
	return sourceFile != nil && strings.HasPrefix(sourceFile.FileName().AsString(), libraryPrefix())
}

// libraryPrefix is bundled.LibPath() as a string, read once. The library's root is fixed for the process (a
// constant scheme path when the library is embedded, the bundled package's own directory otherwise), and rooting
// it again on every IsLibrary was 3.4% of tsc's whole entry census, 48 s of 1,413 (#715zjph).
var libraryPrefix = sync.OnceValue(func() string { return bundled.LibPath().AsString() })

// rootFileName is the name the checker knows a source file by, refusing anything that isn't Adamic.
func rootFileName(fs *sourceFS, currentDirectory tspath.RootedDirectoryPath, path string) (tspath.RootedFilePath, error) {
	fileName := currentDirectory.ResolveFile(path)
	switch filepath.Ext(fileName.AsString()) {
	case ".ts":
		if !fs.FS.FileExists(fileName) {
			return "", fmt.Errorf("load: no file at %s", path)
		}
		return fileName, nil
	case ".a":
		if !fs.FS.FileExists(fileName) {
			return "", fmt.Errorf("load: no file at %s", path)
		}
		if fs.FS.FileExists(fileName.AppendSuffix(".ts")) {
			return "", fmt.Errorf("load: %s and %s.ts both exist, and an import of %s could mean either; rename one", path, path, filepath.Base(path))
		}
		return fileName.AppendSuffix(".ts"), nil
	}
	return "", fmt.Errorf("load: %s is not an Adamic source; want a .ts or .a file", path)
}

// diagnostics is everything the checker says about the program, formatted, in a stable order.
//
// Syntax first, and only syntax when there is any: a file that doesn't parse produces cascades from
// the later phases that bury the real error.
func (p *Program) diagnostics(ctx context.Context) []string {
	all := p.compiler.GetSyntacticDiagnostics(ctx, nil)
	if len(all) == 0 {
		all = append(all, p.compiler.GetConfigFileParsingDiagnostics()...)
		all = append(all, p.compiler.GetProgramDiagnostics()...)
		all = append(all, p.compiler.GetGlobalDiagnostics(ctx)...)
		all = append(all, p.compiler.GetBindDiagnostics(ctx, nil)...)
		all = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...)
	}
	formatted := make([]string, 0, len(all))
	for _, diagnostic := range all {
		formatted = append(formatted, p.formatDiagnostic(diagnostic))
	}
	sort.Strings(formatted)
	return formatted
}

// formatDiagnostic writes one diagnostic as file:line:column: error TS<code>: message, with the
// checker's elaboration indented beneath it.
func (p *Program) formatDiagnostic(diagnostic *ast.Diagnostic) string {
	var builder strings.Builder
	if diagnostic.File() != nil {
		line, column := p.lineAndColumn(diagnostic.File(), diagnostic.Pos())
		fmt.Fprintf(&builder, "%s:%d:%d: ", p.fs.displayName(diagnostic.File().FileName()), line, column)
	}
	message := starCollision(diagnostic)
	if message == "" {
		message = diagnostic.Localize(english)
	}
	fmt.Fprintf(&builder, "error TS%d: %s", diagnostic.Code(), message)
	writeChain(&builder, diagnostic.MessageChain(), 1)
	return builder.String()
}

// english is the compiler's own message text. A diagnostic stores a message key and its arguments,
// not text (MessageText is empty for most of them), and the zero locale is upstream's default.
var english locale.Locale

func writeChain(builder *strings.Builder, chain []*ast.Diagnostic, depth int) {
	for _, link := range chain {
		fmt.Fprintf(builder, "\n%s%s", strings.Repeat("  ", depth), link.Localize(english))
		writeChain(builder, link.MessageChain(), depth+1)
	}
}

// lineAndColumn is a position as people count it: lines from 1, and columns from 1 in UTF-16 code
// units, the way tsc and every editor report them.
func (p *Program) lineAndColumn(sourceFile *ast.SourceFile, position int) (int, int) {
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	line := sort.Search(len(lineStarts), func(index int) bool { return int(lineStarts[index]) > position }) - 1
	line = max(line, 0)
	prefix := sourceFile.Text()[int(lineStarts[line]):position]
	return line + 1, len(utf16.Encode([]rune(prefix))) + 1
}

// CompilerProgram exposes the checked program to public checker and lint adapters.
func (p *Program) CompilerProgram() *compiler.Program { return p.compiler }
