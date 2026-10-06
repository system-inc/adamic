// Package checker is the pinned typescript-go checker used by the C bridge.
package checker

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// Program owns the compiler and its checker pool. Dropping it releases Go's roots.
// The external checker runs Go's runtime and collector; Adamic values never enter it.
type Program struct {
	Compiler    *compiler.Program
	typeIDs     map[*checker.Type]uint64
	symbolIDs   map[*ast.Symbol]uint64
	symbolsByID []*ast.Symbol
	typesByID   []*checker.Type
	exactRanges map[*ast.SourceFile]map[nodeRange][]*ast.Node
}
type Result struct {
	Kind         uint32
	Symbol, Type string
}

func Path(path string) (string, error) {
	if !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return "", fmt.Errorf("path must be UTF-8 without NUL")
	}
	absolute, err := filepath.Abs(path)
	return filepath.ToSlash(absolute), err
}

func Open(configPath string, files []string) (*Program, error) {
	configPath, err := Path(configPath)
	if err != nil {
		return nil, err
	}
	directory := filepath.ToSlash(filepath.Dir(configPath))
	fs := cachedvfs.From(bundled.WrapFS(osvfs.FS()))
	host := compiler.NewCachedFSCompilerHost(directory, fs, bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(configPath, nil, nil, host, nil)
	if config == nil {
		return nil, fmt.Errorf("cannot parse %s: %s", configPath, diagnosticText(diagnostics))
	}
	if len(config.Errors) != 0 {
		return nil, fmt.Errorf("invalid config: %s", diagnosticText(config.Errors))
	}
	if len(diagnostics) != 0 {
		return nil, fmt.Errorf("invalid config: %s", diagnosticText(diagnostics))
	}
	if len(files) > 0 {
		roots := make([]string, len(files))
		for index, file := range files {
			if !filepath.IsAbs(file) {
				file = filepath.Join(directory, file)
			}
			roots[index], err = Path(file)
			if err != nil {
				return nil, err
			}
			if !fs.FileExists(roots[index]) {
				return nil, fmt.Errorf("no file at %s", roots[index])
			}
		}
		for _, root := range roots {
			if strings.HasSuffix(root, ".a") {
				config.CompilerOptions().AllowNonTsExtensions = core.TSTrue
			}
		}
		for _, declaration := range config.FileNames() {
			if strings.HasSuffix(declaration, ".d.ts") {
				found := false
				for _, root := range roots {
					if root == declaration {
						found = true
					}
				}
				if !found {
					roots = append(roots, declaration)
				}
			}
		}
		config = config.WithFileNames(roots)
	}
	if len(config.ContentMappers()) != 0 {
		return nil, fmt.Errorf("content mappers are not supported by this bridge")
	}
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	if program == nil {
		return nil, fmt.Errorf("checker built no program")
	}
	for _, file := range config.FileNames() {
		if program.GetSourceFile(file) == nil {
			return nil, fmt.Errorf("checker did not load %s", file)
		}
	}
	// Initialize the checker pool in the load phase. Type resolution stays lazy.
	if len(config.FileNames()) != 0 {
		_, release := program.GetTypeCheckerForFile(context.Background(), program.GetSourceFile(config.FileNames()[0]))
		release()
	}
	return &Program{Compiler: program, typeIDs: make(map[*checker.Type]uint64)}, nil
}

func diagnosticText(diagnostics []*ast.Diagnostic) string {
	var english locale.Locale
	messages := make([]string, len(diagnostics))
	for index, diagnostic := range diagnostics {
		messages[index] = diagnostic.Localize(english)
	}
	return strings.Join(messages, "; ")
}

// Query selects the deepest AST child containing a zero-based UTF-8 byte offset.
// Ranges are half open; leading trivia belongs to the following node. EOF is refused.
// Type errors in the queried TypeScript are allowed, as they are in a lint checker.
func (p *Program) Query(file string, position uint64) (Result, error) {
	file, err := Path(file)
	if err != nil {
		return Result{}, err
	}
	source := p.Compiler.GetSourceFile(file)
	if source == nil {
		return Result{}, fmt.Errorf("file is not in this program: %s", file)
	}
	if position >= uint64(len(source.Text())) {
		return Result{}, fmt.Errorf("position %d is outside %s", position, file)
	}
	node := ast.GetNodeAtPosition(source, int(position), false)
	typeChecker, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	result := Result{Kind: uint32(node.Kind), Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(node))}
	if symbol := typeChecker.GetSymbolAtLocation(node); symbol != nil {
		result.Symbol = ast.EscapeAllInternalSymbolNames(ast.SymbolName(symbol))
	}
	return result, nil
}

// TypeParts identifies an exact AST node, rather than the deepest token at a
// position. A mismatch between parsers is refused instead of querying a neighbor.
// Frames are flags LF UTF-16-length LF TypeToString, with no trailing separator.
func (p *Program) TypeParts(file string, start, end uint64, kind string) (string, error) {
	file, err := Path(file)
	if err != nil {
		return "", err
	}
	source := p.Compiler.GetSourceFile(file)
	if source == nil {
		return "", fmt.Errorf("file is not in this program: %s", file)
	}
	if start >= end || end > uint64(len(source.Text())) {
		return "", fmt.Errorf("invalid node range")
	}
	// GetNodeAtPosition deliberately skips token kinds (including identifiers).
	// Descend by ranges ourselves so an operand token is selectable too.
	var node *ast.Node
	var find func(*ast.Node) bool
	find = func(candidate *ast.Node) bool {
		if uint64(candidate.Pos()) > start || uint64(candidate.End()) < end {
			return false
		}
		if uint64(candidate.Pos()) == start && uint64(candidate.End()) == end && strings.TrimPrefix(candidate.Kind.String(), "Kind") == kind {
			node = candidate
			return true
		}
		return candidate.ForEachChild(find)
	}
	find(source.AsNode())
	if node == nil {
		return "", fmt.Errorf("no exact %s node at %d:%d in %s", kind, start, end, file)
	}
	typeChecker, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	actual := typeChecker.GetTypeAtLocation(node)
	if constraint := checker.Checker_getBaseConstraintOfType(typeChecker, actual); constraint != nil {
		actual = constraint
	}
	parts := []*checker.Type{actual}
	if actual.Flags()&checker.TypeFlagsUnion != 0 {
		parts = actual.Types()
	}
	var framed strings.Builder
	for _, part := range parts {
		name := typeChecker.TypeToString(part)
		fmt.Fprintf(&framed, "%d\n%d\n%s", part.Flags(), len(utf16.Encode([]rune(name))), name)
	}
	return framed.String(), nil
}
