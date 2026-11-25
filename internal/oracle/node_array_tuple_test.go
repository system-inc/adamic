package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Adamic's ruling 3 on the scout pass (#bbssdcy, Nov 24, families 1 D and 13 B), as she amended it by
// @system_adamic_language's must: NodeArray<E> & readonly [T1..Tn] is held as its NodeArray part's layout, and a
// position read is a checked view read where Ti is narrower than E unless the proof of Ti, the predicate or a kind
// test of that position, dominates it with no call and no write between (internal/lower/node_array.go,
// nodeArrayTupleRead). Each program in testdata/node_array_tuple is a row below, held to Node on both backends.
const nodeArrayTupleDirectory = "internal/oracle/testdata/node_array_tuple"

// nodeArrayTuplePath is a program of the directory, by name, as an absolute path, which a stop names.
func nodeArrayTuplePath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, nodeArrayTupleDirectory, name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Language's t01 byte for byte (723 bytes, sha256 1da61f96eed4c9b8522092f6adc4d93bf220397af58a6dcf24f91fb1951b682b)
// can't reach its read. Its NodeArray declares pos and end, two of the four fields the NodeArray layout requires
// (internal/lower/node_array_layout.go:33 to :36, refused at :58: hasTrailingComma and transformFlags are missing),
// so `const list = makeArray(raw)` stops at the plain NodeArray<Node> binding, at t01.ts:13:7, before any
// intersection or position read, the same with head 1 as before it. Both backends lower the one IR, so the stop
// in lowering is each backend's decline. t01-alias.ts is the program that reaches the read.
func TestNodeArrayTupleExactT01Declines(t *testing.T) {
	t.Parallel()
	path := nodeArrayTuplePath(t, "t01.ts")
	if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "text undefined\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err := lowered(t, path)
	want := "t01.ts:13:7: stage 0 can't lower a value of type NodeArray<Node>, an array type with its own member pos, held as neither an array nor an object"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v, want a stop saying %q", err, want)
	}
}

// t01-alias.ts is t01 on a construction the subset admits: a four-field MutableNodeArray literal the caller keeps
// as raw, handed out as a NodeArray, so isOne proves readonly [Ident], swap() writes raw[0] = { kind: 2 } beneath
// the view, and a length test alone dominates list[0].text. Node prints "text undefined" with exit 0; both backends
// stop at the read, exit 70, naming it and where it is. Its direct-write variant writes raw[0] in place of the call.
//
// t01-kind-lie.ts is the kind test's lie on the same construction: swap() puts a Node of kind 1 with no text
// beneath the view between isOne and the kind test, so the kind test proves the tag and admits list[0] unchecked
// (the position check, by tag, would pass it too), and the stop is the Ident view's own field read of .text, exit 70
// on both backends, where Node prints "text undefined".
//
// Mutants (unapplied, beside the head's report): a length test skipping the check where Ti is narrower, and the
// predicate's proof kept across a call, each of which turns t01-alias.ts red, its stop moving to the .text read.
func TestNodeArrayTupleAliasStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, stderr string }{
		{"t01-alias.ts", "element read failed: list[0] at {path}:25:32 is not a Ident; expected Ident, found object"},
		{"t01-alias-direct-write.ts", "element read failed: list[0] at {path}:25:32 is not a Ident; expected Ident, found object"},
		{"t01-kind-lie.ts", "field read failed: list[0].text is not initialized; expected string, found missing"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := nodeArrayTuplePath(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "text undefined\n" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("the program must lower and stop at run time: %v", err)
			}
			want := run{stderr: []byte("adamic: panic: " + strings.ReplaceAll(stop.stderr, "{path}", path) + "\n"), exitCode: 70}
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			for backendName, got := range map[string]run{"native": native, "release": released(t, program), "javascript": backend} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", backendName, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// nodeArrayTupleAgreeing is each program head 1 admits, with Node's output. On t01-alias.ts's construction: no write
// at all, the read dominated by the predicate with nothing between (unchecked); a console.log between the predicate
// and the read (checked, and the element is still an Ident); a kind test of position 0 dominating the read with
// nothing between (unchecked); NodeArray<Node> & readonly [Node], where Ti is E, with a length test alone before
// list[0].kind, after swap() (the element read). createClassThisAssignmentBlock's statements, named once by a type
// alias and read in a function of their own (checked, since no proof is in that function). Then five programs that
// aren't tsc's shape: a two-position tuple part, the NodeArray part's fields and length read through the
// intersection, the intersection passed to a function and read there, the predicate in a conditional's condition,
// and a NodeArray of a union narrowed to one member. Then family 13's form: a function whose result is the
// intersection, read at its call (checked, since a call is no proof), through signature's own storage. Last, tsc's
// own three-member shape (types.ts:3056, utilities.ts:3974 and :4182): NodeArray<Node> & readonly [Ident, Literal
// | Ident, Shape] & Readonly<TextRange>, whose third member every property of which the NodeArray part already
// declares readonly at an identical type is implied (nodeArrayImpliedMember, Adamic's yes on #tjfrgyq), first
// narrowed to by a predicate and then, in implied-textrange-param.ts, held as a parameter's type, the road tsc's own
// sites take: the parameter's representation, with no predicate, as checkAliasSymbol's node is.
var nodeArrayTupleAgreeing = []struct{ name, node string }{
	{"t01-alias-no-swap.ts", "text string\n"},
	{"t01-alias-call.ts", "before read\ntext string\n"},
	{"t01-alias-kind.ts", "text string\n"},
	{"same-element-alias.ts", "kind 1\n"},
	{"static-block-read.ts", "statement value\n"},
	{"two-position.ts", "pair name 2\n"},
	{"field-read.ts", "fields 0 0 false 0 1\n"},
	{"passed-to-function.ts", "passed name\n"},
	{"conditional.ts", "conditional name\n"},
	{"union-element.ts", "union name\n"},
	{"returned.ts", "returned name\n"},
	{"implied-textrange.ts", "define name value 3 4 9 3\n"},
	{"implied-textrange-param.ts", "param name 3 4 9 3\n"},
}

// Each admitted program agrees with Node natively, sanitized and on JavaScript, and leaks nothing.
func TestNodeArrayTupleAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range nodeArrayTupleAgreeing {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path := nodeArrayTuplePath(t, pin.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != pin.node {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, pin.node, truth.stderr)
			}
			objectIntersectionAgrees(t, path)
		})
	}
}

// shrunk.ts pops the element through the alias after the predicate, so the checked read at list[0] finds no
// element, and Node throws its TypeError reading .text: each backend stops there in Node's words, exit 70, after
// Node's own output, as a narrowed read put back undefined does.
func TestNodeArrayTupleShrunkAgrees(t *testing.T) {
	t.Parallel()
	path := nodeArrayTuplePath(t, "shrunk.ts")
	want := onNode(t, path)
	if want.exitCode != 70 || string(want.stdout) != "length 0\n" || string(want.stderr) != "adamic: panic: TypeError: Cannot read properties of undefined (reading 'text')\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the program must lower and stop at run time: %v", err)
	}
	native, _ := natively(t, program)
	for backendName, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", backendName, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// static-block.ts is createClassThisAssignmentBlock's whole shape on a four-field NodeArray: a fresh static block
// whose statements is NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }], returned and
// read. Its first stop, before head 1 and with it alike, is the up-front method-signature-style refusal where the
// statements go into the returned block's, an existing rule on the meter's list and that unit's next stop, not this
// ruling's. static-block-read.ts, the same statements read in a function of their own, is what reaches the
// intersection, and agrees.
func TestNodeArrayTupleStaticBlockStops(t *testing.T) {
	t.Parallel()
	path := nodeArrayTuplePath(t, "static-block.ts")
	if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "block value\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err := lowered(t, path)
	want := "static-block.ts:7:11: Adamic 0.1 refuses a function taking (value: Statement, index: number, array: readonly Statement[]) => value is S"
	if err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "(method-signature-style)") {
		t.Fatalf("got %v, want a stop saying %q, (method-signature-style)", err, want)
	}
}

// implied-textrange.ts's twins, each a third member nodeArrayImpliedMember refuses, so the intersection is no
// NodeArray tuple and each stops where implied-textrange.ts stopped before the rule: at its predicate's declaration,
// the predicate's target being no shape the membership test knows. FlaggedRange adds a field NodeArray lacks,
// TextRange declares pos and end writable where NodeArray's ReadonlyTextRange holds them readonly (the member's
// side is the mutable one), and StringRange declares pos a string.
//
// Mutant (unapplied, beside the head's report): nodeArrayImpliedMember answering true for any member, which lowers
// implied-extra-field.ts and implied-writable.ts and turns their rows red.
func TestNodeArrayTupleImpliedMemberStops(t *testing.T) {
	t.Parallel()
	for _, stop := range []struct{ name, node, at string }{
		{"implied-extra-field.ts", "define name value 3 4 9 3\n", ":17:1: "},
		{"implied-writable.ts", "define name value 3 4 9 3\n", ":16:1: "},
		{"implied-string-pos.ts", "define name value 3 9 3\n", ":17:1: "},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path := nodeArrayTuplePath(t, stop.name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err := lowered(t, path)
			want := stop.name + stop.at + "stage 0 can't lower checked predicate overload target"
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want a stop saying %q", err, want)
			}
		})
	}
}

// implied-extra-field-param.ts is the parameter road's twin: a parameter typed with a third member that adds a field
// NodeArray lacks, in a function nothing calls, so the only thing that can stop it is the parameter type's
// representation. It stops there in the words tsc's census reports for utilities.ts:3974 and checker.ts:48415, where
// it stopped before the rule, so the parameter road and the predicate road answer one helper alike. The same mutant
// (nodeArrayImpliedMember true for any member) lowers it and turns this row red.
func TestNodeArrayTupleImpliedMemberParameterStops(t *testing.T) {
	t.Parallel()
	path := nodeArrayTuplePath(t, "implied-extra-field-param.ts")
	if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "define name 3 3\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	_, err := lowered(t, path)
	want := "implied-extra-field-param.ts:25:21: stage 0 can't lower a value of type NodeArray<Node> & readonly [Ident, Ident | Literal, Shape] & FlaggedRange, an intersection whose member NodeArray<Node> isn't a plain object shape"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v, want a stop saying %q", err, want)
	}
}

// Every program in the directory is a row above, so one added without its row can't go unheld. Beside each is its
// expected outcome (.stop, `lowers` or the stop), and beside each stop where Node runs its .pending: the two run-time
// stops' await Adamic's ruling on their held-stops rows, which the admission delta reads.
func TestNodeArrayTupleProgramsArePinned(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, nodeArrayTupleDirectory))
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]bool{"t01.ts": true, "t01-alias.ts": true, "t01-alias-direct-write.ts": true, "t01-kind-lie.ts": true, "static-block.ts": true, "shrunk.ts": true, "implied-extra-field.ts": true, "implied-writable.ts": true, "implied-string-pos.ts": true, "implied-extra-field-param.ts": true}
	for _, pin := range nodeArrayTupleAgreeing {
		pinned[pin.name] = true
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		if !pinned[entry.Name()] {
			t.Errorf("%s has no row", entry.Name())
		}
		delete(pinned, entry.Name())
	}
	for name := range pinned {
		t.Errorf("%s has a row and no program", name)
	}
}
