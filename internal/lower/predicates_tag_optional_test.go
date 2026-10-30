package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A tag-proven predicate whose target has an optional field is refused for the view, which can't check
// that field yet (#1yr7rdh), and the refusal says so (#1w5bve2, Correctness's trace, Oct 29). It used to
// return the flow prover's "return is not proven", which blamed a return that is a trusted tag check:
// 287 of tsc's predicates in nodeTests.ts and its neighbours. Correctness's two programs differ by one
// field: plain is admitted. In .ts optional is admitted too, its optional field checked where the view
// reads it (#mr87ptk); in .a it's refused, naming the field.
func TestTagPredicateOptionalFieldNamesItsReason(t *testing.T) {
	t.Parallel()
	plain, err := os.ReadFile("testdata/predicates_tag_optional/plain.ts")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNodeAs(t, string(plain), ".ts")
	optional, err := os.ReadFile("testdata/predicates_tag_optional/optional.ts")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNodeAs(t, string(optional), ".ts")
	// In .a it needs proof, so it's refused, naming the field.
	_, err = lowerSource(t, string(optional))
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a tag-proven predicate whose target has an optional field (jsDoc on Identifier) that the view can't check" {
		t.Fatalf("want the optional field refused in .a, got %v", err)
	}
	// An asserts claim proven by its tag takes the same branch, in .a too: this row pinned the flow
	// prover's "normal return" until the reason was named (predicates_test.go's "assertion wider shape").
	_, err = lowerSource(t, `type A = { readonly kind: 'a' }; type B = { readonly kind: 'b' }; type Claimed = { readonly kind: 'a'; readonly extra?: number }; function assertA(x: A | B): asserts x is Claimed { if (x.kind !== 'a') { throw new Error('bad'); } }`)
	if !errors.As(err, &refused) || refused.What != "a tag-proven predicate whose target has an optional field (extra on Claimed) that the view can't check" {
		t.Fatalf("want the asserts claim's optional field named, got %v", err)
	}
}

// In .ts a tag-proven predicate admits a target with an optional field, and the view checks the field
// where it's read (#mr87ptk, Adamic's conditions on #1w5bve2). Absent, or present and right, agrees with
// Node. A lying layout, an object where the view says Doc, stops at the read with the view's located
// message on native under the sanitizers and on the JavaScript backend, read directly, through a call,
// through a stored field, and destructured. A write is NotYet, never an unchecked store.
const tagOptionalPrelude = `interface Node { readonly kind: number; }
interface Doc extends Node { readonly kind: 320; readonly text: string; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; jsDoc?: Doc; }
function isIdentifier(node: Node): node is Identifier {
    return node.kind === 80;
}
const raw = { kind: 80, escapedText: "x", jsDoc: { kind: 11, text: "lie" } };
`

func TestTagPredicateOptionalFieldIsCheckedAtTheRead(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"absent", "function show(node: Node): void { if (isIdentifier(node)) { console.log(node.escapedText); console.log(String(node.jsDoc === undefined)); } }\nshow({ kind: 80, escapedText: 'x' } as Node);"},
		{"present", "function show(node: Node): void { if (isIdentifier(node)) { const doc = node.jsDoc; console.log(doc === undefined ? 'none' : doc.text); } }\nconst documented = { kind: 80, escapedText: 'x', jsDoc: { kind: 320, text: 'doc' } };\nshow(documented);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, tagOptionalPrelude+probe.source, ".ts")
		})
	}
	for _, probe := range []struct{ name, source, read string }{
		{"lie direct", "function show(node: Node): void { if (isIdentifier(node)) { const doc = node.jsDoc; console.log(doc === undefined ? 'none' : String(doc.kind)); } }\nshow(raw);", "node.jsDoc"},
		{"lie through call", "function docOf(id: Identifier): string { const doc = id.jsDoc; return doc === undefined ? 'none' : String(doc.kind); }\nfunction show(node: Node): void { if (isIdentifier(node)) { console.log(docOf(node)); } }\nshow(raw);", "id.jsDoc"},
		{"lie stored", "interface Holder { id: Identifier; }\nfunction show(node: Node): void { if (isIdentifier(node)) { const holder: Holder = { id: node }; const doc = holder.id.jsDoc; console.log(doc === undefined ? 'none' : String(doc.kind)); } }\nshow(raw);", "holder.id.jsDoc"},
		{"lie destructured", "function show(node: Node): void { if (isIdentifier(node)) { const { jsDoc } = node; console.log(jsDoc === undefined ? 'none' : String(jsDoc.kind)); } }\nshow(raw);", "jsDoc (field jsDoc)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, tagOptionalPrelude+probe.source)
			if err != nil {
				t.Fatal(err)
			}
			// Every read of jsDoc is the checked view read (an ir.Property naming its view and its
			// contract, on a field the view checks), the node Emit's TestViewOptionalReadsStopAtTheRead
			// pins against a read that believes the slot. A predicate-registered view routed to a plain
			// field read turns this red (Lower and Compiler, condition 4 on #mr87ptk).
			reads, checked := 0, 0
			inspect := func(node any) bool {
				if read, ok := node.(ir.Property); ok && read.Name == "jsDoc" && !read.Method {
					reads++
					if read.View != "" && read.ViewContract > 0 && program.CheckedField(read.ViewReceiverTypeID, read.Name) {
						checked++
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if reads == 0 || checked != reads {
				t.Fatalf("want every jsDoc read through the checked view read, got %d of %d", checked, reads)
			}
			directory := t.TempDir()
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			generated := filepath.Join(directory, "generated.mjs")
			if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				t.Fatal(err)
			}
			want := "adamic: panic: field read failed: " + probe.read + " matches no member of Doc | undefined; expected Doc | undefined, found object\n"
			for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				code := 0
				if err := command.Run(); err != nil {
					var exit *exec.ExitError
					if !errors.As(err, &exit) {
						t.Fatal(err)
					}
					code = exit.ExitCode()
				}
				if code != 70 || out.Len() != 0 || errs.String() != want {
					t.Fatalf("%s: exit %d stdout %q stderr %q; want 70 and %q", command.Path, code, out.String(), errs.String(), want)
				}
			}
		})
	}
	_, err := lowerTypeScriptAssertionSource(t, tagOptionalPrelude+"function show(node: Node): void { if (isIdentifier(node)) { node.jsDoc = { kind: 320, text: 'set' }; } }\nshow({ kind: 80, escapedText: 'x' } as Node);")
	var pending *NotYet
	if !errors.As(err, &pending) {
		t.Fatalf("want a write through the view NotYet, never an unchecked store, got %v", err)
	}
}
