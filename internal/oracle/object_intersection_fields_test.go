package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An object intersection whose fields hold objects, arrays and boxed unions (Compiler's ruling on
// #qqvq6n7): a spread copy dropped before its original and after it, an object field the copy
// shares and that is written through one and read through the other, a boxed union field and an
// object that may be undefined, a union with undefined and null that keeps the two apart, and
// Object.assign onto a target with an object field, alone and in a program with graph types. Each
// agrees with Node natively, sanitized and on the JavaScript backend, and lets go of everything it
// made.
func TestObjectIntersectionFieldsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/object_intersection_fields"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			objectIntersectionAgrees(t, path)
		})
	}
	if programs != 6 {
		t.Fatalf("%d programs in %s, want 6: a program added or removed needs this count changed with it", programs, directory)
	}
}

// An intersection whose field holds the intersection itself, a chain walked and summed. Main's
// objectIntersection recursed without end on it, a goroutine stack overflow, so the admission
// delta's base can't classify it: it is written here rather than kept in the corpus.
func TestObjectIntersectionHoldingItselfAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "cyclic.a")
	source := `type Chain = { readonly value: number } & { readonly next: Chain | undefined };
function build(count: number): Chain | undefined {
    let chain: Chain | undefined = undefined;
    for (let index = 1; index <= count; index++) {
        const link: Chain = { value: index, next: chain };
        chain = link;
    }
    return chain;
}
let total = 0;
let at = build(4);
while (at !== undefined) {
    total += at.value;
    at = at.next;
}
console.log(` + "`${total}`" + `);
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	objectIntersectionAgrees(t, path)
}

// objectIntersectionAgrees holds a program to Node natively, sanitized and on JavaScript, and to the
// leak check.
func objectIntersectionAgrees(t *testing.T, path string) {
	t.Helper()
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %#v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Errorf("leaks:\n%s", report)
	}
}
