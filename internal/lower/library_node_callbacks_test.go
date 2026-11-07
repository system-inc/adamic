package lower

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Every registered callback which hands out a host object must judge the listener's view.
// In particular an opaque listener must not reach the dynamic-property path. The concrete
// parameter and nested-function direction are pinned below; Writable.on is the planted control.
func TestNodeRegisteredCallbacksJudgeHandedHostObjects(test *testing.T) {
	test.Parallel()
	registered := implementedNodeMembers
	l, resolved := nodeRegisteredDeclarations(test, registered)
	var opaque *checker.Type
	for _, statement := range l.program.Files()[0].Statements.Nodes {
		if statement.Kind == ast.KindVariableStatement {
			for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				if declaration.Name().Text() == "nodeCallbackOpaque" {
					opaque = l.checker.GetTypeAtLocation(declaration.Name())
				}
			}
		}
	}
	if opaque == nil {
		test.Fatal("opaque listener probe missing")
	}
	judged := 0
	for _, name := range slices.Sorted(maps.Keys(resolved)) {
		for _, signature := range append(
			l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(resolved[name]), checker.SignatureKindCall),
			l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(resolved[name]), checker.SignatureKindConstruct)...,
		) {
			for _, parameter := range signature.Parameters() {
				callback := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(parameter))
				if l.nodeHeldHostObjectInParameters(callback) == "" {
					continue
				}
				judged++
				if reached := l.nodeHostViewAt(opaque, callback); reached == "" {
					test.Errorf("%s hands out a host object without judging its listener parameters", name)
				}
			}
		}
	}
	test.Logf(
		"%d registered names, %d resolved, %d host-object callback signatures judged",
		len(registered),
		len(resolved),
		judged,
	)
}

func TestNodeHostCallbackParameterViews(test *testing.T) {
	test.Parallel()
	prefix := `import type {Stats} from 'node:fs';
 function hand(listener: (curr:Stats)=>void): void {}`
	for _, probe := range []struct{ name, parameter, want string }{
		{"honest", "Stats", ""},
		{"h11", "{ blocks:number }", "node:fs.StatsBase.blocks"},
		{"any", "any", "any"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			_, lowerSourceError := lowerSource(
				test,
				prefix+`function listener(curr: `+probe.parameter+`):void{} hand(listener);`,
			)
			if probe.want == "" {
				if lowerSourceError != nil {
					test.Fatal(lowerSourceError)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), probe.want) {
				test.Fatalf("want NotYet %q, got %v", probe.want, lowerSourceError)
			}
		})
	}
}

// These types are checker probes, not extra library registrations. Each function layer swaps
// direction: the host callback accepts a function supplied by the listener; its result hands
// Stats back, while that function's parameter travels from the listener to the host.
func TestNodeHostCallbackDirectionAndSignaturePositions(test *testing.T) {
	test.Parallel()
	probes := []struct {
		name, own, view string
		reject          bool
	}{
		{"parameter", "(x:{blocks:number})=>void", "(x:Stats)=>void", true},
		{"opaque-result-unchanged", "Stats", "unknown", false},
		{"opaque-nested-parameter", "(x:{nested:any})=>void", "(x:{nested:Stats})=>void", true},
		{"construct", "new (x:{blocks:number})=>object", "new (x:Stats)=>object", true},
		{"intersection", "((x:{blocks:number})=>void)&{tag:number}", "(x:Stats)=>void", true},
		{"depth", "(x:{blocks:number}[][][][][][][][][][])=>void", "(x:Stats[][][][][][][][][][])=>void", true},
		{"unknown", "(x:unknown)=>void", "(x:Stats)=>void", true},
		{"any", "(x:any)=>void", "(x:Stats)=>void", true},
		{"nested-result", "(f:()=>{blocks:number})=>void", "(f:()=>Stats)=>void", true},
		{"nested-parameter", "(f:(x:Stats)=>void)=>void", "(f:(x:{blocks:number})=>void)=>void", true},
		{"nested-honest", "(f:(x:Stats)=>Stats)=>void", "(f:(x:Stats)=>Stats)=>void", false},
		{"optional", "(x?:{blocks:number})=>void", "(x?:Stats)=>void", true},
		{"rest", "(...x:{blocks:number}[])=>void", "(...x:Stats[])=>void", true},
		{"tuple-rest", "(...x:[number,{blocks:number}])=>void", "(...x:[number,Stats])=>void", true},
		{"tuple-rest-honest", "(...x:[number,Stats])=>void", "(...x:[number,Stats])=>void", false},
		{"tuple-rest-empty", "(...x:[])=>void", "(x:Stats)=>void", false},
		{"ignored-parameter", "(x:number)=>void", "(x:number,stat:Stats)=>void", false},
		{"overload", "{(x:Stats):void;(x:{blocks:number}):void}", "(x:Stats)=>void", true},
		{"generic", "<T>(x:T)=>void", "(x:Stats)=>void", true},
		{"indexed-parameter", "<K extends 'stat'>(x:{stat:Stats}[K])=>void", "(x:Stats)=>void", true},
		{"indexed-handed", "(x:unknown)=>void", "<K extends 'stat'>(x:{stat:Stats}[K])=>void", true},
		{"indexed-nested", "<K extends 'stat'>(x:{nested:{stat:Stats}[K]})=>void", "(x:{nested:Stats})=>void", true},
		{"constrained-generic", "<T extends Stats>(x:T)=>void", "(x:Stats)=>void", true},
		{"handed-constraint", "(x:unknown)=>void", "<T extends Stats>(x:T)=>void", true},
		{"any-rest", "(...x:any[])=>void", "(x:Stats)=>void", true},
	}
	for _, probe := range probes {
		test.Run(probe.name, func(test *testing.T) {
			source := `import type {Stats} from 'node:fs'; declare const own: ` +
				probe.own +
				`; declare const view: ` +
				probe.view +
				`;`
			program, loadError := load.Load([]string{writeAgreementSource(test, "main.ts", source)})
			if loadError != nil {
				test.Fatal(loadError)
			}
			checked, release := program.Checker(context.Background(), program.Files()[0])
			defer release()
			l := &lowering{program: program, checker: checked}
			types := map[string]*checker.Type{}
			var visit ast.Visitor
			visit = func(n *ast.Node) bool {
				if n.Kind == ast.KindVariableDeclaration {
					types[n.Name().Text()] = checked.GetTypeAtLocation(n.Name())
				}
				n.ForEachChild(visit)
				return false
			}
			program.Files()[0].AsNode().ForEachChild(visit)
			if types["own"] == nil || types["view"] == nil {
				test.Fatal("probe types missing")
			}
			reached := l.nodeHostViewAt(types["own"], types["view"])
			if (reached != "") != probe.reject {
				test.Fatalf("reject=%v, got %q", probe.reject, reached)
			}
		})
	}
}

func (lowering *lowering) nodeHeldHostObjectInParameters(callback *checker.Type) string {
	for _, called := range append(
		lowering.checker.GetSignaturesOfType(callback, checker.SignatureKindCall),
		lowering.checker.GetSignaturesOfType(callback, checker.SignatureKindConstruct)...,
	) {
		for _, handed := range called.Parameters() {
			if host := lowering.nodeHeldHostObject(lowering.checker.GetTypeOfSymbol(handed), 0); host != nil {
				return lowering.checker.TypeToString(host)
			}
		}
	}
	return ""
}

// The check above can find what it looks for: Writable.on's "pipe" listener is handed a Readable.
func TestNodeRegisteredCallbacksFindAHandedHostObject(test *testing.T) {
	test.Parallel()
	l, resolved := nodeRegisteredDeclarations(test, map[string]bool{"node:stream.Writable.on": true})
	symbol, ok := resolved["node:stream.Writable.on"]
	if !ok {
		test.Fatal("node:stream.Writable.on resolves to no declaration")
	}
	if handed := l.nodeCallbackHostObject(symbol); !strings.Contains(handed, "Readable") {
		test.Fatalf("want Writable.on's pipe listener found handing a Readable, got %q", handed)
	}
}

// nodeRegisteredDeclarations loads a program that brings in the Node library files, and resolves
// each name in registered to the symbol whose declaration nodeLibraryName spells that way.
func nodeRegisteredDeclarations(test *testing.T, registered map[string]bool) (*lowering, map[string]*ast.Symbol) {
	test.Helper()
	source := strings.Join([]string{
		`import { readFileSync } from "node:fs";`,
		`import { cwd } from "node:process";`,
		`import { performance } from "node:perf_hooks";`,
		`import { createHash } from "node:crypto";`,
		`import { Buffer } from "node:buffer";`,
		`import { join } from "node:path";`,
		`import { tmpdir } from "node:os";`,
		`import { Writable } from "node:stream";`,
		`import { Socket } from "node:net";`,
		`declare const nodeCallbackOpaque: (...args: unknown[]) => void;`,
		`console.log(cwd() + join(tmpdir(), "a") + String(performance.now() >= 0) + String(Buffer.from("a").length) + createHash("sha256").update("a").digest("hex") + String(readFileSync === undefined) + String(Writable === undefined) + String(Socket === undefined));`,
	}, "\n")
	program, loadError := load.Load([]string{writeAgreementSource(test, "main.ts", source)})
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	test.Cleanup(release)
	l := &lowering{program: program, checker: checked}
	resolved := map[string]*ast.Symbol{}
	files := 0
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if !load.IsNodeLibrary(file) {
			continue
		}
		files++
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if nodeRegistrableKinds[node.Kind] && node.Name() != nil && node.Name().Kind == ast.KindIdentifier {
				name := node.Name()
				if symbol := checked.GetSymbolAtLocation(name); symbol != nil && registered[nodeLibraryName(symbol)] {
					resolved[nodeLibraryName(symbol)] = symbol
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	if files == 0 || len(resolved) == 0 {
		test.Fatalf(
			"read %d Node library files and resolved %d names: the walk no longer finds the registrations",
			files,
			len(resolved),
		)
	}
	return l, resolved
}

// nodeRegistrableKinds are the declarations a registered name can spell.
var nodeRegistrableKinds = map[ast.Kind]bool{
	ast.KindFunctionDeclaration: true, ast.KindVariableDeclaration: true, ast.KindClassDeclaration: true,
	ast.KindInterfaceDeclaration: true, ast.KindPropertySignature: true, ast.KindMethodSignature: true,
	ast.KindPropertyDeclaration: true, ast.KindMethodDeclaration: true,
}

// nodeCallbackHostObject is the first host object a member hands the program through a callback,
// as "a Readable to its listener", or "".
func (lowering *lowering) nodeCallbackHostObject(member *ast.Symbol) string {
	for _, signature := range lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeOfSymbol(member),
		checker.SignatureKindCall,
	) {
		for _, parameter := range signature.Parameters() {
			callback := lowering.checker.GetNonNullableType(lowering.checker.GetTypeOfSymbol(parameter))
			for _, called := range lowering.checker.GetSignaturesOfType(callback, checker.SignatureKindCall) {
				for _, handed := range called.Parameters() {
					if host := lowering.nodeHeldHostObject(lowering.checker.GetTypeOfSymbol(handed), 0); host != nil {
						return "a " + lowering.checker.TypeToString(host) + " to its " + parameter.Name
					}
				}
			}
		}
	}
	return ""
}

// nodeHeldHostObject is a host object a handed type may be: itself, a union's member, a generic's
// constraint (Writable.on's listener takes ...WritableEventMap[E]), or a tuple's or an array's
// element, as a rest parameter holds them.
func (lowering *lowering) nodeHeldHostObject(proven *checker.Type, depth int) *checker.Type {
	if proven == nil || depth > 8 {
		return nil
	}
	proven = lowering.checker.GetNonNullableType(proven)
	if _, host := lowering.nodeHostObject(proven); host {
		return proven
	}
	if proven.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess) != 0 {
		if constraint := lowering.checker.GetBaseConstraintOfType(proven); constraint != nil && constraint != proven {
			return lowering.nodeHeldHostObject(constraint, depth+1)
		}
		return nil
	}
	members := []*checker.Type{}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	} else if lowering.checker.IsArrayType(proven) || checker.IsTupleType(proven) {
		members = lowering.checker.GetTypeArguments(proven)
	}
	for _, member := range members {
		if host := lowering.nodeHeldHostObject(member, depth+1); host != nil {
			return host
		}
	}
	return nil
}
