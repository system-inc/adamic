package oracle

import "testing"

func TestCheckedViewArrayReferenceWrites(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"reference-array-set", "2:item2\n", ""},
		{"reference-array-push", "2\n2:item2\n", ""},
		{"reference-array-alias", "3:item2\n3:item2:4\n", ""},
		{"reference-array-empty", "1:item2\n", ""},
		{"reference-array-slice", "2:item3:item2\n", ""},
		{"reference-array-lazy", "2:first\n", ""},
		{"reference-array-missing-field", "written\nundefined\n", "element write failed: <array write> expected { value: number; text: string; secret: number; }, found { value: number; text: string; }"},
		{"reference-array-narrow-field", "written\n", "element write failed: <array write> expected { value: 1; text: string; }, found { value: number; text: string; }"},
		{"reference-array-readonly-incoming", "written\n", "element write failed: <array write> expected { value: number; text: string; }, found { readonly value: number; readonly text: string; }"},
		{"reference-array-mutable-incoming", "written\n", "element write failed: <array write> expected { value: number; text: string; }, found { value: 2; text: string; }"},
		{"reference-array-uncertified", "written\n", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"reference-array-source-alias", "written:1\n", "field write failed: property 'value' on record { kind: \"entry\"; value: 7; text: string; } at reference-array-source-alias.a:11 has no compatible declared slot"},
		{"reference-array-uninitialized", "written\n", "cast failed: field read failed: <array write>.value is not initialized; expected number, found uninitialized"},
		{"reference-array-evaluation", "3:item2\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.node)}
			if got := onNode(t, path); disagreement(want, got) != "" {
				t.Fatalf("Node control: %#v", got)
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
