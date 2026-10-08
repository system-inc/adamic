package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestWave05OutputFacts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "entry.a")
	config := filepath.Join(dir, "tsconfig.json")
	sources := map[string]string{
		"tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["entry.a","types.d.ts"]}`,
		"types.d.ts":    "declare namespace NodeJS {interface Process {exit():never;}} declare const process:NodeJS.Process;\n",
		"entry.a":       "import {f} from './helper';import type {T} from './types';declare const target:string;import(target);f();function d({shadowExit}:{shadowExit:()=>void}){shadowExit();}process.exit();export {};\n",
		"helper.ts":     "export async function f(){return 1;}\n",
	}
	for name, text := range sources {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	var call, member, binding *ast.Node
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression && n.Expression().Kind == ast.KindIdentifier && n.Expression().Text() == "f" {
			call = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "exit" {
			member = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "shadowExit" {
			binding = n
		}
		n.ForEachChild(visit)
		return false
	}
	visit(sf.AsNode())
	if call == nil || member == nil {
		t.Fatal("missing probes")
	}
	ask := func(n *ast.Node, q string) []string {
		wire, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), q)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	callee := ask(call, "output-callee")
	if len(callee) != 15 || callee[3] != filepath.Join(dir, "helper.ts") || callee[4] != "FunctionDeclaration" || callee[7] != "1" || callee[8] != "0" || callee[9] != "1" || callee[11] != "1" || callee[12] != "Block" {
		t.Fatalf("resolved callee metadata: %q", callee)
	}
	symbol := ask(member, "output-symbol")
	if !slices.Contains(symbol, "NodeJS") || !slices.Contains(symbol, "Process") || !slices.Contains(symbol, "Identifier") {
		t.Fatalf("declaration identity: %q", symbol)
	}
	if binding == nil {
		t.Fatal("missing destructuring probe")
	}
	if facts := ask(binding, "output-symbol"); !slices.Contains(facts, "ObjectBindingPattern") {
		t.Fatalf("binding ancestry: %q", facts)
	}
	modules := ask(sf.AsNode(), "runtime-modules")
	foundComputed, foundHelper := false, false
	for at := 3; at < len(modules); {
		name := modules[at]
		decl := modules[at+1]
		computed := modules[at+2]
		count := 0
		for _, r := range modules[at+3] {
			count = count*10 + int(r-'0')
		}
		if name == file {
			foundComputed = computed == "1" && decl == "0"
			edges := modules[at+4 : at+4+count]
			foundHelper = slices.Contains(edges, filepath.Join(dir, "helper.ts")) && !slices.Contains(edges, filepath.Join(dir, "types.d.ts"))
		}
		at += 4 + count
	}
	if !foundComputed || !foundHelper {
		t.Fatalf("runtime import edges: %q", modules)
	}
	for _, probe := range []struct {
		n *ast.Node
		q string
	}{{call, "output-callee\nextra"}, {member, "output-symbol\nextra"}, {sf.AsNode(), "runtime-modules\nextra"}} {
		if _, err := p.Inspect(file, uint64(probe.n.Pos()), uint64(probe.n.End()), strings.TrimPrefix(probe.n.Kind.String(), "Kind"), probe.q); err == nil {
			t.Fatalf("accepted suffix %q", probe.q)
		}
	}
	// Native consumes the pinned Go checker flag, not TypeScript's differently ordered enum.
	if checker.TypeFlagsNever != 262144 {
		t.Fatalf("never flag changed: %d", checker.TypeFlagsNever)
	}
}
