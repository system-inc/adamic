package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// tsc's visitNodes (visitorPublic.ts) in miniature, its body as 99-apple-visit-nodes-guard adapts it: a
// NodeArray over ReadonlyArray with tsc's four metadata fields, overload 1 promising NodeArray<TOut> |
// (TInArray & undefined) for a predicate test, overload 2 for a boolean one, and an implementation that
// returns its array parameter under `nodes === undefined`, or an array it checked element by element with
// the test in the statement right before the return. visitArrayWorker keeps its shape and its ungated
// check of each kept node; it hands back undefined when nothing changed, and createNodeArray builds the
// NodeArray as tsc's factory does, an array literal cast to MutableNodeArray, here written in place and
// handed straight to its NodeArray result, so no writable view of it is kept. start and count are left
// out, as nothing here reads them.
const visitNodesTypes = `interface Node {
    readonly kind: number;
    readonly flags: number;
}
interface Identifier extends Node {
    readonly kind: 80;
    readonly text: string;
}
interface Keyword extends Node {
    readonly kind: 100;
}
interface NodeArray<T extends Node> extends ReadonlyArray<T> {
    readonly pos: number;
    readonly end: number;
    readonly hasTrailingComma: boolean;
    transformFlags: number;
}
interface MutableNodeArray<T extends Node> extends Array<T> {
    pos: number;
    end: number;
    hasTrailingComma: boolean;
    transformFlags: number;
}
`

const visitNodesHelpers = `type Visitor = (node: Node) => Node | undefined;
namespace Debug {
    export function fail(message: string): never {
        throw new Error("Debug Failure. " + message);
    }
}
function createNodeArray(elements: readonly Node[]): NodeArray<Node> {
    const first = elements[0];
    const second = elements[1];
    if (elements.length === 1 && first !== undefined) {
        return [first] as MutableNodeArray<Node>;
    }
    if (elements.length === 2 && first !== undefined && second !== undefined) {
        return [first, second] as MutableNodeArray<Node>;
    }
    throw new Error("createNodeArray takes one or two elements");
}
function visitArrayWorker(nodes: NodeArray<Node>, visitor: Visitor, test: ((node: Node) => boolean) | undefined): readonly Node[] | undefined {
    const updated: Node[] = [];
    let changed = false;
    for (const node of nodes) {
        const visited = visitor(node);
        if (visited === undefined) {
            changed = true;
            continue;
        }
        if (test !== undefined && !test(visited)) Debug.fail("Node did not pass test.");
        if (visited !== node) {
            changed = true;
        }
        updated.push(visited);
    }
    return changed ? updated : undefined;
}
`

const visitNodesOverloads = `function visitNodes<TIn extends Node, TInArray extends NodeArray<TIn> | undefined, TOut extends Node>(nodes: TInArray, visitor: Visitor, test: (node: Node) => node is TOut): NodeArray<TOut> | (TInArray & undefined);
function visitNodes<TIn extends Node, TInArray extends NodeArray<TIn> | undefined>(nodes: TInArray, visitor: Visitor, test?: (node: Node) => boolean): NodeArray<Node> | (TInArray & undefined);
`

const visitNodesImplementation = `function visitNodes(nodes: NodeArray<Node> | undefined, visitor: Visitor, test?: (node: Node) => boolean): NodeArray<Node> | undefined {
    if (nodes === undefined) {
        return nodes;
    }
    const updated = visitArrayWorker(nodes, visitor, test);
    if (updated !== undefined) {
        const updatedArray = createNodeArray(updated);
        if (test !== undefined) {
            for (const testedNode of updatedArray) {
                if (!test(testedNode)) Debug.fail("False expression: Unexpected node.\r\nVerbose Debug Information: Node array did not pass test.");
            }
        }
        return updatedArray;
    }
    if (test !== undefined) {
        for (const testedNode of nodes) {
            if (!test(testedNode)) Debug.fail("False expression: Unexpected node.\r\nVerbose Debug Information: Node array did not pass test.");
        }
    }
    return nodes;
}
`

// The nodes: two identifiers made as Identifier and held as Node, so the literal NodeArray's element is
// Node, a keyword, an identity visitor and one that renames each identifier.
const visitNodesValues = `function isIdentifier(node: Node): node is Identifier {
    return node.kind === 80;
}
function ident(text: string): Node {
    const made: Identifier = { kind: 80, flags: 0, text };
    return made;
}
function keyword(): Node {
    const made: Keyword = { kind: 100, flags: 0 };
    return made;
}
const same = (node: Node): Node => node;
const renamed = (node: Node): Node => isIdentifier(node) ? ident(node.text + "2") : node;
const list = createNodeArray([ident("a"), ident("b")]);
function maybe(give: boolean): NodeArray<Node> | undefined {
    return give ? list : undefined;
}
function texts(nodes: NodeArray<Identifier>): string {
    let joined = "";
    for (const node of nodes) {
        joined += node.text + ";";
    }
    return joined;
}
`

// The calls every road takes: overload 1's unchanged arm (return nodes), its changed arm (return
// updatedArray), its undefined arm through an argument typed NodeArray<Node> | undefined, both ways, and
// overload 2's boolean call.
const visitNodesCalls = `console.log(texts(visitNodes(list, same, isIdentifier)));
console.log(texts(visitNodes(list, renamed, isIdentifier)));
const absent = visitNodes(maybe(false), same, isIdentifier);
console.log(absent === undefined ? "none" : texts(absent));
const present = visitNodes(maybe(true), same, isIdentifier);
console.log(present === undefined ? "none" : texts(present));
console.log(String(visitNodes(list, same).length));
`

// What Node prints for visitNodesCalls.
const visitNodesPrinted = "a;b;\na2;b2;\nnone\na;b;\n2\n"

// visitNodesProgram is the shape with implementation as visitNodes' body, then tail.
func visitNodesProgram(implementation, tail string) string {
	return visitNodesTypes + visitNodesHelpers + visitNodesOverloads + implementation + visitNodesValues + tail
}

// The words the census says at a call to overload 1 when the body isn't the certificate's.
const visitNodesResultRefused = "overload 1 of visitNodes result NodeArray<TOut> | (TInArray & undefined) cannot be served by implementation result NodeArray<Node> | undefined"

// visitNodesNatives builds program natively, unsanitized and under the sanitizers, runs each, and fails t
// unless each exits 0 printing want: the natives read the result as the implementation built it.
func visitNodesNatives(t *testing.T, program *ir.Program, want string) {
	t.Helper()
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "native")
		if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(binary)
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
		if code != 0 || out.String() != want {
			t.Fatalf("native (sanitized %v): exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", sanitize, code, out.String(), errs.String(), want)
		}
	}
}

// visitNodesStops fails t unless source stops lowering, NotYet or Refused, with words in what it says.
func visitNodesStops(t *testing.T, source, words string) {
	t.Helper()
	_, err := lowerSourceAs(t, source, ".ts")
	what := ""
	var notYet *NotYet
	var refused *Refused
	switch {
	case errors.As(err, &notYet):
		what = notYet.What
	case errors.As(err, &refused):
		what = refused.What
	}
	if !strings.Contains(what, words) {
		t.Fatalf("want a stop saying %q, got %v", words, err)
	}
}

// The certificate's own program: every call lowers, the census serving overload 1 by the guarded-array
// certificate where main refused it for its result, and agrees with Node on the JavaScript backend and
// natively, unsanitized and under the sanitizers.
func TestVisitNodesCertificateLowers(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, visitNodesProgram(visitNodesImplementation, visitNodesCalls), ".ts")
	visitNodesNatives(t, program, visitNodesPrinted)
}

// Condition 5, the storage: each direct call's instance stores NodeArray<Node> | undefined where the call
// reads NodeArray<Identifier>, and the two are held alike only through the NodeArray's own layout, which
// isArrayType never takes, so the natives read each element's text off the array as it was built. Mutant:
// overloadCertificateSameHeld's NodeArray arm dropped, and every call is NotYet at the storage condition.
func TestVisitNodesCertificateStorage(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, visitNodesProgram(visitNodesImplementation, `const kept = visitNodes(list, renamed, isIdentifier);
console.log(String(kept.length) + " " + texts(kept));
`), ".ts")
	if err != nil {
		t.Fatalf("want the direct call's storage matched, got %v", err)
	}
	visitNodesNatives(t, program, "2 a2;b2;\n")
}

// Condition 1: a writable view of a NodeArray anywhere in the program makes a visitNodes call NotYet, naming
// the view. Construction (iv-a), an array literal cast to MutableNodeArray and kept raw, is one; the program
// without it, whose one such cast createNodeArray hands straight to its NodeArray result, is
// TestVisitNodesCertificateLowers. A top type hop is another (Lower's must candidate on #ytagqs5): the list
// cast to unknown, object or {}, or handed to a type parameter or an unknown parameter, then cast to Node[],
// pushed or written through during a later check and after the return. Each of those stops today before the
// certificate is asked, at view_arrays.go's "a cast that makes a viewed readonly array writable" (an unknown,
// object or ReadonlyArray source cast to a writable array) or, for {}, castRefusal's unchecked cast (probes
// J-v8b36r7 and J-3ffrmk8), so those rows pin those words and it's a latent should. The scan reads such a
// source as one that may hold a NodeArray all the same (overloadGuardedMayHoldNodeArray), which the two
// rows with the call ahead of the hop show by its words. Mutants: WritableView, the scan answering nothing,
// and the (iv-a) row lowers; MayHold, today's reading restored, and the two call-first rows stop at
// view_arrays.go instead, a red of the scan's reading only, since that stop still holds them.
func TestVisitNodesCertificateWritableViewStops(t *testing.T) {
	t.Parallel()
	const viewed = "a cast that makes a viewed readonly array writable"
	const scanned = overloadGuardedArrayWords + "in a program holding a writable view of a NodeArray at "
	const later = `let calls = 0;
function isIdentifierPoking(node: Node): node is Identifier {
    calls++;
    const answer = node.kind === 80;
    if (calls === 4) {
        DURING;
    }
    return answer;
}
const checked = visitNodes(list, same, isIdentifierPoking);
AFTER;
console.log(String(checked.length) + " " + texts(checked));
`
	hop := func(setup, during, after string) string {
		return setup + strings.Replace(strings.Replace(later, "DURING", during, 1), "AFTER", after, 1)
	}
	const callFirst = `const checked = visitNodes(list, same, isIdentifier);
const kept = list as HOP;
const raw = kept as Node[];
raw.push(keyword());
console.log(String(checked.length) + " " + texts(checked));
`
	for _, row := range []struct{ name, tail, words string }{
		{"(iv-a) kept raw", "const raw = [ident(\"c\")] as MutableNodeArray<Node>;\nconsole.log(String(raw.length));\n" + visitNodesCalls, scanned},
		{"(a) through unknown", hop("const kept = list as unknown;\nconst raw = kept as Node[];\n", "raw[0] = keyword()", "raw.push(keyword())"), viewed},
		{"(b) through object", hop("const kept = list as object;\nconst raw = kept as Node[];\n", "raw[0] = keyword()", "raw.push(keyword())"), viewed},
		{"(c) through {}", hop("const kept = list as {};\nconst raw = kept as Node[];\n", "raw[0] = keyword()", "raw.push(keyword())"), "a cast the runtime can't check"},
		{"(d) through a type parameter", hop("function poke<T>(x: T): void {\n    (x as unknown as Node[]).push(keyword());\n}\n", "poke(list)", "poke(list)"), viewed},
		{"(d) through a type parameter over readonly Node[]", hop("function poke<T extends readonly Node[]>(x: T): void {\n    (x as unknown as Node[]).push(keyword());\n}\n", "poke(list)", "poke(list)"), viewed},
		{"(e) through an unknown parameter", hop("function poke(x: unknown): void {\n    (x as Node[]).push(keyword());\n}\n", "poke(list)", "poke(list)"), viewed},
		{"the call ahead of a hop through unknown", strings.Replace(callFirst, "HOP", "unknown", 1), scanned},
		{"the call ahead of a hop through object", strings.Replace(callFirst, "HOP", "object", 1), scanned},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			visitNodesStops(t, visitNodesProgram(visitNodesImplementation, row.tail), row.words)
		})
	}
}

// Condition 2: a test that lies, true for a keyword, is a checked predicate in .ts, passed as its checked
// forwarder, so the lie stops at its own check the first time it says yes to the keyword, inside the
// worker, on the JavaScript backend and natively alike, before anything prints; Node prints the keyword's
// missing text. No guard of the certificate's is the only thing holding this row: predicateArguments and
// the forwarder hold it, so it has no mutant here.
func TestVisitNodesCertificateLyingTestStopsAtItsCheck(t *testing.T) {
	t.Parallel()
	source := visitNodesProgram(visitNodesImplementation, `function isIdentifierLie(node: Node): node is Identifier {
    return node.flags === 0;
}
const mixed = createNodeArray([ident("a"), keyword()]);
console.log(texts(visitNodes(mixed, same, isIdentifierLie)));
`)
	path := writeAgreementSource(t, "main.ts", source)
	node := runAgreementNode(t, path)
	if node.code != 0 || string(node.stdout) != "a;undefined;\n" {
		t.Fatalf("source Node: exit %d stdout %q; want exit 0 and the lie read", node.code, node.stdout)
	}
	program := lowerAgreementEntry(t, path)
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	backend := runAgreementNode(t, generated)
	if backend.code != 70 || string(backend.stdout) != "" || !strings.HasPrefix(string(backend.stderr), "adamic: panic: predicate ") {
		t.Fatalf("JavaScript backend: exit %d stdout %q stderr %q; want exit 70 at the predicate's own check, before anything prints", backend.code, backend.stdout, backend.stderr)
	}
	code, out, errs := brandedNative(t, program)
	if code != 70 || out != "" || !bytes.Equal([]byte(errs), backend.stderr) {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 70 and the JavaScript backend's %q", code, out, errs, backend.stderr)
	}
}

// Condition 3: the loop is exactly the statement before `return R`, over that R, which nothing writes.
// Each row's body breaks one part, and overload 1 stays refused for its result at its call, as on main.
// Mutant: the guard read from any earlier statement of the block, not the one right before, and "a
// statement between" lowers.
func TestVisitNodesCertificateLoopShape(t *testing.T) {
	t.Parallel()
	const guard = `        if (test !== undefined) {
            for (const testedNode of updatedArray) {
                if (!test(testedNode)) Debug.fail("False expression: Unexpected node.\r\nVerbose Debug Information: Node array did not pass test.");
            }
        }
        return updatedArray;`
	for _, row := range []struct{ name, replacement string }{
		{"a statement between", strings.Replace(guard, "        return updatedArray;", "        console.log(\"between\");\n        return updatedArray;", 1)},
		{"a label on the loop", strings.Replace(guard, "            for (const testedNode", "            checked: for (const testedNode", 1)},
		{"a second statement in the loop", strings.Replace(guard, "Node array did not pass test.\");\n", "Node array did not pass test.\");\n                console.log(\"checked\");\n", 1)},
		{"the loop over another array", strings.Replace(guard, "for (const testedNode of updatedArray)", "for (const testedNode of nodes)", 1)},
		{"the returned array written", strings.Replace(strings.Replace(guard, "updatedArray", "rebuilt", -1), "        if (test !== undefined) {", "        let rebuilt = updatedArray;\n        rebuilt = createNodeArray(updated);\n        if (test !== undefined) {", 1)},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			implementation := strings.Replace(visitNodesImplementation, guard, row.replacement, 1)
			if implementation == visitNodesImplementation {
				t.Fatal("the row changed nothing")
			}
			visitNodesStops(t, visitNodesProgram(implementation, visitNodesCalls), visitNodesResultRefused)
		})
	}
}

// Condition 6: `TInArray & undefined` is served only by the array parameter, narrowed to undefined, as the
// first statement under `nodes === undefined`. A `return undefined` where nodes is present is the lie the
// call typed NodeArray<Identifier> would read; under `nodes === undefined` it's honest, and still not the arm.
// Mutant: any return typed undefined taken as the arm, and both `return undefined` rows lower.
func TestVisitNodesCertificateUndefinedArm(t *testing.T) {
	t.Parallel()
	const arm = `    if (nodes === undefined) {
        return nodes;
    }
`
	for _, row := range []struct{ name, replacement string }{
		{"undefined returned where nodes is present", arm + "    if (nodes.length === 0) {\n        return undefined;\n    }\n"},
		{"undefined returned under the test", "    if (nodes === undefined) {\n        return undefined;\n    }\n"},
		{"a statement before the return", "    if (nodes === undefined) {\n        console.log(\"none\");\n        return nodes;\n    }\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			implementation := strings.Replace(visitNodesImplementation, arm, row.replacement, 1)
			if implementation == visitNodesImplementation {
				t.Fatal("the row changed nothing")
			}
			visitNodesStops(t, visitNodesProgram(implementation, visitNodesCalls), visitNodesResultRefused)
		})
	}
}

// Condition 4: a getter on kind answers each read afresh, so the yes a check read says nothing about the
// array the call gets: a class whose kind reads 80 twice and then 100. The call is NotYet, naming the getter.
// It has no mutant: with the getter scan dropped the program still stops, at "a getter in a checked field
// contract" (refuseReachedContractGetters, class_accessors.go), since the test's proof registers the view
// a getter reaches (probe J-mh34cmy), so the scan isn't the only guard of this row.
func TestVisitNodesCertificateGetterStops(t *testing.T) {
	t.Parallel()
	visitNodesStops(t, visitNodesProgram(visitNodesImplementation, `class Flipping implements Node {
    readonly flags = 0;
    reads = 0;
    get kind(): number {
        this.reads++;
        return this.reads <= 2 ? 80 : 100;
    }
}
const flipping: Node = new Flipping();
const flipped = createNodeArray([flipping, ident("b")]);
console.log(String(visitNodes(flipped, same, isIdentifier).length));
`), overloadGuardedArrayWords+"whose test's claim reads kind, a getter's name at ")
}

// Condition 7 (Language's must 1): a test that splices the caller's NodeArray with
// Array.prototype.splice.call puts a keyword in it with no writable view anywhere. The certificate builds no
// scan of its own for this: #3cee4gb's refusal of a delayed mutator on a readonly receiver
// (delayedReadonlyReceiver, library_method_values.go) stops it program-wide, as Adamic ruled. Mutant: that
// refusal dropped, and the row lowers.
func TestVisitNodesCertificateDelayedMutatorStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, receiver string }{
		{"on the NodeArray", "list"},
		{"through a readonly Node[] view", "view"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			visitNodesStops(t, visitNodesProgram(visitNodesImplementation, `const view: readonly Node[] = list;
let spliced = false;
function isIdentifierSplicing(node: Node): node is Identifier {
    if (!spliced) {
        spliced = true;
        Array.prototype.splice.call(`+row.receiver+`, 0, 1, keyword());
    }
    return node.kind === 80;
}
console.log(String(view.length) + " " + texts(visitNodes(list, same, isIdentifierSplicing)));
`), "a delayed Array.prototype.splice on a readonly array")
		})
	}
}

// Condition 8 (Language's must 2): membership read from a field some view can write is a moment's answer.
// Language's program reads flags; a claim Adamic can check is read from its kind tag (predicateMembership),
// so here kind is the field, and the test, on its fourth call, writes the first element's kind through a
// writable view after that element's own check said yes. Node keeps kind readonly while the program holds
// the first element as Settable too, a writable view beside it, and the call is NotYet, naming the field.
// Where Node itself declares kind writable, the invariance pass refuses Identifier seen as Node first
// (adamic/invariant-mutable, probe J-mh34cmy), so that row pins its words and no mutant reaches it.
// Mutant: the readonly test dropped, and the writable view row lowers.
func TestVisitNodesCertificateWritableMembershipStops(t *testing.T) {
	t.Parallel()
	writableNode := strings.Replace(visitNodesTypes, "    readonly kind: number;\n    readonly flags: number;\n}\ninterface Identifier", "    kind: number;\n    readonly flags: number;\n}\ninterface Identifier", 1)
	if writableNode == visitNodesTypes {
		t.Fatal("the types changed nothing")
	}
	const tail = `let calls = 0;
function isIdentifierWriting(node: Node): node is Identifier {
    calls++;
    const answer = node.kind === 80;
    if (calls === 4) {
        WRITE;
    }
    return answer;
}
const same = (node: Node): Node => node;
const checked = visitNodes(pair, same, isIdentifierWriting);
console.log(String(checked.length));
`
	for _, row := range []struct{ name, types, values, write, words string }{
		{"kind writable on Node", writableNode, `const first: Node = { kind: 80, flags: 0 };
const second: Node = { kind: 80, flags: 0 };
const pair = createNodeArray([first, second]);
`, "first.kind = 100", "whose readonly field kind becomes writable"},
		{"a writable view of the element", visitNodesTypes, `interface Settable {
    kind: number;
    flags: number;
}
const settable: Settable = { kind: 80, flags: 0 };
const first: Node = settable;
const second: Node = { kind: 80, flags: 0 };
const pair = createNodeArray([first, second]);
`, "settable.kind = 100", overloadGuardedArrayWords + "whose test's claim reads kind, writable in a view the program holds at "},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			source := row.types + visitNodesHelpers + visitNodesOverloads + visitNodesImplementation + row.values + strings.Replace(tail, "WRITE", row.write, 1)
			visitNodesStops(t, source, row.words)
		})
	}
}
