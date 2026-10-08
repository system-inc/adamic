package parser

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: establish a positive native control before parallel mutant builds.
func TestJsxMutants(t *testing.T) {
	t.Parallel()
	manifest, _ := jsxManifest(t)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	directory, _ := filepath.Abs(".")
	control := execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")
	if diff := difference(control.output, want); diff != "" {
		t.Fatal(diff)
	}
	changes := []struct{ name, from, to string }{
		{"text payload", "this.parser.node(id).text = this.parser.scanner.value;", "this.parser.node(id).text = this.parser.scanner.value + '!';"},
		{"whitespace flag", "whitespace ? '1' : '0'", "whitespace ? '0' : '0'"},
		{"namespace kind", "this.make('JsxNamespacedName', pos, [name, right])", "this.make('QualifiedName', pos, [name, right])"},
		{"self closing kind", "this.make('JsxSelfClosingElement', pos, header)", "this.make('JsxOpeningElement', pos, header)"},
		{"type argument comma", "this.parser.node(id).semantic = `${typeCount}:${typeTrailing ? 1 : 0}`;", "this.parser.node(id).semantic = `${typeCount}:0`;"},
		{"attribute list", "this.parser.node(id).list = attributes.length;", "this.parser.node(id).list = attributes.length + 1;"},
		{"expression kind", "this.make('JsxExpression', pos, children)", "this.make('ParenthesizedExpression', pos, children)"},
		{"child order", "this.make(fragment ? 'JsxFragment' : 'JsxElement', pos, children)", "this.make(fragment ? 'JsxFragment' : 'JsxElement', pos, children.slice().reverse())"},
		{"text start", "const id = this.make('JsxText', pos);", "const id = this.make('JsxText', pos + 1);"},
	}
	slots := make(chan struct{}, 3)
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			slots <- struct{}{}
			defer func() { <-slots }()
			mutant := copyPort(t, "jsx.ts", change.from, change.to)
			node := wholeNode(t, mutant, manifest, false)
			binary := buildPort(t, mutant, true)
			native := execute(t, "", binary, "--manifest", manifest, "--whole")
			for _, side := range []struct {
				name   string
				output []byte
			}{{"Node", node.output}, {"native", native.output}} {
				if bytes.Equal(side.output, want) {
					t.Fatalf("%s mutant survived", side.name)
				}
				t.Logf("%s compiled mutant caught: %s", side.name, strings.Split(difference(side.output, want), "\n")[0])
			}
		})
	}
}
