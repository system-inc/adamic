// Package skipcensus makes skipped verification an explicit gate failure.
package skipcensus

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	File      string   `json:"file"`
	Test      string   `json:"test"`
	ID        string   `json:"id"`
	Line      int      `json:"line"`
	Condition string   `json:"condition"`
	Reads     []string `json:"reads"`
	Callers   []string `json:"callers"`
	Message   string   `json:"message"`
	Class     string   `json:"class"`
	Provides  string   `json:"provides"`
}

type function struct {
	name, file   string
	calls, reads []string
}

func printed(set *token.FileSet, node ast.Node) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, set, node)
	return b.String()
}

// testingReceiver follows source declarations without loading platform-specific
// packages. It also recognizes aliases of testing parameters and local variables.
func testingReceiver(set *token.FileSet, expression ast.Expr, alias string, seen map[*ast.Object]bool) bool {
	id, ok := expression.(*ast.Ident)
	if !ok || id.Obj == nil || seen[id.Obj] {
		return false
	}
	seen[id.Obj] = true
	validType := func(typ ast.Expr) bool {
		if typ == nil {
			return false
		}
		text := printed(set, typ)
		prefix := alias + "."
		if alias == "." {
			prefix = ""
		}
		return text == "*"+prefix+"T" || text == "*"+prefix+"B" || text == prefix+"TB"
	}
	switch declaration := id.Obj.Decl.(type) {
	case *ast.Field:
		return validType(declaration.Type)
	case *ast.ValueSpec:
		if validType(declaration.Type) {
			return true
		}
		for index, name := range declaration.Names {
			if name.Name == id.Name && index < len(declaration.Values) {
				return testingReceiver(set, declaration.Values[index], alias, seen)
			}
		}
	case *ast.AssignStmt:
		for index, name := range declaration.Lhs {
			if declared, ok := name.(*ast.Ident); ok && declared.Name == id.Name && index < len(declaration.Rhs) {
				return testingReceiver(set, declaration.Rhs[index], alias, seen)
			}
		}
	}
	return false
}

// Scan parses all Adamic-owned Go tests, including overlays in testdata. The
// separately pinned cohere submodule is a different repository, outside this gate.
// Build constraints are deliberately ignored: platform skips must be declared too.
func Scan(root string) ([]Row, error) {
	set := token.NewFileSet()
	functions := map[string]*function{}
	var rows []Row
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "cohere" && path == filepath.Join(root, "cohere") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		tree, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		testingAlias := "testing"
		for _, imported := range tree.Imports {
			name, _ := strconv.Unquote(imported.Path.Value)
			if name == "testing" && imported.Name != nil {
				testingAlias = imported.Name.Name
			}
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		for _, decl := range tree.Decls {
			f, ok := decl.(*ast.FuncDecl)
			if !ok || f.Body == nil {
				continue
			}
			key := filepath.Dir(relative) + ":" + f.Name.Name
			fn := &function{name: f.Name.Name, file: relative}
			functions[key] = fn
			ast.Inspect(f.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if id, ok := call.Fun.(*ast.Ident); ok {
					fn.calls = append(fn.calls, id.Name)
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					switch sel.Sel.Name {
					case "Getenv", "LookupEnv", "ReadFile", "Open", "Stat", "Lstat", "WalkDir", "LookPath":
						fn.reads = append(fn.reads, printed(set, call))
					}
				}
				return true
			})
			var walk func(ast.Node, []string)
			walk = func(n ast.Node, guards []string) {
				if n == nil {
					return
				}
				if branch, ok := n.(*ast.IfStmt); ok {
					condition := printed(set, branch.Cond)
					if branch.Init != nil {
						condition = printed(set, branch.Init) + "; " + condition
					}
					walk(branch.Init, guards)
					walk(branch.Body, append(append([]string{}, guards...), condition))
					walk(branch.Else, append(append([]string{}, guards...), "!("+condition+")"))
					return
				}
				if call, ok := n.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && (sel.Sel.Name == "Skip" || sel.Sel.Name == "Skipf" || sel.Sel.Name == "SkipNow") {
						// Recognize testing receivers by their declared type, including arbitrary names.
						isTesting := testingReceiver(set, sel.X, testingAlias, map[*ast.Object]bool{})
						if isTesting {
							condition := strings.Join(guards, " && ")
							if condition == "" {
								condition = "true"
							}
							sum := sha256.Sum256([]byte(condition))
							message := ""
							if len(call.Args) > 0 {
								message = printed(set, call.Args[0])
							}
							rows = append(rows, Row{File: relative, Test: f.Name.Name, ID: fmt.Sprintf("%s:%x", f.Name.Name, sum[:]), Line: set.Position(call.Pos()).Line, Condition: condition, Message: message})
						}
					}
				}
				ast.Inspect(n, func(child ast.Node) bool {
					if child == nil {
						return false
					}
					if child == n {
						return true
					}
					walk(child, guards)
					return false
				})
			}
			walk(f.Body, nil)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for i := range rows {
		row := &rows[i]
		directory := filepath.Dir(row.File)
		seen := map[string]bool{}
		var reads []string
		var collect func(string)
		collect = func(name string) {
			if seen[name] {
				return
			}
			seen[name] = true
			if f := functions[directory+":"+name]; f != nil {
				reads = append(reads, f.reads...)
				for _, call := range f.calls {
					collect(call)
				}
			}
		}
		collect(row.Test)
		row.Reads = unique(reads)
		for key, f := range functions {
			if filepath.Dir(f.file) != directory || !strings.HasPrefix(f.name, "Test") && !strings.HasPrefix(f.name, "Benchmark") {
				continue
			}
			seen = map[string]bool{}
			var reaches func(string) bool
			reaches = func(name string) bool {
				if name == row.Test {
					return true
				}
				if seen[name] {
					return false
				}
				seen[name] = true
				if callee := functions[directory+":"+name]; callee != nil {
					for _, call := range callee.calls {
						if reaches(call) {
							return true
						}
					}
				}
				return false
			}
			if reaches(f.name) {
				row.Callers = append(row.Callers, strings.Split(key, ":")[1])
			}
		}
		row.Callers = unique(row.Callers)
		if len(row.Callers) > 0 {
			sum := sha256.Sum256([]byte(row.Condition))
			row.ID = fmt.Sprintf("%s:%x", strings.Join(row.Callers, ","), sum[:])
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].File != rows[j].File {
			return rows[i].File < rows[j].File
		}
		return rows[i].Line < rows[j].Line
	})
	return rows, nil
}
func unique(values []string) []string {
	sort.Strings(values)
	result := []string{}
	for _, v := range values {
		if len(result) == 0 || v != result[len(result)-1] {
			result = append(result, v)
		}
	}
	return result
}
func Load(r io.Reader) ([]Row, error) {
	var rows []Row
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	err := d.Decode(&rows)
	return rows, err
}
func key(row Row) string { return row.File + ":" + row.ID }

// Validate ignores line numbers. Moving a skip does not change its identity.
func Validate(actual, declared []Row) error {
	remaining := map[string]Row{}
	for _, r := range declared {
		k := key(r)
		if _, ok := remaining[k]; ok {
			return fmt.Errorf("duplicate declaration %s", k)
		}
		switch r.Class {
		case "required-input", "measurement", "not-applicable", "opt-in-lane":
		default:
			return fmt.Errorf("invalid class for %s", k)
		}
		if r.Provides == "" {
			return fmt.Errorf("missing provision or rationale for %s", k)
		}
		remaining[k] = r
	}
	var errors []string
	for _, r := range actual {
		k := key(r)
		d, ok := remaining[k]
		if !ok {
			errors = append(errors, fmt.Sprintf("undeclared skip %s (%s:%d)", k, r.Test, r.Line))
			continue
		}
		delete(remaining, k)
		a, b := r, d
		a.Line = 0
		b.Line = 0
		a.Class = ""
		b.Class = ""
		a.Provides = ""
		b.Provides = ""
		aj, _ := json.Marshal(a)
		bj, _ := json.Marshal(b)
		if !bytes.Equal(aj, bj) {
			errors = append(errors, "stale metadata "+k)
		}
	}
	for k := range remaining {
		errors = append(errors, "removed skip "+k)
	}
	sort.Strings(errors)
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return nil
}
