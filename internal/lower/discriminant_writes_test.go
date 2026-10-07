package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestDiscriminantWrites(t *testing.T) {
	t.Parallel()
	const types = "type Leaf = { kind: 'leaf'; readonly value: number };\ntype Branch = { kind: 'branch'; readonly label: string };\ntype SyntaxNode = Leaf | Branch;\n"
	for _, probe := range []struct{ name, source, field string }{
		{"union", types + "function change(node: SyntaxNode): void { node.kind = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"base interface", types + "interface Base { kind: string }\nfunction change(node: Base): void { node.kind = 'branch'; }\nchange({ kind: 'leaf' });", "kind"},
		{"same literal member", types + "function change(node: Leaf): void { node.kind = 'leaf'; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"bracket", types + "function change(node: SyntaxNode): void { node['kind'] = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"constant bracket", types + "const field = 'kind';\nfunction change(node: SyntaxNode): void { node[field] = 'branch'; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"compound", types + "function change(node: SyntaxNode): void { node.kind += ''; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"array destructuring", types + "function change(node: SyntaxNode): void { [node.kind] = ['branch']; }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"object destructuring", types + "function change(node: SyntaxNode): void { ({ kind: node.kind } = { kind: 'branch' }); }\nchange({ kind: 'leaf', value: 4 });", "kind"},
		{"method", "class Leaf { kind: 'leaf' = 'leaf'; change(): void { this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nnew Leaf().change();", "kind"},
		{"number increment", "type Node = { tag: 1 } | { tag: 2 };\nfunction change(node: Node): void { node.tag++; }\nchange({ tag: 1 });", "tag"},
		{"escaped constructor", "let escaped: Leaf | undefined;\nclass Leaf { kind: 'leaf' = 'leaf'; constructor() { escaped = this; this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);", "kind"},
		{"closure in constructor", "class Leaf { kind: 'leaf' = 'leaf'; constructor() { const change = (): void => { this.kind = 'leaf'; }; change(); } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);", "kind"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want discriminant refusal, got %v", err)
			}
			want := "Adamic 0.1 refuses a write to discriminant field '" + probe.field + "' after construction; changing variant means building a new object"
			if !strings.HasSuffix(refused.Error(), want) {
				t.Fatalf("want exact refusal %q, got %v", want, err)
			}
		})
	}
}

func TestDiscriminantConstruction(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Leaf { kind: 'leaf'; constructor() { this.kind = 'leaf'; this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);",
		"class Leaf { kind: 'leaf' = 'leaf'; constructor() { this.kind = 'leaf'; } }\ntype Node = Leaf | { kind: 'branch' };\nconsole.log(new Leaf().kind);",
		"type Node = { kind: 'leaf' } | { kind: string; value: number };\nfunction change(node: { kind: string; value: number }): void { node.kind = 'new'; }\nchange({ kind: 'old', value: 1 });",
		"class Plain { kind: 'leaf' = 'leaf'; change(): void { this.kind = 'leaf'; } }\nnew Plain().change();",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("want accepted construction or non-discriminant, got %v", err)
		}
	}
}

func TestDiscriminantCompoundConstruction(t *testing.T) {
	t.Parallel()
	source := "class Leaf { tag: 1 = 1; value = 4; constructor() { this.tag++; } }\ntype Node = Leaf | { tag: 2; label: string };\nfunction show(node: Node): void { if (node.tag === 2) { console.log(node.label); } }\nshow(new Leaf());"
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want constructor tag update refused, got %v", err)
	}
	want := "Adamic 0.1 refuses a compound write to discriminant field 'tag' whose result is not proven to keep its declared literal; initialize the declared literal directly; changing variant means building a new object"
	if !strings.HasSuffix(refused.Error(), want) {
		t.Fatalf("want exact refusal %q, got %v", want, err)
	}
}
