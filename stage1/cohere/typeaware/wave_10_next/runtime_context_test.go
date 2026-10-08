package wave10next_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
	adapter "github.com/system-inc/adamic/stage1/cohere/typeaware/wave_10_next"
)

func decode(t *testing.T, wire string) []string {
	t.Helper()
	var result []string
	for wire != "" {
		colon := strings.IndexByte(wire, '\n')
		if colon < 1 {
			t.Fatal("missing frame length")
		}
		count, err := strconv.Atoi(wire[:colon])
		if err != nil {
			t.Fatal(err)
		}
		rest := wire[colon+1:]
		bytes, units := 0, 0
		for _, r := range rest {
			if units == count {
				break
			}
			units += len(utf16.Encode([]rune{r}))
			bytes += len(string(r))
		}
		if units != count {
			t.Fatal("invalid frame extent")
		}
		result = append(result, rest[:bytes])
		wire = rest[bytes:]
	}
	return result
}

func TestRuntimeContextFactsAndLiveDispatch(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	decl := filepath.Join(directory, "node.d.ts")
	source := `declare const work:Promise<string>;Promise.race([work,new Promise((_r,reject)=>setTimeout(reject,10))]);process.exit(0);`
	declarations := `export {};declare global {namespace NodeJS {interface Process {exit(code?:number):never}}var process:NodeJS.Process;function setTimeout(callback:(...args:any[])=>void,delay?:number):number;}`
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["node.d.ts"]}`, file: source, decl: declarations} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := bridge.Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	loaded := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file)))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), loaded)
	defer release()
	nodes := map[string]*ast.Node{}
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			nodes[n.Text()] = n
		}
		n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
	}
	visit(loaded.AsNode())
	for _, name := range []string{"race", "Promise", "setTimeout", "exit"} {
		node := nodes[name]
		wire, err := adapter.Inspect(p.Compiler, loaded, node, c, "runtime-context\norigin")
		if err != nil {
			t.Fatal(err)
		}
		values := decode(t, wire)
		if len(values) < 7 || values[0] != "1" || values[1] != "runtime-context" || values[2] != "origin" || values[3] != "1" || values[5] != name {
			t.Fatalf("missing origin %s: %q", name, values)
		}
		if len(values) < 24 {
			t.Fatalf("short declaration: %q", values)
		}
		if name == "race" && (values[12] != "1" || values[18] != "InterfaceDeclaration" || values[19] != "PromiseConstructor") {
			t.Fatal("library owner absent")
		}
		if name == "exit" && (len(values) < 39 || values[11] != "1" || values[18] != "InterfaceDeclaration" || values[19] != "Process" || values[32] != "ModuleDeclaration" || values[33] != "NodeJS" || values[38] != "Identifier") {
			t.Fatal("namespace ancestry absent")
		}
		if name == "setTimeout" && (len(values) < 32 || values[13] != "1" || values[18] != "ModuleBlock" || values[26] != "global" || values[30] != "1") {
			t.Fatal("global augmentation ancestry absent")
		}
		if live, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "runtime-context\norigin"); err != nil || live != wire {
			t.Fatalf("live dispatch differs: %v", err)
		}
	}
	wire, err := adapter.Inspect(p.Compiler, loaded, loaded.AsNode(), c, "runtime-context\nprogram")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(wire, "node.d.ts") {
		t.Fatal("program declaration roots absent")
	}
	if _, err := adapter.Inspect(p.Compiler, loaded, nodes["race"], c, "runtime-context\norigin\nextra"); err == nil {
		t.Fatal("suffix accepted")
	}
	if _, err := adapter.Inspect(p.Compiler, loaded, nil, c, "runtime-context\norigin"); err == nil {
		t.Fatal("missing anchor accepted")
	}
	call := nodes["exit"].Parent.Parent
	signatureWire, err := adapter.Inspect(p.Compiler, loaded, call, c, "runtime-context\nsignature")
	if err != nil {
		t.Fatal(err)
	}
	signature := decode(t, signatureWire)
	if signature[len(signature)-1] != "262144" {
		t.Fatal("resolved never return flag absent")
	}
	t.Log("raw origin/program facts and live dispatch pass")
}

func TestSyntaxMembershipAndDestructuringAncestry(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "syntax.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `declare function use(x?:unknown):void; const {a=1}={}; class C{static{use(a)}}for(;;){break;}try{use()}catch({code=0}){use(code)}export {};`
	for path, text := range map[string]string{config: `{"compilerOptions":{"target":"ES2022","types":[]},"files":["stub.d.ts"]}`, filepath.Join(directory, "stub.d.ts"): "export {};", file: source} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := bridge.Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	loaded := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file)))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), loaded)
	defer release()
	staticBody := false
	bareFor := false
	ancestry := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		wire, err := adapter.Inspect(p.Compiler, loaded, node, c, "runtime-context\nsyntax")
		if err != nil {
			t.Fatal(err)
		}
		values := decode(t, wire)
		if len(values) < 8 || values[2] != "syntax" {
			t.Fatalf("invalid syntax fields: %q", values)
		}
		count, err := strconv.Atoi(values[7])
		if err != nil {
			t.Fatal(err)
		}
		roles := map[string][]string{}
		offset := 8
		for i := 0; i < count; i++ {
			name := values[offset]
			length, err := strconv.Atoi(values[offset+1])
			if err != nil {
				t.Fatal(err)
			}
			offset += 2
			roles[name] = values[offset : offset+length*3]
			offset += length * 3
		}
		if offset != len(values) {
			t.Fatal("syntax trailing fields")
		}
		if node.Kind == ast.KindClassStaticBlockDeclaration {
			body := roles["body"]
			if len(body) != 3 || body[0] != "Block" {
				t.Fatal("static block body absent")
			}
			staticBody = true
		}
		if node.Kind == ast.KindForStatement {
			if len(roles["initializer"])+len(roles["condition"])+len(roles["incrementor"]) != 0 {
				t.Fatal("omitted loop field invented")
			}
			bareFor = true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "code" {
			wire, err := adapter.Inspect(p.Compiler, loaded, node, c, "runtime-context\norigin")
			if err != nil {
				t.Fatal(err)
			}
			decode(t, wire)
			ancestry = true
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(loaded.AsNode())
	if !staticBody || !bareFor || !ancestry {
		t.Fatal("syntax fixture did not exercise required fields")
	}
}
