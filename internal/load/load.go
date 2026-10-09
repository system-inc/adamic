// Package load turns Adamic source files into a checked program: every file parsed, bound and
// type-checked by typescript-go in this process. Adamic sources use its defaults;
// TypeScript project roots use their project's checking options and libraries,
// with source-extension imports enabled and TypeScript emission disabled.
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
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
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

	// A configured TypeScript check reports only the explicitly requested files.
	requestedDiagnostics bool
}

// CheckError is a program the checker rejected, with every diagnostic it gave.
type CheckError struct {
	Diagnostics []string
}

func (e *CheckError) Error() string {
	return strings.Join(e.Diagnostics, "\n")
}

// compilerOptions supplies standalone Adamic defaults, including unconfigured inputs.
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
	return load(paths, nil)
}

// LoadOverlay checks paths while replacing the named files with source held by the caller. The
// overlay is read-only and never written to disk; names are absolute or relative to the current
// directory. This is for tools which test semantics-preserving source adaptations.
func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) {
	return load(paths, overlay)
}

func load(paths []string, overlay map[string]string) (*Program, error) {
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
	fs := &sourceFS{FS: osvfs.FS(), overlay: normalizedOverlay}
	roots := make([]tspath.RootedFilePath, 0, len(paths)+1)
	for _, path := range paths {
		root, err := rootFileName(fs, currentDirectory, path)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	userRoots := roots

	// bundled.WrapFS lays the embedded lib.*.d.ts files over the source view, and cachedvfs memoizes
	// the stats module resolution repeats.
	fileSystem := cachedvfs.From(&regexpLibraryFS{FS: &nodeLibraryFS{FS: bundled.WrapFS(fs)}})
	config, project, err := projectConfig(fs, currentDirectory, paths)
	if err != nil {
		return nil, err
	}
	fs.projectConsole = config != nil
	if config == nil {
		config = tsoptions.NewParsedCommandLine(compilerOptions(), roots, nil, currentDirectory, fileSystem.CaseSensitivity())
	} else {
		// Composite membership is a property of the whole project, even when a
		// census asks for diagnostics from only one file. Keep explicit ambient
		// roots too, and deduplicate the embedded prelude and Node declarations.
		seen := make(map[tspath.RootedFilePath]bool, len(roots))
		for _, name := range roots {
			seen[name] = true
		}
		for _, name := range config.FileNames() {
			if name.IsDeclarationFile() {
				text, _ := fs.ReadFile(name)
				if text == prelude || callerNodeTypes(name.AsString()) {
					continue
				}
			}
			if strings.HasSuffix(name.AsString(), ".a") {
				name = name.AppendSuffix(".ts")
			}
			if !seen[name] {
				roots = append(roots, name)
				seen[name] = true
			}
		}
	}
	roots = append(roots, preludePath)
	if !project {
		roots = append(roots, setPreludePath)
	}
	config = config.WithFileNames(roots)
	host := compiler.NewCachedFSCompilerHost(fileSystem, bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})
	if program == nil {
		return nil, errors.New("load: the compiler built no program")
	}

	if usesNodeModules(program) {
		index, err := nodeTypesIndex()
		if err != nil {
			return nil, err
		}
		fs.nodeTypes = true
		roots = append(roots, index)
		options := *config.CompilerOptions()
		// Preserve the project's other type roots, including the checker's
		// default ancestor search when no explicit typeRoots was supplied.
		if options.TypeRoots == nil {
			directory := workingDirectory
			if options.ConfigFilePath != "" {
				directory = filepath.Dir(options.ConfigFilePath.AsString())
			}
			for {
				options.TypeRoots = append(options.TypeRoots, tspath.RootedDirectoryPathFromAbsolute(filepath.Join(directory, "node_modules", "@types")))
				parent := filepath.Dir(directory)
				if parent == directory {
					break
				}
				directory = parent
			}
		}
		options.TypeRoots = append([]tspath.RootedDirectoryPath{tspath.RootedDirectoryPathFromNormalized(nodeTypesRoot)}, options.TypeRoots...)
		config.SetCompilerOptions(&options)
		fileSystem = cachedvfs.From(&regexpLibraryFS{FS: &nodeLibraryFS{FS: bundled.WrapFS(fs)}})
		config = config.WithFileNames(roots)
		host = compiler.NewCachedFSCompilerHost(fileSystem, bundled.LibPath(), nil, nil, nil)
		program = compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
		if program == nil {
			return nil, errors.New("load: the compiler built no Node program")
		}
	}

	loaded := &Program{compiler: program, fs: fs, requestedDiagnostics: fs.projectConsole}

	// Every root must be in the program. One that is not would be a file silently left unchecked.
	byPath := make(map[tspath.PathKey]*ast.SourceFile)
	for _, sourceFile := range program.GetSourceFiles() {
		byPath[sourceFile.PathKey()] = sourceFile
	}
	for _, root := range userRoots {
		sourceFile, isLoaded := byPath[fileSystem.CaseSensitivity().PathKey(root.AsPath())]
		if !isLoaded {
			return nil, fmt.Errorf("load: %s was named but the compiler did not load it", fs.displayName(root))
		}
		loaded.files = append(loaded.files, sourceFile)
	}
	if diagnostics := loaded.diagnostics(context.Background()); len(diagnostics) > 0 {
		return nil, &CheckError{Diagnostics: diagnostics}
	}
	return loaded, nil
}

// Files is the program's own source files, in the order Load was given them: no prelude, no lib.
func (p *Program) Files() []*ast.SourceFile {
	return p.files
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

// IsPrelude reports whether a declaration comes from Adamic's prelude rather than from the program,
// so a local named console is never mistaken for the real one.
func IsPrelude(sourceFile *ast.SourceFile) bool {
	return sourceFile != nil && (sourceFile.FileName() == preludePath || sourceFile.FileName() == setPreludePath)
}

// IsLibrary reports whether a declaration comes from TypeScript's bundled library (lib.es2024.d.ts and
// the files it includes), which is where Math, Array and String are declared.
func IsLibrary(sourceFile *ast.SourceFile) bool {
	return sourceFile != nil && strings.HasPrefix(sourceFile.FileName().AsString(), bundled.LibPath().AsString())
}

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
	// Unconfigured inputs still check their entire import graph. Configured
	// project inputs retain that graph but select diagnostics by requested file.
	collect := func(get func(context.Context, *ast.SourceFile) []*ast.Diagnostic) []*ast.Diagnostic {
		if !p.requestedDiagnostics {
			return get(ctx, nil)
		}
		var result []*ast.Diagnostic
		for _, file := range p.files {
			result = append(result, get(ctx, file)...)
		}
		return result
	}
	all := collect(p.compiler.GetSyntacticDiagnostics)
	if len(all) == 0 {
		all = append(all, p.compiler.GetConfigFileParsingDiagnostics()...)
		for _, diagnostic := range p.compiler.GetProgramDiagnostics() {
			if p.requestedDiagnostics && diagnostic.File() != nil {
				requested := false
				for _, file := range p.files {
					requested = requested || diagnostic.File() == file
				}
				if !requested {
					continue
				}
			}
			all = append(all, diagnostic)
		}
		all = append(all, p.compiler.GetGlobalDiagnostics(ctx)...)
		all = append(all, collect(p.compiler.GetBindDiagnostics)...)
		all = append(all, collect(p.compiler.GetSemanticDiagnostics)...)
		if len(all) == 0 && p.compiler.Options().GetEmitDeclarations() {
			all = append(all, collect(p.compiler.GetDeclarationDiagnostics)...)
		}
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
