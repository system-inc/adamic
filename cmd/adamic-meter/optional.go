package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/load"
)

var optionalDiagnostic = regexp.MustCompile(`^([^\n]+):([0-9]+):([0-9]+): error TS(2412|2375|2379):`)

const optionalRewrite = "present undefined in optional property declarations"

// This program is used only to resolve declarations in rejected source. It must never reach
// lowering: LoadOverlay remains the authoritative checker gate after every adaptation.
// The options match load's fixed options; the resolver needs no Adamic prelude declarations.
func optionalProgram(paths []string, overlay map[string]string) (*compiler.Program, map[string]bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	owned := make(map[string]bool)
	fs := &optionalFS{FS: osvfs.FS(), source: make(map[string]string)}
	roots := make([]tspath.RootedFilePath, 0, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, err
		}
		name := filepath.ToSlash(absolute)
		owned[name] = true
		if strings.HasSuffix(name, ".a") {
			source, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, err
			}
			if fs.FS.FileExists(tspath.RootedFilePathFromAbsolute(name + ".ts")) {
				return nil, nil, fmt.Errorf("ambiguous adaptation source %s", name)
			}
			fs.source[name+".ts"] = string(source)
			owned[name+".ts"] = true
			name += ".ts"
		}
		roots = append(roots, tspath.RootedFilePathFromAbsolute(name))
	}
	for path, source := range overlay {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, err
		}
		name := filepath.ToSlash(absolute)
		if strings.HasSuffix(name, ".a") {
			name += ".ts"
		}
		fs.source[name] = source
	}
	options := &core.CompilerOptions{
		Strict: core.TSTrue, NoUncheckedIndexedAccess: core.TSTrue, ExactOptionalPropertyTypes: core.TSTrue,
		NoImplicitReturns: core.TSTrue, NoFallthroughCasesInSwitch: core.TSTrue, ErasableSyntaxOnly: core.TSTrue,
		VerbatimModuleSyntax: core.TSTrue, AllowImportingTsExtensions: core.TSTrue, NoEmit: core.TSTrue,
		Module: core.ModuleKindESNext, ModuleDetection: core.ModuleDetectionKindForce,
		ModuleResolution: core.ModuleResolutionKindBundler, Target: core.ScriptTargetES2024,
		Lib: []string{"lib.es2024.d.ts"}, Types: []string{},
	}
	fileSystem := cachedvfs.From(bundled.WrapFS(fs))
	config := tsoptions.NewParsedCommandLine(options, roots, nil, tspath.RootedDirectoryPathFromAbsolute(filepath.ToSlash(cwd)), fileSystem.CaseSensitivity())
	host := compiler.NewCachedFSCompilerHost(fileSystem, bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	if program == nil {
		return nil, nil, errors.New("building optional-property declaration resolver")
	}
	return program, owned, nil
}

type optionalFS struct {
	vfs.FS
	source map[string]string
}

func (fs *optionalFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	if source, ok := fs.source[path.AsString()]; ok {
		return source, true
	}
	return fs.FS.ReadFile(path)
}
func (fs *optionalFS) FileExists(path tspath.RootedFilePath) bool {
	_, ok := fs.source[path.AsString()]
	return ok || fs.FS.FileExists(path)
}

type optionalEdit struct {
	start, end  int
	replacement string
	ownerStart  int
}

func optionalAdaptations(paths []string, overlay map[string]string, before error) (map[string]string, int, error) {
	var diagnostics *load.CheckError
	if !errors.As(before, &diagnostics) {
		return overlay, 0, nil
	}
	relevant := false
	for _, d := range diagnostics.Diagnostics {
		if optionalDiagnostic.MatchString(d) {
			relevant = true
			break
		}
	}
	if !relevant {
		return overlay, 0, nil
	}
	program, owned, err := optionalProgram(paths, overlay)
	if err != nil {
		return nil, 0, err
	}
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		files[file.FileName().AsString()] = file
	}
	selected := make(map[*ast.Node]bool)
	for _, diagnostic := range diagnostics.Diagnostics {
		match := optionalDiagnostic.FindStringSubmatch(diagnostic)
		if match == nil || strings.Contains(diagnostic, "could be instantiated") {
			continue
		}
		name, err := filepath.Abs(match[1])
		if err != nil {
			return nil, 0, err
		}
		name = filepath.ToSlash(name)
		if strings.HasSuffix(name, ".a") {
			name += ".ts"
		}
		file := files[name]
		if file == nil {
			continue
		}
		line, _ := strconv.Atoi(match[2])
		column, _ := strconv.Atoi(match[3])
		pos := optionalPosition(file, line, column)
		if pos < 0 {
			return nil, 0, fmt.Errorf("invalid optional diagnostic position %s", match[0])
		}
		node := optionalNodeAt(file.AsNode(), pos)
		c, release := program.GetTypeCheckerForFile(context.Background(), file)
		if match[4] == "2412" {
			for n := node; n != nil; n = n.Parent {
				if n.Kind != ast.KindBinaryExpression {
					continue
				}
				assignment := n.AsBinaryExpression()
				if assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Left.Pos() <= pos && pos < assignment.Left.End() && !unprovenIndexRead(c, assignment.Right) && !genericReceiver(c, assignment.Left) {
					selectOptional(c, c.GetSymbolAtLocation(assignment.Left), c.GetTypeAtLocation(assignment.Right), owned, selected)
				}
				break
			}
		} else {
			// A contextual object view covers literals, forwarded options, returns, and arguments.
			// A required field, a union discriminator, or a generic target is not an optional slot.
			for n := node; n != nil; n = n.Parent {
				expression := contextualExpression(n)
				if expression == nil {
					continue
				}
				target := c.GetContextualType(expression, checker.ContextFlagsNone)
				if target == nil || target.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsAnyOrUnknown) != 0 {
					continue
				}
				properties := c.GetPropertiesOfType(target)
				if len(properties) == 0 {
					continue
				}
				source := c.GetTypeAtLocation(expression)
				for _, property := range properties {
					actual := c.GetPropertyOfType(source, property.Name)
					if actual == nil {
						continue
					}
					value := c.GetTypeOfSymbolAtLocation(actual, expression)
					if len(actual.Declarations) == 1 {
						declaration := actual.Declarations[0]
						if declaration.Kind == ast.KindPropertyAssignment && unprovenIndexRead(c, declaration.AsPropertyAssignment().Initializer) {
							continue
						}
					}
					// An optional source property without an explicit undefined union does not supply
					// evidence of a present-undefined write. Its read type includes absent/missing too.
					if actual.Flags&ast.SymbolFlagsOptional != 0 {
						if len(actual.Declarations) != 1 || actual.Declarations[0].Type() == nil {
							continue
						}
						value = c.GetTypeFromTypeNode(actual.Declarations[0].Type())
					}
					selectOptional(c, property, value, owned, selected)
				}
				break
			}
		}
		release()
	}
	edits := make(map[string][]optionalEdit)
	sources := make(map[string]string)
	for declaration := range selected {
		file := ast.GetSourceFileOfNode(declaration)
		name := file.FileName().AsString()
		if strings.HasSuffix(name, ".a.ts") && owned[strings.TrimSuffix(name, ".ts")] {
			name = strings.TrimSuffix(name, ".ts")
		}
		sources[name] = file.Text()
		typ := declaration.Type()
		start := scanner.GetTokenPosOfNode(typ, file, false)
		ownerStart := declaration.Pos()
		edits[name] = append(edits[name], optionalEdit{typ.End(), typ.End(), ") | undefined", ownerStart})
		if declaration.Kind == ast.KindMethodSignature {
			start = declaration.AsMethodSignatureDeclaration().PostfixToken.End()
			colon := typ.Pos() - 1
			// Type.Pos includes the trivia after the return colon. Validate the delimiter;
			// never interpret the method's return annotation as the callable slot type.
			if colon < start || file.Text()[colon] != ':' {
				return nil, 0, fmt.Errorf("invalid method signature delimiter in %s", name)
			}
			edits[name] = append(edits[name], optionalEdit{start, start, ": (", ownerStart}, optionalEdit{colon, colon + 1, " =>", ownerStart})
		} else {
			edits[name] = append(edits[name], optionalEdit{start, start, "(", ownerStart})
		}
	}
	if overlay == nil {
		overlay = make(map[string]string)
	}
	count := len(selected)
	for path, list := range edits {
		// Insert wrappers rather than replacing whole types, so nested optional declarations
		// compose. At a shared end, insert the outer suffix first and the inner one before it.
		sort.Slice(list, func(i, j int) bool {
			if list[i].start != list[j].start {
				return list[i].start > list[j].start
			}
			if list[i].end != list[j].end {
				return list[i].end > list[j].end
			}
			return list[i].ownerStart < list[j].ownerStart
		})
		source := sources[path]
		boundary := len(source)
		for _, edit := range list {
			if edit.start < 0 || edit.end > boundary || edit.start > edit.end {
				return nil, 0, fmt.Errorf("overlapping optional declaration edits in %s", path)
			}
			source = source[:edit.start] + edit.replacement + source[edit.end:]
			boundary = edit.start
		}
		overlay[path] = source
	}
	return overlay, count, nil
}

func selectOptional(c *checker.Checker, symbol *ast.Symbol, value *checker.Type, owned map[string]bool, selected map[*ast.Node]bool) {
	if symbol == nil || symbol.Flags&ast.SymbolFlagsOptional == 0 || len(symbol.Declarations) == 0 || value == nil {
		return
	}
	parts := []*checker.Type{value}
	if value.Flags()&checker.TypeFlagsUnion != 0 {
		parts = value.AsUnionType().Types()
	}
	hasUndefined := false
	for _, part := range parts {
		if part.Flags()&checker.TypeFlagsUndefined != 0 {
			hasUndefined = true
		}
	}
	if !hasUndefined {
		return
	}
	// Resolve every declaration of a merged symbol. Never rewrite external libraries,
	// required properties, live methods/accessors, inferred slots, or a narrower generic contract.
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindPropertySignature && declaration.Kind != ast.KindPropertyDeclaration && declaration.Kind != ast.KindMethodSignature {
			return
		}
		if !ast.HasQuestionToken(declaration) || declaration.Type() == nil || !owned[ast.GetSourceFileOfNode(declaration).FileName().AsString()] {
			return
		}
		var target *checker.Type
		if declaration.Kind == ast.KindMethodSignature {
			if len(symbol.Declarations) != 1 {
				return
			}
			target = c.GetTypeOfSymbolAtLocation(symbol, declaration)
			parts := target.Distributed()
			present := make([]*checker.Type, 0, len(parts))
			for _, part := range parts {
				if part.Flags()&checker.TypeFlagsUndefined == 0 {
					present = append(present, part)
				}
			}
			target = c.GetUnionType(present)
		} else {
			target = c.GetTypeFromTypeNode(declaration.Type())
		}
		if target.Flags()&(checker.TypeFlagsAnyOrUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess) != 0 || c.IsTypeAssignableTo(c.GetUndefinedType(), target) {
			return
		}
		for _, part := range parts {
			if part.Flags()&checker.TypeFlagsUndefined != 0 {
				continue
			}
			if part.Flags()&(checker.TypeFlagsAnyOrUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess) != 0 || !c.IsTypeAssignableTo(part, target) {
				return
			}
		}
	}
	for _, declaration := range symbol.Declarations {
		selected[declaration] = true
	}
}

func hasIndexRead(node *ast.Node) bool {
	if node.Kind == ast.KindElementAccessExpression {
		return true
	}
	return node.ForEachChild(hasIndexRead)
}
func optionalNodeAt(node *ast.Node, pos int) *ast.Node {
	result := node
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Pos() <= pos && pos < child.End() {
			result = optionalNodeAt(child, pos)
			return true
		}
		return false
	})
	return result
}
func optionalPosition(file *ast.SourceFile, line, column int) int {
	starts := scanner.GetECMALineStarts(file)
	if line < 1 || line > len(starts) || column < 1 {
		return -1
	}
	pos := int(starts[line-1])
	units := 1
	for units < column && pos < len(file.Text()) {
		r, size := utf8.DecodeRuneInString(file.Text()[pos:])
		if r == '\n' || r == '\r' {
			return -1
		}
		units++
		if r > 0xffff {
			units++
		}
		pos += size
	}
	if units != column {
		return -1
	}
	return pos
}

func genericReceiver(c *checker.Checker, left *ast.Node) bool {
	if left.Kind != ast.KindPropertyAccessExpression {
		return true
	}
	receiver := c.GetTypeAtLocation(left.AsPropertyAccessExpression().Expression)
	return receiver.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess|checker.TypeFlagsAnyOrUnknown) != 0
}

func contextualExpression(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindReturnStatement:
		return node.AsReturnStatement().Expression
	case ast.KindVariableDeclaration:
		return node.AsVariableDeclaration().Initializer
	case ast.KindBinaryExpression:
		if node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			return node.AsBinaryExpression().Right
		}
	}
	if ast.IsExpression(node) {
		return node
	}
	return nil
}

// A missing-value branch is already part of the program's meaning. Its optional payload may
// truthfully include undefined even if another branch indexes an array. A bare required read
// still needs its own invariant and is never the seed for this mechanical adaptation.
func unprovenIndexRead(c *checker.Checker, node *ast.Node) bool {
	if !hasIndexRead(node) {
		return false
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		if c.GetTypeAtLocation(conditional.WhenTrue).Flags()&checker.TypeFlagsUndefined != 0 || c.GetTypeAtLocation(conditional.WhenFalse).Flags()&checker.TypeFlagsUndefined != 0 {
			return false
		}
	}
	return true
}
