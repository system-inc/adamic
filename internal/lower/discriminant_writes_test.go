package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscriminantWrites(t *testing.T) {
	t.Parallel()
	const types = "type Leaf = { kind: 'leaf'; readonly value: number };\ntype Branch = { kind: 'branch'; readonly label: string };\ntype SyntaxNode = Leaf | Branch;\n"
	for _, probe := range []struct{ name, source, field, variant string }{
		{"literal flags", "/// <reference path='./flags.d.a.ts' />\ntype Node = { kind: 'a'; flags: Flags.B } | { kind: 'b'; flags: 1 }; function change(node: Node): void { node.flags = 2; } change({ kind: 'b', flags: 1 });", "flags", "Flags.B"},
		{"escaped unsafe constructor", "let escaped: Node | undefined; type Node = { kind: 'a' } | { kind: 'b' }; class Holder { constructor(node: Node) { escaped = node; node.kind = 'b'; } } const holder = new Holder({ kind: 'a' });", "kind", "\"b\""},

		{"method", "class Base { kind: string = 'leaf'; change(): void { this.kind = 'branch'; } } class Leaf extends Base { kind: 'leaf' = 'leaf'; } class Branch extends Base { kind: 'branch' = 'branch'; } type Node = Leaf | Branch; new Leaf().change();", "kind", "\"branch\""},

		{"this escapes constructor", "let escaped: Base | undefined; class Base { kind: string = 'leaf'; constructor() { escaped = this; this.kind = 'branch'; } } type Node = { kind: 'leaf' } | { kind: 'branch' }; console.log(new Base().kind);", "kind", "\"branch\""},
		{"local escapes", "let escaped: Base | undefined; interface Base { kind: string } type Node = { kind: 'leaf' } | { kind: 'branch' }; function build(): void { const node: Base = { kind: 'new' }; escaped = node; node.kind = 'branch'; } build();", "kind", "\"branch\""},

		{"union", types + "function change(node: SyntaxNode): void { node.kind = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\""},
		{"base interface", types + "interface Base { kind: string }\nfunction change(node: Base): void { node.kind = 'branch'; }\nchange({ kind: 'leaf' });", "kind", "\"branch\""},
		{"bracket", types + "function change(node: SyntaxNode): void { node['kind'] = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\""},
		{"constant bracket", types + "const field = 'kind';\nfunction change(node: SyntaxNode): void { node[field] = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\""},
		{"compound", types + "function change(node: SyntaxNode): void { node.kind += ''; }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\" or \"leaf\""},
		{"array destructuring", types + "function change(node: SyntaxNode): void { [node.kind] = ['branch']; }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\""},
		{"object destructuring", types + "function change(node: SyntaxNode): void { ({ kind: node.kind } = { kind: 'branch' }); }\nchange({ kind: 'leaf', value: 4 });", "kind", "\"branch\""},
		{"number increment", "type Node = { tag: 1 } | { tag: 2 };\nfunction change(node: Node): void { node.tag++; }\nchange({ tag: 1 });", "tag", "1 or 2"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerDiscriminantSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want discriminant refusal, got %v", err)
			}
			prefix := "a write to discriminant field '" + probe.field + "' that could move the object to variant "
			if refused.What != prefix+probe.variant || refused.Fix != "changing variant means building a new object" {
				t.Fatalf("wrong refusal: %v", err)
			}

		})
	}
}

func TestDiscriminantConstruction(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"type Node = { kind: 'a' | 'b'; value: number } | { kind: 'c'; label: string }; function change(node: { kind: 'a' | 'b'; value: number }): void { node.kind = 'b'; } change({ kind: 'a', value: 1 });",
		"interface Base { kind: string } type Node = { kind: 'leaf' } | { kind: 'branch' }; function build(): void { const node: Base = { kind: 'new' }; node.kind = 'branch'; console.log(node.kind); } build();",

		"type Leaf = { kind: 'leaf'; value: number }; type Node = Leaf | { kind: 'branch' }; function change(node: Leaf): void { node.kind = 'leaf'; } change({ kind: 'leaf', value: 4 });",
		"class Leaf { kind: 'leaf' = 'leaf'; change(): void { this.kind = 'leaf'; } } type Node = Leaf | { kind: 'branch' }; new Leaf().change();",
		"let escaped: Leaf | undefined; class Leaf { kind: 'leaf' = 'leaf'; constructor() { escaped = this; this.kind = 'leaf'; } } type Node = Leaf | { kind: 'branch' }; console.log(new Leaf().kind);",
		"/// <reference path='./flags.d.a.ts' />\ntype Node = { kind: 'a'; flags: Flags } | { kind: 'b'; flags: OtherFlags }; function change(node: Node): void { node.flags |= 2; } change({ kind: 'a', flags: 1 });",

		"class Leaf { kind: 'leaf'; constructor() { this.kind = 'leaf'; this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);",
		"class Leaf { kind: 'leaf' = 'leaf'; constructor() { this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);",
		"type Node = { kind: 'leaf'; value: string } | { kind: string; value: number };\nfunction change(node: { kind: string; value: number }): void { node.kind = 'new'; }\nchange({ kind: 'old', value: 1 });",
		"class Plain { kind: 'leaf' = 'leaf'; change(): void { this.kind = 'leaf'; } }\nnew Plain().change();",
	} {
		if _, err := lowerDiscriminantSource(t, source); err != nil {
			t.Fatalf("want accepted construction or non-discriminant, got %v", err)
		}
	}
}

func TestDiscriminantCompoundConstruction(t *testing.T) {
	t.Parallel()
	source := "class Leaf { tag: 1 = 1; value = 4; constructor() { this.tag++; } }\ntype Node = Leaf | { tag: 2; label: string };\nfunction show(node: Node): void { if (node.tag === 2) { console.log(node.label); } }\nshow(new Leaf());"
	_, err := lowerDiscriminantSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want constructor tag update refused, got %v", err)
	}
	want := "Adamic 0.1 refuses a compound write to discriminant field 'tag' whose result is not proven to keep its declared literal; initialize the declared literal directly; changing variant means building a new object"
	if !strings.HasSuffix(refused.Error(), want) {
		t.Fatalf("want exact refusal %q, got %v", want, err)
	}
}

// Ambient enum types are supplied by a declaration file: Adamic has no runtime enum
// syntax. The executable program uses numeric values, as Node's stripped program does.
func lowerDiscriminantSource(t *testing.T, source string) (*ir.Program, error) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "main.a")
	if err := os.WriteFile(filepath.Join(dir, "flags.d.a"), []byte("export {}; declare global { enum Flags { A = 1, B = 2 } enum OtherFlags { A = 1, B = 2 } }"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	return Lower(context.Background(), program)
}
