package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/load"
)

const returnRewrite = "explicit undefined returns in annotated functions"

var returnDiagnostic = regexp.MustCompile(`^([^\n]+):([0-9]+):([0-9]+): error TS7030:`)

// Only a declared return type containing undefined proves the contract. Inferred
// returns, void, any and unknown are not permission to change that contract.
// void 0 denotes undefined even when the identifier undefined is shadowed.
func returnAdaptations(paths []string, overlay map[string]string, before error) (map[string]string, int, error) {
	var diagnostics *load.CheckError
	if !errors.As(before, &diagnostics) || diagnosticCount(before, "7030") == 0 {
		return overlay, 0, nil
	}
	program, owned, err := optionalProgram(paths, overlay)
	if err != nil {
		return nil, 0, err
	}
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		files[file.FileName()] = file
	}
	selected := make(map[*ast.Node]bool)
	for _, diagnostic := range diagnostics.Diagnostics {
		match := returnDiagnostic.FindStringSubmatch(diagnostic)
		if match == nil {
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
		if file == nil || !owned[name] {
			continue
		}
		line, _ := strconv.Atoi(match[2])
		column, _ := strconv.Atoi(match[3])
		pos := optionalPosition(file, line, column)
		if pos < 0 {
			return nil, 0, fmt.Errorf("invalid return diagnostic position %s", match[0])
		}
		node := optionalNodeAt(file.AsNode(), pos)
		for node != nil && !ast.IsFunctionLike(node) {
			node = node.Parent
		}
		if node == nil || node.Type() == nil || node.Body() == nil || node.Body().Kind != ast.KindBlock || ast.GetFunctionFlags(node) != ast.FunctionFlagsNormal {
			continue
		}
		c, release := program.GetTypeCheckerForFile(context.Background(), file)
		typ := c.GetTypeFromTypeNode(node.Type())
		includesUndefined := false
		for _, part := range typ.Distributed() {
			if part.Flags()&checker.TypeFlagsUndefined != 0 {
				includesUndefined = true
			}
		}
		release()
		if includesUndefined {
			selected[node] = true
		}
	}
	edits := make(map[*ast.SourceFile][]optionalEdit)
	for node := range selected {
		file := ast.GetSourceFileOfNode(node)
		body := node.Body()
		end := body.End() - 1
		if end < 0 || file.Text()[end] != '}' {
			return nil, 0, fmt.Errorf("invalid return body in %s", file.FileName())
		}
		edits[file] = append(edits[file], optionalEdit{start: end, end: end, replacement: "\nreturn void 0;\n"})
		// Bare returns also complete with undefined. Keep their completion kind and
		// finally behavior; never descend into a nested function's returns.
		var visit func(*ast.Node) bool
		visit = func(child *ast.Node) bool {
			if ast.IsFunctionLike(child) {
				return false
			}
			if child.Kind == ast.KindReturnStatement && child.Expression() == nil {
				start := scanner.GetTokenPosOfNode(child, file, false) + len("return")
				edits[file] = append(edits[file], optionalEdit{start: start, end: start, replacement: " void 0"})
			}
			child.ForEachChild(visit)
			return false
		}
		body.ForEachChild(visit)
	}
	if len(selected) == 0 {
		return overlay, 0, nil
	}
	if overlay == nil {
		overlay = make(map[string]string)
	}
	for file, changes := range edits {
		sort.Slice(changes, func(i, j int) bool { return changes[i].start > changes[j].start })
		source := file.Text()
		for _, change := range changes {
			source = source[:change.start] + change.replacement + source[change.end:]
		}
		name := file.FileName()
		if strings.HasSuffix(name, ".a.ts") && owned[strings.TrimSuffix(name, ".ts")] {
			name = strings.TrimSuffix(name, ".ts")
		}
		overlay[name] = source
	}
	return overlay, len(selected), nil
}
