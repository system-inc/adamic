package lower

import (
	"context"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// nodeHostViewReach follows a method's result and not its parameters, because no registered member
// hands the program a host object through a callback (#8k12t4s). That's a claim about the
// registrations, so this holds it: every name RegisterNodeLibraryMembers filled is resolved to its
// declaration in the Node library files, and for each call signature, each parameter that's callable
// is read for a callback parameter that's a host object per nodeHostObject. One that is fails here
// before a view of it reaches native. ADAMIC_HOST_CALLBACK_MUTANT=1 registers Writable.on, whose
// "pipe" listener is handed the Readable piped in, and the test must go red.
func TestNodeRegisteredCallbacksHandNoHostObject(t *testing.T) {
	t.Parallel()
	registered := maps.Clone(implementedNodeMembers)
	if os.Getenv("ADAMIC_HOST_CALLBACK_MUTANT") == "1" {
		registered["node:stream.Writable.on"] = true
	}
	l, resolved := nodeRegisteredDeclarations(t, registered)
	for _, name := range slices.Sorted(maps.Keys(resolved)) {
		if handed := l.nodeCallbackHostObject(resolved[name]); handed != "" {
			t.Errorf("%s hands the program %s through a callback: follow parameters contravariantly in nodeHostViewReach, or keep it out", name, handed)
		}
	}
	unresolved := 0
	for name := range registered {
		if _, ok := resolved[name]; !ok {
			unresolved++
		}
	}
	t.Logf("%d registered names, %d resolved to a Node library declaration and read, %d name no declaration (a module or a value the walk can't see through)", len(registered), len(resolved), unresolved)
}

// The check above can find what it looks for: Writable.on's "pipe" listener is handed a Readable.
func TestNodeRegisteredCallbacksFindAHandedHostObject(t *testing.T) {
	t.Parallel()
	l, resolved := nodeRegisteredDeclarations(t, map[string]bool{"node:stream.Writable.on": true})
	symbol, ok := resolved["node:stream.Writable.on"]
	if !ok {
		t.Fatal("node:stream.Writable.on resolves to no declaration")
	}
	if handed := l.nodeCallbackHostObject(symbol); !strings.Contains(handed, "Readable") {
		t.Fatalf("want Writable.on's pipe listener found handing a Readable, got %q", handed)
	}
}

// nodeRegisteredDeclarations loads a program that brings in the Node library files, and resolves
// each name in registered to the symbol whose declaration nodeLibraryName spells that way.
func nodeRegisteredDeclarations(t *testing.T, registered map[string]bool) (*lowering, map[string]*ast.Symbol) {
	t.Helper()
	source := strings.Join([]string{
		`import { readFileSync } from "node:fs";`, `import { cwd } from "node:process";`, `import { performance } from "node:perf_hooks";`,
		`import { createHash } from "node:crypto";`, `import { Buffer } from "node:buffer";`, `import { join } from "node:path";`,
		`import { tmpdir } from "node:os";`, `import { Writable } from "node:stream";`, `import { Socket } from "node:net";`,
		`console.log(cwd() + join(tmpdir(), "a") + String(performance.now() >= 0) + String(Buffer.from("a").length) + createHash("sha256").update("a").digest("hex") + String(readFileSync === undefined) + String(Writable === undefined) + String(Socket === undefined));`,
	}, "\n")
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
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
		t.Fatalf("read %d Node library files and resolved %d names: the walk no longer finds the registrations", files, len(resolved))
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
func (l *lowering) nodeCallbackHostObject(member *ast.Symbol) string {
	for _, signature := range l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(member), checker.SignatureKindCall) {
		for _, parameter := range signature.Parameters() {
			callback := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(parameter))
			for _, called := range l.checker.GetSignaturesOfType(callback, checker.SignatureKindCall) {
				for _, handed := range called.Parameters() {
					if host := l.nodeHeldHostObject(l.checker.GetTypeOfSymbol(handed), 0); host != nil {
						return "a " + l.checker.TypeToString(host) + " to its " + parameter.Name
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
func (l *lowering) nodeHeldHostObject(proven *checker.Type, depth int) *checker.Type {
	if proven == nil || depth > 8 {
		return nil
	}
	proven = l.checker.GetNonNullableType(proven)
	if _, host := l.nodeHostObject(proven); host {
		return proven
	}
	if proven.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess) != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil && constraint != proven {
			return l.nodeHeldHostObject(constraint, depth+1)
		}
		return nil
	}
	members := []*checker.Type{}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	} else if l.checker.IsArrayType(proven) || checker.IsTupleType(proven) {
		members = l.checker.GetTypeArguments(proven)
	}
	for _, member := range members {
		if host := l.nodeHeldHostObject(member, depth+1); host != nil {
			return host
		}
	}
	return nil
}
