package oracle

import (
	"path/filepath"
	"testing"
)

func TestCheckedViewRankedArrayUnionContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"ranked3-pattern", "1:1\n", ""},
		{"ranked3-pattern-wrong-array", "undefined\n", "field read failed: pattern(raw).elements is not a NodeArray<BindingElement> | NodeArray<ArrayBindingElement>; expected NodeArray<BindingElement> | NodeArray<ArrayBindingElement>, found number"},
		{"ranked3-types", "1:1\n", ""},
		{"ranked3-types-wrong-array", "undefined\n", "field read failed: types(raw).types is not a Type[]; expected Type[], found number"},
		{"ranked3-signature", "1:1\n", ""},
		{"ranked3-signature-wrong-array", "undefined\n", "field read failed: signature(raw).parameters is not a NodeArray<ParameterDeclaration>; expected NodeArray<ParameterDeclaration>, found number"},
		{"ranked3-class", "1:1\n", ""},
		{"ranked3-class-wrong-array", "undefined\n", "field read failed: classNode(raw).members is not a NodeArray<ClassElement>; expected NodeArray<ClassElement>, found number"},
		{"ranked3-reference", "1:1\n", ""},
		{"ranked3-reference-wrong-array", "undefined\n", "field read failed: reference(raw).typeArguments matches no member of NodeArray<TypeNode> | undefined; expected NodeArray<TypeNode> | undefined, found number"},
		{"ranked3-pattern-lazy", "2:1\n", ""},
		{"ranked3-pattern-wrong-element", "undefined\n", "element read failed: values[0] expected ArrayBindingElement, found number"},
		{"ranked3-pattern-wrong-tag", "1\n", "field read failed: values[0].kind expected \"binding\" | \"omitted\", found string other"},
		{"ranked3-mutable-array-unread", "pattern\n", ""},
		{"ranked3-reference-absent", "absent\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			truth := onNode(t, path)
			want := run{stdout: []byte(probe.node)}
			if difference := disagreement(want, truth); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; stderr %q; stdout %q; exit %d", difference, got.stderr, got.stdout, got.exitCode)
				}
			}
			if probe.diagnostic == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewMutableArrayUnionRefusal(t *testing.T) {
	t.Parallel()
	path, pathErr := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane2/ranked3-mutable-array-read.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	_, err := lowered(t, path)
	if got := onNode(t, path); got.exitCode != 0 || string(got.stdout) != "1\n" {
		t.Fatalf("Node: %#v", got)
	}
	if err == nil {
		t.Fatal("mutable array union read was admitted")
	}
	if got := err.Error(); got != path+":23:18: Adamic 0.1 refuses checked view read of field elements with unsupported mutable array union contract; prove or implement the mutable array union contract before reading this field" {
		t.Fatalf("refusal: %s", got)
	}
}
