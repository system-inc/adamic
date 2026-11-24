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
// Mutants (unapplied, beside the head's report): a length test skipping the check where Ti is narrower, and the
// predicate's proof kept across a call, each of which lets t01-alias.ts print "text undefined" and turns it red.
func TestNodeArrayTupleAliasStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"t01-alias.ts", "t01-alias-direct-write.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := nodeArrayTuplePath(t, name)
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != "text undefined\n" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("the program must lower and stop at run time: %v", err)
			}
			want := run{stderr: []byte("adamic: panic: element read failed: list[0] at " + path + ":25:32 is not a Ident; expected Ident, found object\n"), exitCode: 70}
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
// list[0].kind, after swap() (the element read). Then four programs that aren't tsc's shape: a two-position tuple
// part, the NodeArray part's fields and length read through the intersection, the intersection passed to a function
// and read there (checked, since no proof is in that function), and the predicate in a conditional's condition.
var nodeArrayTupleAgreeing = []struct{ name, node string }{
	{"t01-alias-no-swap.ts", "text string\n"},
	{"t01-alias-call.ts", "before read\ntext string\n"},
	{"t01-alias-kind.ts", "text string\n"},
	{"same-element-alias.ts", "kind 1\n"},
	{"two-position.ts", "pair name 2\n"},
	{"field-read.ts", "fields 0 0 false 0 1\n"},
	{"passed-to-function.ts", "passed name\n"},
	{"conditional.ts", "conditional name\n"},
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

// static-block.ts is createClassThisAssignmentBlock's shape on a four-field NodeArray: a fresh static block whose
// statements is NodeArray<Statement> & readonly [ExpressionStatement & { expression: "value" }], returned and
// read. It reaches the intersection now and stops at the next rule on its road, method-signature-style, where the
// intersection's statements go into the block's: that unit's next stop, not this ruling's.
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

// Every program in the directory is a row above, so one added without its row can't go unheld.
func TestNodeArrayTupleProgramsArePinned(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(filepath.Join(repository, nodeArrayTupleDirectory))
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]bool{"t01.ts": true, "t01-alias.ts": true, "t01-alias-direct-write.ts": true, "static-block.ts": true}
	for _, pin := range nodeArrayTupleAgreeing {
		pinned[pin.name] = true
	}
	for _, entry := range entries {
		if !pinned[entry.Name()] {
			t.Errorf("%s has no row", entry.Name())
		}
		delete(pinned, entry.Name())
	}
	for name := range pinned {
		t.Errorf("%s has a row and no program", name)
	}
}
