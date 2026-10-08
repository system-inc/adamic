package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWave07FlowAndModuleFacts(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	helper := filepath.Join(directory, "helper.ts") // TypeScript oracle input, not an Adamic implementation.
	for path, text := range map[string]string{
		config:                                   `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["ambient.d.ts"]}`,
		filepath.Join(directory, "ambient.d.ts"): `declare const flag:boolean;`,
		helper:                                   `export function helper(){return 1;}`,
		file:                                     `import {helper} from "./helper.js";if(flag){helper();}else{helper();}`,
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	wire, err := p.wave07ControlFlow(source.AsNode(), "wave07-control-flow")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[0] != "1" || fields[1] != "wave07-control-flow" {
		t.Fatal(fields)
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil || count < 3 {
		t.Fatalf("missing branch blocks: %v", fields)
	}
	if !strings.Contains(wire, "CallExpression") {
		t.Fatal("call syntax events absent")
	}
	if _, err := p.wave07ControlFlow(source.Statements.Nodes[0], "wave07-control-flow"); err == nil {
		t.Fatal("non-root accepted")
	}
	if _, err := p.wave07ControlFlow(source.AsNode(), "wave07-control-flow\nextra"); err == nil {
		t.Fatal("suffix accepted")
	}
	wire, err = p.wave07ProgramModules(source.AsNode(), "wave07-program-modules")
	if err != nil {
		t.Fatal(err)
	}
	fields = decodedFields(t, wire)
	if fields[0] != "1" || fields[1] != "wave07-program-modules" {
		t.Fatal(fields)
	}
	if !strings.Contains(wire, helper) {
		t.Fatal("resolved target absent")
	}
	if _, err := p.wave07ProgramModules(source.Statements.Nodes[0], "wave07-program-modules"); err == nil {
		t.Fatal("non-source accepted")
	}
	if _, err := p.wave07ProgramModules(source.AsNode(), "wave07-program-modules\nextra"); err == nil {
		t.Fatal("suffix accepted")
	}
	// No verdict, write, exit, blocking, or finding tags are introduced by either question.
	for _, kind := range []string{"Identifier", "IfStatement", "CallExpression"} {
		var found bool
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if strings.TrimPrefix(node.Kind.String(), "Kind") == kind {
				found = true
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(source.AsNode())
		if !found {
			t.Fatal("fixture lacks " + kind)
		}
	}
}
