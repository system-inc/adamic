// adamic-rename-dot-a prepares the integration rename without changing source by default.
package main

import (
	"flag"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

type edit struct {
	start, end    int
	before, after string
}
type change struct {
	name, source string
	edits        []edit
}
type plan struct {
	renames  map[string]string
	changes  []change
	excluded []string
}

func main() {
	apply := flag.Bool("apply", false, "apply the printed plan (default: dry run)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: adamic-rename-dot-a [--apply] repository-root")
		os.Exit(2)
	}
	root, err := filepath.Abs(flag.Arg(0))
	if err == nil {
		var p plan
		p, err = prepare(root)
		if err == nil {
			p.print()
			if *apply {
				err = p.apply(root)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Classification is by purpose, not by whether today's compiler accepts the program.
// Gap and refusal fixtures belong to the language suite and move with their ports.
func classification(name string) (bool, string) {
	if strings.HasSuffix(name, ".d.ts") {
		return false, "TypeScript ambient declaration, not an executable Adamic program"
	}
	if strings.HasPrefix(name, "cmd/adamic-meter/testdata/") {
		return false, "TypeScript input corpus for measuring rejection and source adaptation before the gate"
	}
	for _, prefix := range []string{"stage1/", "internal/load/testdata/0.1/", "internal/oracle/testdata/", "examples/", "dedication/", "bench/"} {
		if strings.HasPrefix(name, prefix) {
			return true, ""
		}
	}
	return false, ""
}

func prepare(root string) (plan, error) {
	p := plan{renames: map[string]string{}}
	output, err := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "-z").Output()
	if err != nil {
		return p, err
	}
	names := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	sort.Strings(names)
	names = slices.Compact(names)
	for _, name := range names {
		if _, err := os.Lstat(filepath.Join(root, name)); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return p, err
		}
		if !strings.HasSuffix(name, ".ts") || strings.HasPrefix(name, "cohere/") {
			continue
		}
		yes, reason := classification(name)
		if !yes {
			if reason == "" {
				return p, fmt.Errorf("unclassified TypeScript file %s: classify its purpose before applying", name)
			}
			p.excluded = append(p.excluded, name+": "+reason)
			continue
		}
		destination := strings.TrimSuffix(name, ".ts") + ".a"
		if _, err := os.Lstat(filepath.Join(root, destination)); err == nil {
			return p, fmt.Errorf("rename collision: %s already exists", destination)
		} else if !os.IsNotExist(err) {
			return p, err
		}
		p.renames[name] = destination
	}
	for _, name := range names {
		if name == "cohere" || strings.HasPrefix(name, "cohere/") || strings.HasPrefix(name, "cmd/adamic-rename-dot-a/") || name == "cloud/dot-a-dry-run.txt" || name == "docs/dot-a.md" {
			continue
		}
		extension := filepath.Ext(name)
		if extension != ".a" && extension != ".ts" && extension != ".go" && extension != ".md" && extension != ".json" && extension != ".mjs" && extension != ".sh" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return p, err
		}
		source := string(data)
		var edits []edit
		if extension == ".a" || extension == ".ts" {
			edits, err = sourceEdits(name, source, p.renames)
			if err != nil {
				return p, err
			}
		} else {
			edits = referenceEdits(name, source, p.renames)
			if extension == ".go" && strings.HasPrefix(name, "stage1/") {
				filters, filterError := sourceFilterEdits(name, source)
				if filterError != nil {
					return p, filterError
				}
				edits = append(edits, filters...)
				sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
			}
		}
		if len(edits) > 0 {
			p.changes = append(p.changes, change{name, source, edits})
		}
	}
	return p, nil
}

func sourceEdits(name, source string, renames map[string]string) ([]edit, error) {
	rooted := "/" + name
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: rooted, Path: tspath.Path(rooted)}, source, core.ScriptKindTS)
	if len(file.Diagnostics()) > 0 {
		return nil, fmt.Errorf("cannot parse %s: %d syntax diagnostics", name, len(file.Diagnostics()))
	}
	var edits []edit
	add := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindStringLiteral {
			return
		}
		value := node.Text()
		target := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(name), value)))
		if _, ok := renames[target]; !ok || !strings.HasPrefix(value, ".") {
			return
		}
		start := scanner.GetTokenPosOfNode(node, file, false)
		// Preserve the original quote and escape spelling: only the terminal extension changes.
		end := node.End() - 1
		if end < start+4 || source[end-3:end] != ".ts" {
			edits = append(edits, edit{start, node.End(), source[start:node.End()], strconv.Quote(strings.TrimSuffix(value, ".ts") + ".a")})
			return
		}
		edits = append(edits, edit{end - 3, end, ".ts", ".a"})
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindImportDeclaration:
			add(node.AsImportDeclaration().ModuleSpecifier)
		case ast.KindExportDeclaration:
			add(node.AsExportDeclaration().ModuleSpecifier)
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			if call.Expression.Kind == ast.KindImportKeyword && len(call.Arguments.Nodes) == 1 {
				add(call.Arguments.Nodes[0])
			}
		case ast.KindImportType:
			add(node.AsImportTypeNode().Argument.AsLiteralTypeNode().Literal)
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	return edits, nil
}

// Non-source references include prose, shell commands and Go fixture strings.
// Match path tokens, never TypeScript source. Resolve a reference against its containing
// directory and the repository, then against known suffixes for split Go filepath.Join calls.
var reference = regexp.MustCompile(`[A-Za-z0-9_./*+-]+\.ts`)

func referenceEdits(name, source string, renames map[string]string) []edit {
	var edits []edit
	for _, span := range reference.FindAllStringIndex(source, -1) {
		before := source[span[0]:span[1]]
		if strings.HasPrefix(strings.TrimLeft(before, "./"), "cohere/") || strings.HasSuffix(before, ".d.ts") || strings.HasSuffix(before, ".a.ts") {
			continue
		}
		target := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(name), before)))
		_, local := renames[target]
		_, root := renames[strings.TrimPrefix(before, "./")]
		known := local || root
		if !known {
			for old := range renames {
				if strings.HasSuffix(old, "/"+strings.TrimPrefix(before, "./")) {
					// Bare names in fixture tests and port reports are relative to the suite.
					if strings.HasPrefix(name, "stage1/") || strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "docs/") || strings.HasPrefix(name, "bench/") {
						known = true
						break
					}
				}
			}
		}
		if strings.Contains(before, "*") {
			directory := strings.TrimSuffix(before, "*.ts")
			for old := range renames {
				if strings.HasPrefix(old, filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(name), directory)))+"/") || strings.HasPrefix(old, directory) && directory != "" {
					known = true
					break
				}
			}
			if before == "*.ts" {
				for old := range renames {
					if filepath.Dir(old) == filepath.Dir(name) {
						known = true
						break
					}
				}
			}
		}
		if known {
			edits = append(edits, edit{span[0], span[1], before, strings.TrimSuffix(before, ".ts") + ".a"})
		}
	}
	return edits
}

// Corpus walkers still need upstream TypeScript, and must include renamed Adamic.
// Parse Go so parentheses and compound directory predicates stay correct.
func sourceFilterEdits(name, source string) ([]edit, error) {
	positions := token.NewFileSet()
	file, err := goparser.ParseFile(positions, name, source, 0)
	if err != nil {
		return nil, err
	}
	var edits []edit
	parents := map[goast.Node]goast.Node{}
	var stack []goast.Node
	goast.Inspect(file, func(node goast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	goast.Inspect(file, func(node goast.Node) bool {
		call, ok := node.(*goast.CallExpr)
		if !ok || len(call.Args) != 2 {
			return true
		}
		selector, ok := call.Fun.(*goast.SelectorExpr)
		if !ok || selector.Sel.Name != "HasSuffix" {
			return true
		}
		receiver, ok := selector.X.(*goast.Ident)
		if !ok || receiver.Name != "strings" {
			return true
		}
		literal, ok := call.Args[1].(*goast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil || value != ".ts" {
			return true
		}
		// Recognize the immediate alternative, not an unrelated .a filter elsewhere.
		if parent, ok := parents[call].(*goast.BinaryExpr); ok && parent.Op == token.LOR {
			alternative := parent.Y
			if parent.Y == call {
				alternative = parent.X
			}
			if other, ok := alternative.(*goast.CallExpr); ok && len(other.Args) == 2 {
				otherSelector, selected := other.Fun.(*goast.SelectorExpr)
				otherLiteral, literalArgument := other.Args[1].(*goast.BasicLit)
				if selected && literalArgument && otherSelector.Sel.Name == "HasSuffix" {
					otherReceiver, received := otherSelector.X.(*goast.Ident)
					otherValue, _ := strconv.Unquote(otherLiteral.Value)
					argument := source[positions.Position(call.Args[0].Pos()).Offset:positions.Position(call.Args[0].End()).Offset]
					otherArgument := source[positions.Position(other.Args[0].Pos()).Offset:positions.Position(other.Args[0].End()).Offset]
					if received && otherReceiver.Name == "strings" && otherValue == ".a" && argument == otherArgument {
						return true
					}
				}
			}
		}
		start, end := positions.Position(call.Pos()).Offset, positions.Position(call.End()).Offset
		argument := source[positions.Position(call.Args[0].Pos()).Offset:positions.Position(call.Args[0].End()).Offset]
		before := source[start:end]
		edits = append(edits, edit{start, end, before, "(" + before + " || strings.HasSuffix(" + argument + ", \".a\"))"})
		return true
	})
	return edits, nil
}

func (p plan) print() {
	fmt.Println("EXCLUDE cohere/: upstream submodule, including its TypeScript submodule")
	fmt.Println("EXCLUDE *.mjs: Node tooling, not Adamic source (references may be edited)")
	for _, name := range p.excluded {
		fmt.Println("EXCLUDE " + name)
	}
	names := make([]string, 0, len(p.renames))
	for name := range p.renames {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("RENAME %s -> %s\n", name, p.renames[name])
	}
	count := 0
	for _, c := range p.changes {
		for _, e := range c.edits {
			count++
			line := strings.Count(c.source[:e.start], "\n") + 1
			fmt.Printf("EDIT %s:%d byte %d: %q -> %q\n", c.name, line, e.start, e.before, e.after)
		}
	}
	fmt.Printf("TOTAL files=%d references=%d reference_files=%d\n", len(p.renames), count, len(p.changes))
}

func (p plan) apply(root string) error {
	// The complete plan and collision checks finish before the first write.
	for _, c := range p.changes {
		source := c.source
		for i := len(c.edits) - 1; i >= 0; i-- {
			e := c.edits[i]
			source = source[:e.start] + e.after + source[e.end:]
		}
		path := filepath.Join(root, c.name)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(source), info.Mode().Perm()); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(p.renames))
	for name := range p.renames {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := os.Rename(filepath.Join(root, name), filepath.Join(root, p.renames[name])); err != nil {
			return err
		}
	}
	return nil
}
