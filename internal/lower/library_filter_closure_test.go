package lower

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

// libraryFilterTags is tsc's JSDoc tags in miniature: a readonly array of tags, one kind of which an
// isA predicate narrows to, and getAll, getAllJSDocTags' shape at utilitiesPublic.ts:1289, forwarding
// its predicate parameter to the library's ReadonlyArray filter.
const libraryFilterTags = `interface Tag {
	readonly kind: "a" | "b";
	readonly text: string;
}
interface ATag extends Tag {
	readonly kind: "a";
	readonly extra: number;
}
function getAll<T extends Tag>(tags: readonly Tag[], predicate: (tag: Tag) => tag is T): readonly T[] {
	return tags.filter(predicate);
}
const a: ATag = { kind: "a", text: "x", extra: 1 };
const b: Tag = { kind: "b", text: "y" };
const tags: readonly Tag[] = [a, b];
`

// The printing tail: each tag the filter kept, read as ATag.
const libraryFilterPrint = `for (const tag of getAll(tags, isA)) {
	console.log(tag.text + " " + String(tag.extra));
}
`

// A predicate parameter forwarded into the library's filter is passed as its checked closure (#ea4d18v,
// #jqf1drp's closure reaching lib's filter): filter makes a fresh array, so its narrower result aliases
// nothing wider, and every element it keeps passed the claim at run time. The honest generic shape lowers in
// .ts and agrees with Node on the JavaScript backend and natively under the sanitizers. In .a, where isA's
// body proves its claim, getAll's forward stays refused, naming the fix.
// Mutant: libraryFilterContract answering false, and getAll's filter is refused as an unproven predicate
// argument again, as on main.
func TestForwardedPredicateIntoLibraryFilter(t *testing.T) {
	t.Parallel()
	source := libraryFilterTags + `function isA(tag: Tag): tag is ATag {
	return tag.kind === "a";
}
` + libraryFilterPrint
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, out, errs := brandedNative(t, program); code != 0 || out != "x 1\n" || errs != "" {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", code, out, errs, "x 1\n")
	}
	_, err := lowerSourceAs(t, source, ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "a parameter carries no proof of its claim") {
		t.Fatalf("want .a to keep refusing the forwarded predicate, naming why, got %v", err)
	}
}

// A predicate that says yes to the B tag is a lie the filter would keep: Node prints "y undefined" for it.
// Adamic stops it at a run-time check of the claim inside the filter, when the predicate says yes to B, so
// before the loop prints anything, on the JavaScript backend and natively alike, with one panic.
func TestForwardedPredicateIntoLibraryFilterStopsALie(t *testing.T) {
	t.Parallel()
	source := libraryFilterTags + `function isA(tag: Tag): tag is ATag {
	return tag.text.length > 0;
}
` + libraryFilterPrint
	path := writeAgreementSource(t, "main.ts", source)
	node := runAgreementNode(t, path)
	if node.code != 0 || string(node.stdout) != "x 1\ny undefined\n" {
		t.Fatalf("source Node: exit %d stdout %q, want exit 0 and the lie read", node.code, node.stdout)
	}
	program := lowerAgreementEntry(t, path)
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	backend := runAgreementNode(t, generated)
	if backend.code != 70 || string(backend.stdout) != "" || !strings.HasPrefix(string(backend.stderr), "adamic: panic: predicate ") {
		t.Fatalf("JavaScript backend: exit %d stdout %q stderr %q; want exit 70 at a predicate check, inside the filter, before anything prints", backend.code, backend.stdout, backend.stderr)
	}
	code, out, errs := brandedNative(t, program)
	if code != 70 || out != "" || !bytes.Equal([]byte(errs), backend.stderr) {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 70 and the JavaScript backend's %q", code, out, errs, backend.stderr)
	}
}
