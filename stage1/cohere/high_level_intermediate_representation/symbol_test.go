package high_level_intermediate_representation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	nativechecker "github.com/system-inc/adamic/bridge/tsgo/checker"
)

func symbolFields(t *testing.T, wire string) []string {
	t.Helper()
	units := utf16.Encode([]rune(wire))
	var fields []string
	for cursor := 0; cursor < len(units); {
		last := cursor
		for last < len(units) && units[last] != '\n' {
			last++
		}
		if last == len(units) {
			t.Fatal("missing fact frame")
		}
		size, err := strconv.Atoi(string(utf16.Decode(units[cursor:last])))
		if err != nil || size < 0 || last+1+size > len(units) {
			t.Fatal("invalid fact length")
		}
		cursor = last + 1 + size
		fields = append(fields, string(utf16.Decode(units[last+1:cursor])))
	}
	return fields
}
func TestResidentSymbolFacts(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	temp := t.TempDir()
	config, file := filepath.Join(temp, "tsconfig.json"), filepath.Join(temp, "input.ts")
	source := `import Default from './dep';
import * as Namespace from './dep';
import { exported as Renamed } from './dep';
const moduleLocal = 1;
function Component(value) {
  const text = "世界🌍";
  value;
  { let value = 2; value; }
  const object = { value };
  const computed = { [text]: value };
  return value + moduleLocal + Default + Namespace + Renamed + unknownGlobal;
}
`
	write := func(path string, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(config, `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext"}}`)
	write(file, source)
	write(filepath.Join(temp, "dep.ts"), `export default 0; export const exported = 0;`)
	program, err := nativechecker.Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			nodes = append(nodes, n)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	var queries, manifest, want strings.Builder
	ids := map[*ast.Symbol]string{}
	seenIds := map[string]*ast.Symbol{}
	values := map[string]bool{}
	shorthand := false
	importKinds := map[string]bool{}
	for i, n := range nodes {
		wire, err := program.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol")
		if err != nil {
			t.Fatal(err)
		}
		repeated, err := program.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol")
		if err != nil || repeated != wire {
			t.Fatal("symbol facts not stable")
		}
		f := symbolFields(t, wire)
		if len(f) < 5 || f[0] != "1" || f[1] != "symbol" {
			t.Fatal("invalid symbol schema")
		}
		symbol := c.GetSymbolAtLocation(n)
		if n.Parent != nil && n.Parent.Kind == ast.KindShorthandPropertyAssignment && n.Parent.Name() == n {
			shorthand = true
			symbol = c.GetShorthandAssignmentValueSymbol(n.Parent)
		}
		if symbol == nil {
			symbol = n.Symbol()
		}
		if symbol == nil {
			if f[2] != "0" || f[3] != "" || f[4] != "0" {
				t.Fatal("unresolved symbol has identity")
			}
		} else {
			if f[2] == "0" {
				t.Fatal("resolved symbol lost identity")
			}
			if previous, ok := ids[symbol]; ok && previous != f[2] {
				t.Fatal("one symbol gained multiple IDs")
			}
			if previous, ok := seenIds[f[2]]; ok && previous != symbol {
				t.Fatal("different symbols share an ID")
			}
			ids[symbol] = f[2]
			seenIds[f[2]] = symbol
			if n.Text() == "value" {
				values[f[2]] = true
			}
		}
		fmt.Fprintf(&want, "symbol %s %s %s\n", f[2], f[3], f[4])
		count, err := strconv.Atoi(f[4])
		if err != nil || len(f) != 5+7*count {
			t.Fatal("invalid declarations")
		}
		for j := 0; j < count; j++ {
			d := f[5+j*7 : 12+j*7]
			same := "false"
			if d[1] == "1" {
				same = "true"
			}
			fmt.Fprintf(&want, "%s %s %s:%s %s %s %s\n", d[0], same, d[2], d[3], d[4], d[5], d[6])
			if d[6] == "./dep" {
				importKinds[d[0]] = true
			}
		}
		path := filepath.Join(temp, fmt.Sprintf("facts-%d", i))
		write(path, wire)
		fmt.Fprintln(&manifest, path)
		fmt.Fprintf(&queries, "%d\t%d\n", n.Pos(), n.End())
	}
	if len(values) != 2 || !shorthand || len(importKinds) != 3 {
		t.Fatalf("binding paths missing: %d shadow IDs, shorthand=%v, imports=%v", len(values), shorthand, importKinds)
	}
	queryPath, manifestPath := filepath.Join(temp, "queries.tsv"), filepath.Join(temp, "facts.txt")
	write(queryPath, queries.String())
	write(manifestPath, manifest.String())
	driver := filepath.Join(lane, "symbol_main.ts")
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", driver, "--replay", manifestPath)
	archive := filepath.Join(temp, "tsgo.a")
	command(t, root, nil, "go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive")
	binary := filepath.Join(temp, "symbols")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", driver, "-o", binary, "--tsgo", archive)
	for mode, got := range map[string][]byte{"Node replay": node, "native live": command(t, root, nil, binary, "--live", config, file, queryPath), "native replay": command(t, root, nil, binary, "--replay", manifestPath)} {
		if !bytes.Equal(got, []byte(want.String())) {
			t.Fatalf("%s differs: %s", mode, firstDifference(got, []byte(want.String())))
		}
	}
	// Successful-but-wrong reader must lose identity distinctions on both runtimes.
	mutant, err := os.MkdirTemp(filepath.Dir(lane), "hir-symbol-mutant-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mutant)
	for _, name := range []string{"symbol.ts", "symbol_main.ts", "symbol_live.ts"} {
		data, err := os.ReadFile(filepath.Join(lane, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "symbol.ts" {
			anchor := "this.identity = frames.natural();"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor not unique")
			}
			data = []byte(strings.Replace(string(data), anchor, "this.identity = frames.natural() === 0 ? 0 : 1;", 1))
		}
		write(filepath.Join(mutant, name), string(data))
	}
	mutantDriver := filepath.Join(mutant, "symbol_main.ts")
	mutantBinary := filepath.Join(temp, "symbol-mutant")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", mutantDriver, "-o", mutantBinary, "--tsgo", archive)
	for mode, got := range map[string][]byte{"Node": command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", mutantDriver, "--replay", manifestPath), "native": command(t, root, nil, mutantBinary, "--live", config, file, queryPath)} {
		if bytes.Equal(got, []byte(want.String())) {
			t.Fatalf("%s symbol identity collapse mutant survived", mode)
		}
	}
	t.Logf("%d exact symbol selectors agree across Go, native live, Node/native replay; shadow identity mutant caught", len(nodes))
}
