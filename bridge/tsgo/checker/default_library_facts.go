package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"strconv"
	"strings"
)

// Mirrors cohere checking.IsSourceFileDefaultLibrary, including configured-lib
// fallback. The harness requires ReadsCompilerOptions and ReadsDefaultLibrary.
func (p *Program) sourceDefaultLibrary(file *ast.SourceFile) bool {
	if !file.IsDeclarationFile {
		return false
	}
	if p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()) {
		return true
	}
	options := p.Compiler.Options()
	if options.NoLib.IsTrue() {
		return false
	}
	var libs []string
	directory := p.Compiler.Host().DefaultLibraryPath().AsString()
	if options.Lib == nil {
		libs = append(libs, tspath.CombinePaths(directory, tsoptions.GetDefaultLibFileName(options)))
	} else {
		for _, lib := range options.Lib {
			if name, ok := tsoptions.GetLibFileName(lib); ok {
				libs = append(libs, tspath.CombinePaths(directory, name))
			}
		}
	}
	sensitive := tspath.CaseInsensitive
	if p.Compiler.UseCaseSensitiveFileNames() {
		sensitive = tspath.CaseSensitive
	}
	for _, lib := range libs {
		if tspath.ComparePaths(file.FileName().AsString(), lib, sensitive) == 0 {
			return true
		}
	}
	return false
}

// symbol-default-library v1: a root selector with a same-file byte span and kind,
// then one boolean: any declaration belongs to the program's default library.
func (p *Program) symbolDefaultLibrary(source *ast.SourceFile, node *ast.Node, c *checker.Checker, out *fields, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 4 || node != source.AsNode() {
		return "", fmt.Errorf("symbol-default-library requires a SourceFile selector")
	}
	start, e1 := strconv.ParseUint(parts[1], 10, 64)
	end, e2 := strconv.ParseUint(parts[2], 10, 64)
	if e1 != nil || e2 != nil || strconv.FormatUint(start, 10) != parts[1] || strconv.FormatUint(end, 10) != parts[2] {
		return "", fmt.Errorf("invalid library node selector")
	}
	_, selected, err := p.exact(source.FileName().AsString(), start, end, parts[3])
	if err != nil {
		return "", err
	}
	present := false
	if symbol := c.GetSymbolAtLocation(selected); symbol != nil {
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil {
				return "", fmt.Errorf("symbol declaration has no source")
			}
			if p.sourceDefaultLibrary(file) {
				present = true
				break
			}
		}
	}
	out.yes(present)
	return out.String(), nil
}
