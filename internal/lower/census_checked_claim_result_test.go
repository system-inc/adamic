package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// The census serves an overload whose result is a function claiming `p is R` from an implementation
// whose result relates with the claim erased to boolean, in .ts, where every call through the promised
// signature checks the claim (censusCheckedClaimResult, Adamic's ruling on #1w5bve2).

// censusClaimNodes is a node family in tsc's shape, with findAncestor's callback a checked claim.
const censusClaimNodes = `interface N { readonly kind: number; readonly parent: N | undefined; }
interface A extends N { readonly kind: 1; readonly a: string; }
interface B extends N { readonly kind: 2; readonly b: string; }
interface C extends N { readonly kind: 3; readonly c: string; }
function isA(n: N): n is A { return n.kind === 1; }
export function findAncestor<T extends N>(node: N | undefined, callback: (element: N) => element is T): T | undefined {
  while (node !== undefined) {
    if (callback(node)) { return node; }
    node = node.parent;
  }
  return undefined;
}
`

// censusClaimOr is or with an implementation whose binders rename the overload's by position, holding
// its predicates in a rest of plain callbacks, as Optimizer's by-position or on #1w5bve2 does.
const censusClaimOr = `export function or<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2;
export function or<P, R1 extends P, R2 extends P>(...fs: ((p: P) => boolean)[]): (p: P) => boolean {
    return (p: P): boolean => {
        for (const f of fs) {
            if (f(p)) {
                return true;
            }
        }
        return false;
    };
}
`

const censusClaimHonestB = "function isB(n: N): n is B { return n.kind === 2; }\n"

const censusClaimCalls = `const root: C = { kind: 3, parent: undefined, c: "root" };
const middle: B = { kind: 2, parent: root, b: "middle" };
const leaf: N = { kind: 9, parent: middle };
const two = findAncestor(leaf, or(isA, isB));
console.log(two === undefined ? "none" : String(two.kind));
const none = findAncestor(root, or(isA, isB));
console.log(none === undefined ? "none" : String(none.kind));
`

// censusClaimRuns builds program natively under the sanitizers and as JavaScript, and runs both.
func censusClaimRuns(t *testing.T, source, extension string) []nodeObservation {
	t.Helper()
	program, err := lowerSourceAs(t, source, extension)
	if err != nil {
		t.Fatalf("want it lowered, got %v", err)
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
	observations := []nodeObservation{}
	for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		code := 0
		if err := command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		observations = append(observations, nodeObservation{stdout.Bytes(), stderr.Bytes(), code})
	}
	return observations
}

// or by position, honest, lowers and prints what Node prints on both backends.
func TestCensusCheckedClaimResultAgrees(t *testing.T) {
	t.Parallel()
	source := censusClaimNodes + censusClaimHonestB + censusClaimOr + censusClaimCalls
	want := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
	if want.code != 0 || string(want.stdout) != "2\nnone\n" {
		t.Fatalf("Node: exit %d, stdout %q; want 0 and %q", want.code, want.stdout, "2\nnone\n")
	}
	for index, got := range censusClaimRuns(t, source, ".ts") {
		if got.code != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("backend %d: exit %d, stdout %q, stderr %q; want Node's %q", index, got.code, got.stdout, got.stderr, want.stdout)
		}
	}
}

// A lie through or (isB says yes to a kind-3 node, then read as a B) stops at the checked call with
// the same words natively and on JavaScript, where Node reads the lie as undefined.
func TestCensusCheckedClaimResultLieStops(t *testing.T) {
	t.Parallel()
	source := censusClaimNodes + "function isB(n: N): n is B { return n.kind === 2 || n.kind === 3; }\n" + censusClaimOr + `const root: C = { kind: 3, parent: undefined, c: "root" };
const leaf: N = { kind: 9, parent: root };
const found = findAncestor(leaf, or(isA, isB));
if (found !== undefined && found.kind !== 1) { console.log(found.b); }
`
	if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); node.code != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node: exit %d, stdout %q; want 0 and the lie read as %q", node.code, node.stdout, "undefined\n")
	}
	runs := censusClaimRuns(t, source, ".ts")
	for index, got := range runs {
		t.Logf("backend %d: exit %d, stderr %q", index, got.code, got.stderr)
		if got.code != 70 || len(got.stdout) != 0 || !strings.HasPrefix(string(got.stderr), "adamic: panic: predicate ") {
			t.Fatalf("backend %d: exit %d, stdout %q, stderr %q; want 70, nothing printed and a predicate check's words", index, got.code, got.stdout, got.stderr)
		}
	}
	if !bytes.Equal(runs[0].stderr, runs[1].stderr) {
		t.Fatalf("native stopped with %q and JavaScript with %q; want the same words", runs[0].stderr, runs[1].stderr)
	}
}

// Each of these keeps the census's result refusal: .a, where no claim is checked, and a promised
// intersection, which is more than the one call signature the erasure reads. The census judges the
// declaration, so neither twin calls through its result. An asserts result isn't here: the census's
// ordinary relation serves it before this rule is asked (Language's a01 and a02 on #1w5bve2).
func TestCensusCheckedClaimResultStaysRefused(t *testing.T) {
	t.Parallel()
	tagged := `export function tag<P, R extends P>(f1: (p1: P) => p1 is R): ((p: P) => p is R) & { tag: string };
export function tag<P, R extends P>(...fs: ((p: P) => boolean)[]): ((p: P) => boolean) & { tag: string } {
    return Object.assign((p: P): boolean => fs.length > 0 && fs[0](p), { tag: "t" });
}
const tagB = tag(isB);
console.log(tagB === undefined ? "none" : "made");
`
	for _, probe := range []struct{ name, source, extension string }{
		// or's declarations alone, so nothing but the census judges them before a predicate value is.
		{"a", censusClaimOr + "console.log(\"made\");\n", ".a"},
		{"intersection", censusClaimNodes + censusClaimHonestB + tagged, ".ts"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			t.Logf("%s stops at: %v", probe.name, err)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "cannot be served by implementation result") {
				t.Fatalf("want the census's result refusal, got %v", err)
			}
		})
	}
}
