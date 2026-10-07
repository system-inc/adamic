package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestProcessFactsLocationsAndGuards(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	helper := filepath.Join(directory, "helper.ts")
	helperSource := filepath.Join(directory, "helper.a")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{config: `{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ESNext","module":"ESNext","moduleResolution":"Bundler","lib":["ESNext"]}}`, helperSource: `export function write(){return '世界🌍'}`, file: `import {write} from './helper';function f(){for(;write();){break}}write();const {value}= {value:1};class C {['computed'](){} }export {};`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Base(helperSource), helper); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	nodes := map[ast.Kind]*ast.Node{}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		nodes[n.Kind] = n
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	for _, row := range []struct {
		mode string
		kind ast.Kind
	}{{"process-node-fields", ast.KindForStatement}, {"resolved-call-target", ast.KindCallExpression}, {"symbol-declaration-paths", ast.KindIdentifier}, {"program-module-resolution", ast.KindSourceFile}, {"source-parse-context", ast.KindSourceFile}} {
		n := nodes[row.kind]
		if n == nil {
			t.Fatalf("missing %s", row.kind)
		}
		if row.kind == ast.KindSourceFile {
			n = source.AsNode()
		}
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), row.mode)
		if err != nil {
			t.Fatal(err)
		}
		if !utf8.ValidString(wire) {
			t.Fatalf("invalid UTF8 %s", row.mode)
		}
		fields := decodedFields(t, wire)
		if len(fields) < 3 || fields[0] != "1" || fields[1] != row.mode {
			t.Fatalf("bad %s header: %q", row.mode, fields)
		}
		if row.mode == "process-node-fields" {
			if !strings.Contains(wire, "CallExpression") || !strings.Contains(wire, "Block") {
				t.Fatal("missing named for slots")
			}
		}
		if row.mode == "resolved-call-target" {
			if !strings.Contains(wire, helper) || !strings.Contains(wire, "世界🌍") {
				t.Fatal("missing resolved source")
			}
		}
		if row.mode == "program-module-resolution" {
			if !strings.Contains(wire, helper) || !strings.Contains(wire, "StringLiteral") {
				t.Fatal("missing module target")
			}
		}
		if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), row.mode+"\nextra"); err == nil {
			t.Fatalf("%s accepted suffix", row.mode)
		}
	}
	for _, kind := range []ast.Kind{ast.KindBindingElement, ast.KindMethodDeclaration} {
		n := nodes[kind]
		if n == nil {
			t.Fatalf("missing %s", kind)
		}
		if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), "symbol-declaration-paths"); err != nil {
			t.Fatal(err)
		}
	}

	for _, mode := range []string{"resolved-call-target", "program-module-resolution", "source-parse-context"} {
		n := nodes[ast.KindIdentifier]
		if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", mode); err == nil {
			t.Fatalf("%s accepted wrong node kind", mode)
		}
	}
}
