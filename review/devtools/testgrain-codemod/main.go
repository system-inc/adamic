// Command testgrain-codemod applies the mechanical part of wave 1 (six files). Domain-specific
// identity, union, planted-failure and shared-harness edits are finished by hand.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

type recipe struct {
	wrap    map[string]string // function -> result type (empty means void)
	remove  map[string]bool
	bodies  map[string]string
	timings map[string]bool
	units   map[string]bool
}

var recipes = map[string]recipe{
	"grain_rules_agree_shards_test.go": {
		wrap:    map[string]string{"rulesAgreeLowered": "string", "rulesAgreeNative": "string", "rulesAgreeOracle": "string", "rulesAgreeCapture": "string", "rulesAgreeCorpus": "*rulesAgreeProducts", "rulesAgreeSetup": "*rulesAgreeProducts"},
		timings: map[string]bool{"rulesAgreeSetup": true, "rulesAgreeRunShard": true},
		units:   map[string]bool{"rulesAgreeRunShard": true},
	},
	"quote_layout_shards_test.go": {
		wrap:    map[string]string{"quoteLayoutSetup": "quoteLayoutProducts", "quoteLayoutReady": "quoteLayoutProducts"},
		timings: map[string]bool{"quoteLayoutSetup": true, "quoteLayoutReady": true, "quoteLayoutNativeSetup": true, "quoteLayoutRunShard": true},
		units:   map[string]bool{"quoteLayoutRunShard": true},
		remove:  map[string]bool{"quoteLayoutContextCommand": true},
		bodies:  map[string]string{"quoteLayoutCommand": `t.Helper(); return testgrain.Command(t,name,arguments...)`, "quoteLayoutSetupCommand": `t.Helper(); return testgrain.Command(t,name,arguments...)`},
	},
	"node_table_split_test.go": {
		wrap:    map[string]string{"nodeTableSetup": ""},
		timings: map[string]bool{"nodeTableSetup": true, "nodeTableRunShard": true},
		units:   map[string]bool{"nodeTableRunShard": true},
	},
	"text_shards_test.go": {
		wrap:    map[string]string{"textReadySetup": "textSetupState"},
		remove:  map[string]bool{"textPhaseDeadline": true},
		timings: map[string]bool{"textReadySetup": true},
		bodies: map[string]string{
			"textDeadline": `testgrain.Unit(t); return func() {}`,
			"textBounded":  `t.Helper(); return testgrain.CommandContext(t, t.Context(), name, arguments...)`,
		},
	},
	"list_layout_shards_test.go": {
		wrap:    map[string]string{"buildListLayoutSetup": "", "listLayoutSetup": "listLayoutProducts"},
		remove:  map[string]bool{"listLayoutDeadline": true},
		timings: map[string]bool{"buildListLayoutSetup": true, "runListLayoutShard": true},
		units:   map[string]bool{"runListLayoutShard": true},
		bodies:  map[string]string{"listLayoutCommand": `t.Helper(); return testgrain.CommandContext(t, t.Context(), name, arguments...)`},
	},
	"compiler_expressions_shards_test.go": {
		timings: map[string]bool{"compilerExpressionsSetup": true, "compilerExpressionsShard": true},
		units:   map[string]bool{"compilerExpressionsShard": true},
	},
}

// Replacement syntax has no source positions, so printer uses the edited file's
// positions/comments rather than positions from a second parse.
func clearPositions(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if !v.IsNil() {
			clearPositions(v.Elem())
		}
		return
	}
	if v.Type() == reflect.TypeOf(token.Pos(0)) && v.CanSet() {
		v.SetInt(0)
		return
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == "Ellipsis" && v.Field(i).Type() == reflect.TypeOf(token.Pos(0)) && v.Field(i).Int() != 0 {
				v.Field(i).SetInt(1)
			} else {
				clearPositions(v.Field(i))
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			clearPositions(v.Index(i))
		}
	}
}
func parseFunction(source string) *ast.FuncDecl {
	f, err := parser.ParseFile(token.NewFileSet(), "replacement.go", "package replacement\n"+source, parser.SkipObjectResolution)
	if err != nil {
		panic(err)
	}
	decl := f.Decls[0].(*ast.FuncDecl)
	clearPositions(reflect.ValueOf(decl))
	return decl
}
func syntax(n ast.Node) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, token.NewFileSet(), n)
	return b.String()
}
func unit() ast.Stmt { return parseFunction("func replacement(){testgrain.Unit(t)}").Body.List[0] }
func strip(b *ast.BlockStmt, addUnit bool) {
	out := make([]ast.Stmt, 0, len(b.List))
	for _, stmt := range b.List {
		text := syntax(stmt)
		// The listed functions use these bindings exclusively for grain timing.
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			clock := false
			for _, rhs := range assign.Rhs {
				value := syntax(rhs)
				clock = clock || strings.Contains(value, "time.Now()") || strings.Contains(value, "time.Since(started)")
			}
			if clock {
				continue
			}
			if strings.Contains(text, "time.AfterFunc(") || strings.Contains(text, "listLayoutDeadline(") || strings.Contains(text, "textPhaseDeadline(") {
				if addUnit {
					out = append(out, unit())
				}
				continue
			}
		}
		if deferStmt, ok := stmt.(*ast.DeferStmt); ok {
			call := syntax(deferStmt.Call)
			if call == "timer.Stop()" || call == "deadline.Stop()" || call == "finish()" || call == "stop()" || strings.Contains(call, "time.Since(started)") {
				continue
			}
		}
		if _, ok := stmt.(*ast.IfStmt); ok && strings.Contains(text, "cooked:") {
			continue
		}
		if expr, ok := stmt.(*ast.ExprStmt); ok && strings.HasPrefix(syntax(expr.X), "t.Logf(") && (strings.Contains(text, "cooked") || strings.Contains(text, "time.Since(started)") || strings.Contains(text, "time.Since(setupStarted)")) {
			continue
		}
		ast.Inspect(stmt, func(n ast.Node) bool {
			if block, ok := n.(*ast.BlockStmt); ok {
				strip(block, addUnit)
				return false
			}
			return true
		})
		out = append(out, stmt)
	}
	b.List = out
}
func importGrain(f *ast.File) {
	for _, imp := range f.Imports {
		if imp.Path.Value == `"github.com/system-inc/adamic/internal/testgrain"` {
			return
		}
	}
	imp := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/system-inc/adamic/internal/testgrain"`}}
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.IMPORT {
			d.Specs = append(d.Specs, imp)
			f.Imports = append(f.Imports, imp)
			return
		}
	}
	panic("no import declaration")
}
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./review/devtools/testgrain-codemod <wave-1 files>")
		os.Exit(2)
	}
	for _, file := range os.Args[1:] {
		r, ok := recipes[path.Base(file)]
		if !ok {
			panic("file is outside wave 1: " + file)
		}
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, file, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			panic(err)
		}
		for _, imp := range f.Imports {
			if imp.Path.Value == `"github.com/system-inc/adamic/internal/testgrain"` {
				panic("already converted: " + file)
			}
		}
		comments := ast.NewCommentMap(fs, f, f.Comments)
		declarations := make([]ast.Decl, 0, len(f.Decls))
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				declarations = append(declarations, decl)
				continue
			}
			name := fn.Name.Name
			if r.remove[name] {
				fmt.Printf("%s: remove %s\n", file, name)
				continue
			}
			if body, ok := r.bodies[name]; ok {
				fn.Body = parseFunction("func replacement(){" + body + "}").Body
				fmt.Printf("%s: delegate %s to testgrain\n", file, name)
			}
			if r.timings[name] {
				strip(fn.Body, r.units[name])
				fmt.Printf("%s: remove timers/verdicts from %s\n", file, name)
			}
			if result, ok := r.wrap[name]; ok {
				// Keep recipe bodies intact; hand finishing removes the now-redundant once.
				fn.Name.Name = name + "Prepare"
				body := "testgrain.Setup(t, " + strconv.Quote("wave1/"+name) + ", func()(bool,error){" + fn.Name.Name + "(t); return true,nil})"
				results := ""
				if result != "" {
					results = " " + result
					body = "return testgrain.Setup(t, " + strconv.Quote("wave1/"+name) + ", func()(" + result + ",error){return " + fn.Name.Name + "(t),nil})"
				}
				wrapper := parseFunction("func " + name + "(t *testing.T)" + results + " {t.Helper(); " + body + "}")
				wrapper.Type.Func = fn.Type.Func
				wrapper.Name.NamePos = fn.Name.NamePos
				declarations = append(declarations, wrapper)
				fmt.Printf("%s: wrap %s in Setup\n", file, name)
			}
			declarations = append(declarations, fn)
		}
		f.Decls = declarations
		importGrain(f)
		f.Comments = comments.Filter(f).Comments()
		var output bytes.Buffer
		if err := printer.Fprint(&output, fs, f); err != nil {
			panic(err)
		}
		if err := os.WriteFile(file, output.Bytes(), 0644); err != nil {
			panic(err)
		}
	}
}
