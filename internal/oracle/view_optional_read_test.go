package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An optional field read through a checked view (#1yr7rdh), in tsc's shape: a tag picks the
// declaration and its optional node-typed field is read, whole, through ?. or destructured (which main
// panicked on), an optional field of a union of nodes, ?. of a required union, optional and ?.
// scalars, and ?. through a receiver that is null. An absent field and an absent or null receiver
// read undefined; a present one is checked as any other read is. Each agrees with Node natively,
// sanitized and on the JavaScript backend. Each casts through an interface that extends its source
// with no tag of its own, the downcast .a refuses (#995z2fa), so each is .ts.
func TestViewOptionalReadsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_optional_read"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, path)
		})
	}
	if programs != 5 {
		t.Fatalf("%d programs in %s, want 5: a program added or removed needs this count changed with it", programs, directory)
	}
}

// Where the object doesn't hold what the view says, an optional read stops at the read with the
// view's located message and exit 70, after the absent case has printed what Node prints: an
// optional name that holds a number, read or destructured, or a node of the wrong kind, which only the
// union's membership check can tell, a present name whose escapedText is a number, read through ?.,
// and a required union through ?. that holds a boolean. Node, without the checks,
// prints on. Mutants (stage3/fixtures/assertions/19_identifier_kind-mutants.md): an absent or ?. read
// that believes the slot, in either backend, which runs on past the stop.
func TestViewOptionalReadsStopAtTheRead(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/view_optional_read_stops/"
	for _, stop := range []struct{ name, stdout, stderr string }{
		{"name_number.ts", "anonymous\n", "field read failed: declaration.name matches no member of Identifier | undefined; expected Identifier | undefined, found number"},
		{"destructured_number.ts", "anonymous:-\n", "field read failed: name (field name) matches no member of Identifier | undefined; expected Identifier | undefined, found number"},
		{"name_wrong_node.ts", "anonymous\n", "field read failed: declaration.name matches no member of Identifier | undefined; expected Identifier | undefined, found object"},
		{"chained_text.ts", "none\n", "field read failed: (node as FunctionDeclaration).name?.escapedText is not a string; expected string, found number"},
		{"specifier_name.ts", "none\n", "field read failed: specifier(node)?.name matches no member of Identifier | StringLiteral; expected Identifier | StringLiteral, found boolean"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, filepath.Join(repository, directory+stop.name))
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte("adamic: panic: " + stop.stderr + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}

