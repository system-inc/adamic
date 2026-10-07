package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	checkerapi "github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestTypeArgumentsQuestion(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := `interface Item {value:string} declare const promise:Promise<Item[]>;declare const readonly:Readonly<Item>;promise;readonly;`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Parent.Kind == ast.KindExpressionStatement {
			nodes = append(nodes, node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if len(nodes) != 2 {
		t.Fatal("missing controls")
	}
	for _, node := range nodes {
		ask := func(question string) []string {
			t.Helper()
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question)
			if err != nil {
				t.Fatal(err)
			}
			return decodedFields(t, wire)
		}
		shape := ask("raw-shape")
		id := shape[5]
		fields := ask("type-arguments\n" + id)
		subject := c.GetTypeAtLocation(node)
		symbol := ""
		if subject.Symbol() != nil {
			symbol = subject.Symbol().Name
		}
		if fields[2] != strings.ToValidUTF8(symbol, "�") {
			t.Fatalf("symbol: %q", fields)
		}
		var arguments []*checkerapi.Type
		if subject.ObjectFlags()&checkerapi.ObjectFlagsReference != 0 {
			arguments = checkerapi.Checker_getTypeArguments(c, subject)
		}
		count, err := strconv.Atoi(fields[4])
		if err != nil || count != len(arguments) {
			t.Fatalf("arguments: %q", fields)
		}
		for i, argument := range arguments {
			argumentID, _ := strconv.ParseUint(fields[5+i], 10, 64)
			if argumentID == 0 || p.typesByID[argumentID-1] != argument {
				t.Fatal("wrong reference argument")
			}
		}
		cursor := 5 + count
		aliasName := ""
		var aliases []*checkerapi.Type
		if alias := checkerapi.Type_alias(subject); alias != nil {
			if alias.Symbol() != nil {
				aliasName = alias.Symbol().Name
			}
			aliases = alias.TypeArguments()
		}
		if fields[3] != aliasName {
			t.Fatal("wrong alias name")
		}
		aliasCount, err := strconv.Atoi(fields[cursor])
		if err != nil || aliasCount != len(aliases) {
			t.Fatal("wrong alias count")
		}
		for i, argument := range aliases {
			argumentID, _ := strconv.ParseUint(fields[cursor+1+i], 10, 64)
			if argumentID == 0 || p.typesByID[argumentID-1] != argument {
				t.Fatal("wrong alias argument")
			}
		}
		for _, invalid := range []string{"type-arguments", "type-arguments\n0", "type-arguments\n01", "type-arguments\n9007199254740991", "type-arguments\n" + id + "\nextra"} {
			if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), invalid); err == nil {
				t.Fatalf("accepted %q", invalid)
			}
		}
	}
}
