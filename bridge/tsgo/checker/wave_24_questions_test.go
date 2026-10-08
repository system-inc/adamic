package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWave24CheckerQuestions(t *testing.T) {
	if checker.TypeFlagsTypeParameter != 524288 || checker.TypeFlagsUnion != 134217728 || checker.TypeFlagsStringLiteral != 1024 {
		t.Fatal("pinned type flags changed")
	}
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	source := `export class C { f(): C { return this; } }
declare const promise: Promise<number>;
promise.catch((error: unknown) => {});
`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "globals.d.ts"), []byte("export {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePath(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var class, arrow, receiver *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindClassDeclaration {
			class = node
		}
		if node.Kind == ast.KindArrowFunction {
			arrow = node
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "promise" {
			receiver = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if class == nil || arrow == nil || receiver == nil {
		t.Fatal("missing positive node controls")
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.InspectWave24(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	identities := ask(class, "class-this-types")
	instanceID, err := strconv.ParseUint(identities[2], 10, 64)
	if err != nil || instanceID == 0 {
		t.Fatal("missing class identity", identities)
	}
	thisID, err := strconv.ParseUint(identities[3], 10, 64)
	if err != nil || thisID == 0 {
		t.Fatal("missing receiver identity", identities)
	}
	instance := c.GetTypeAtLocation(class)
	if p.typesByID[instanceID-1] != instance || p.typesByID[thisID-1] != checker.InterfaceType_thisType(instance.AsInterfaceType()) {
		t.Fatal("class identity disagrees with direct checker")
	}
	receiverShape := ask(receiver, "raw-shape")
	signatures := ask(receiver, "then-callback-signatures\n"+receiverShape[5])
	if len(signatures) < 4 || signatures[2] == "0" || signatures[3] == "0" {
		t.Fatal("missing Promise callback signatures", signatures)
	}
	parameters := ask(arrow, "handler-parameter-types")
	if strings.Join(parameters[2:], "|") != "1|1|1|2|0|0|0|0" {
		t.Fatal("unknown parameter facts", parameters)
	}
	for _, question := range []string{"class-this-types", "handler-parameter-types", "then-callback-signatures\n0", "then-callback-signatures\n01", "then-callback-signatures\n999999"} {
		if _, err := p.InspectWave24(file, uint64(receiver.Pos()), uint64(receiver.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted invalid question %q", question)
		}
	}
}
