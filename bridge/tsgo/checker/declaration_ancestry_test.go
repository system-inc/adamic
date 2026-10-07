package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestWave19DeclarationAncestry(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for name, text := range map[string]string{
		"input.a":       "setTimeout(()=>{},1);Promise.race([]);export {};",
		"timers.d.ts":   "export {};declare global{function setTimeout(callback:()=>void,delay:number):number;namespace setTimeout{const marker:number;}}",
		"tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a","timers.d.ts"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := p.Compiler.GetSourceFile(source)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	count, global := 0, false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && (node.Text() == "setTimeout" || node.Text() == "race") {
			count++
			out := &fields{}
			out.number(1)
			out.text("declaration-ancestry")
			wire, err := p.declarationAncestry(out, c, node, "declaration-ancestry")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(node)
			if symbol == nil || got[2] != "1" || got[4] != strconv.FormatUint(uint64(symbol.Flags), 10) || got[5] != strconv.Itoa(len(symbol.Declarations)) {
				t.Fatalf("symbol metadata %q", got)
			}
			at := 6
			for _, declaration := range symbol.Declarations {
				origin := ast.GetSourceFileOfNode(declaration)
				if got[at] != origin.FileName() {
					t.Fatal("wrong origin")
				}
				at += 4
				length, err := strconv.Atoi(got[at])
				if err != nil {
					t.Fatal(err)
				}
				at++
				observed := 0
				for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
					if got[at] != strings.TrimPrefix(ancestor.Kind.String(), "Kind") || got[at+2] != strconv.Itoa(int(ancestor.Pos())) || got[at+3] != strconv.Itoa(int(ancestor.End())) || got[at+4] != strconv.FormatUint(uint64(ancestor.Flags), 10) {
						t.Fatal("wrong raw ancestor")
					}
					expected := "0"
					if ancestor.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(ancestor) {
						expected = "1"
						global = true
					}
					if got[at+5] != expected {
						t.Fatal("wrong global-augmentation metadata")
					}
					at += 6
					observed++
				}
				if length != observed {
					t.Fatal("wrong ancestor count")
				}
			}
			if at != len(got) {
				t.Fatal("trailing fields")
			}
			for _, question := range []string{"declaration-ancestry\nextra", "different"} {
				if _, err := p.declarationAncestry(&fields{}, c, node, question); err == nil {
					t.Fatal("accepted malformed question")
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	if count != 2 || !global {
		t.Fatal("missing positive controls")
	}
	if _, err := p.declarationAncestry(&fields{}, c, file.AsNode(), "declaration-ancestry"); err == nil {
		t.Fatal("accepted wrong node kind")
	}
	t.Log("raw ancestry equals checker declarations; global augmentation, library member and refusal controls pass")
}

func TestWave19ResolvedCalleeAndProgramModules(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for name, text := range map[string]string{
		"input.a":         "import {value} from './dependency';export function helper(){return value;}helper();export {};",
		"dependency.d.ts": "export const value:number;",
		"tsconfig.json":   `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a","dependency.d.ts"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := p.Compiler.GetSourceFile(source)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	calls := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			calls++
			out := &fields{}
			out.number(1)
			out.text("wave19-resolved-callee")
			wire, err := p.wave19ResolvedCallee(out, c, node, "wave19-resolved-callee")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			signature := c.GetResolvedSignature(node)
			declaration := signature.Declaration()
			body := declaration.Body()
			if got[2] != "1" || got[3] != strconv.FormatUint(uint64(c.GetReturnTypeOfSignature(signature).Flags()), 10) || got[4] != "1" || got[5] != source || got[6] != "0" || got[7] != "1" || got[8] != "FunctionDeclaration" || got[9] != strconv.Itoa(int(declaration.Pos())) || got[10] != strconv.Itoa(int(declaration.End())) || got[12] != "1" || got[13] != strings.TrimPrefix(body.Kind.String(), "Kind") || got[14] != strconv.Itoa(int(body.Pos())) || got[15] != strconv.Itoa(int(body.End())) || got[16] != file.Text() {
				t.Fatalf("wrong resolved callee: %q", got)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	if calls != 1 {
		t.Fatal("missing callee positive")
	}
	out := &fields{}
	out.number(1)
	out.text("wave19-program-modules")
	wire, err := p.wave19ProgramModules(out, file.AsNode(), "wave19-program-modules")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if got[2] != strconv.Itoa(len(p.Compiler.SourceFiles())) {
		t.Fatal("wrong program file count")
	}
	at := 3
	found := false
	for range p.Compiler.SourceFiles() {
		path := got[at]
		length, err := strconv.Atoi(got[at+4])
		if err != nil {
			t.Fatal(err)
		}
		at += 5
		for item := 0; item < length; item++ {
			if path == source && got[at] == "ImportDeclaration" && got[at+3] == "./dependency" {
				if got[at+1] != "0" || got[at+2] != "StringLiteral" || got[at+4] != filepath.ToSlash(filepath.Join(directory, "dependency.d.ts")) {
					t.Fatal("wrong resolved module edge")
				}
				found = true
			}
			at += 5
		}
	}
	if !found || at != len(got) {
		t.Fatal("missing module-resolution positive")
	}
	if _, err := p.wave19ResolvedCallee(&fields{}, c, file.AsNode(), "wave19-resolved-callee"); err == nil {
		t.Fatal("wave19-resolved-callee accepted wrong kind")
	}
	if _, err := p.wave19ProgramModules(&fields{}, file.AsNode(), "wave19-program-modules\nextra"); err == nil {
		t.Fatal("wave19-program-modules accepted suffix")
	}
	t.Log("resolved signature declaration/body and program import targets agree with direct checker APIs")
}

func TestWave19AncestryAliasAndBindingNames(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for name, text := range map[string]string{
		"input.a":         "import { exit as imported } from './dependency';const {exit}= {exit(){}};exit();imported();",
		"dependency.d.ts": "export declare function exit():never;",
		"tsconfig.json":   `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["input.a","dependency.d.ts"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := p.Compiler.GetSourceFile(source)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	calls := 0
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			name := node.AsCallExpression().Expression
			out := &fields{}
			out.number(1)
			out.text("declaration-ancestry")
			wire, err := p.declarationAncestry(out, c, name, "declaration-ancestry\nalias")
			if err != nil {
				t.Fatal(err)
			}
			symbol := c.GetSymbolAtLocation(name)
			if symbol.Flags&ast.SymbolFlagsAlias != 0 {
				symbol = c.GetAliasedSymbol(symbol)
			}
			got := decodedFields(t, wire)
			if got[2] != "1" || got[3] != strconv.FormatUint(p.symbolID(symbol), 10) || got[4] != strconv.FormatUint(uint64(symbol.Flags), 10) {
				t.Fatal("wrong alias ancestry")
			}
			if name.Text() == "imported" && !strings.Contains(wire, filepath.Join(directory, "dependency.d.ts")) {
				t.Fatal("alias declaration not followed")
			}
			calls++
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(file.AsNode())
	if calls != 2 {
		t.Fatal("missing alias and binding controls")
	}
	t.Log("aliases match the checker and binding-pattern ancestor names do not panic")
}
