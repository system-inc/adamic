package lower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// A soundness probe for a proposed enum-tag rule, never for landing. Today a refinement of B | E, E's
// kind the whole enum K, is refused (enumTagHazardOf's open hazard). The rule would admit it when no
// declared type assignable to E has kind B. The hinge: a value built at an open-kind type, kind B
// chosen at run time, E's layout, enters the union by a legal road and a narrowing hands it back as
// B. Do Adamic's reads on it stay checked (exit 70 at the read) or read B's layout raw?
//
// TestEnumTagHinge reports and never asserts an outcome: one HINGE line per program, the same run
// at main and under the mutant (scratchpad mutant-hazard.diff, enumTagHazardOf reporting no open
// hazard to any caller; mutant-hazard-walk-only.diff lifts it at enumRefusal's read only, so the
// predicate provers still see it). reads(f)=total/viewed/checked counts every ir.Property read of
// the named fields: a view read (View set) and a read CheckedRead calls checked.
//
// The comparison form is expected not to load: the checker narrows `v.kind === K.B` by filtering
// members whose kind is comparable to K.B, and E's K is, so the true side stays B | E and `v.b` is
// TS2339 (narrowTypeByDiscriminant). Only a predicate's declared type hands back B alone. Road f
// is the one a comparison does narrow by: a write through a wider alias puts kind A in a B, and
// `v.kind === K.A` removes B, handing the B back as E.
const hingePrelude = `enum K { A = 1, B = 2, C = 3 }
interface B { readonly kind: K.B; readonly b: string }
interface E { readonly kind: K; readonly e: number }
function kb(): K { return K.B; }
function yes(): boolean { return kb() === K.B; }
function isB(n: B | E): n is B { return n.kind === K.B; }
const honest: B = { kind: K.B, b: 'honest' };
`

const hingeRaw = "class Raw { constructor(readonly kind: K, readonly e: number) {} }\n"

// hingeRead is the narrowing and the read of B's field on reference, as a string, in one form.
func hingeRead(form, reference string) string {
	if form == "predicate" {
		return "(isB(" + reference + ") ? " + reference + ".b : 'other')"
	}
	return "(" + reference + ".kind === K.B ? " + reference + ".b : 'other')"
}

type hingeProgram struct {
	name   string
	source string
	fields []string
}

func hingePrograms() []hingeProgram {
	programs := []hingeProgram{}
	for _, form := range []string{"predicate", "comparison"} {
		read := func(reference string) string { return hingeRead(form, reference) }
		show := "function show(v: B | E): string { return " + read("v") + "; }\n"
		lie := "const lie: E = { kind: kb(), e: 1 };\n"
		for _, road := range []struct{ name, source string }{
			// a, direct. Assigned alone, the checker narrows v to E by the assignment, so the predicate
			// leaves E & B; either way keeps the declared B | E at the narrowing.
			{"a-direct-assigned", "const v: B | E = { kind: kb(), e: 1 };\nconsole.log(" + read("v") + ");\n"},
			{"a-direct-either", "const v: B | E = yes() ? { kind: kb(), e: 1 } : honest;\nconsole.log(" + read("v") + ");\n"},
			// b, a class whose kind comes from its constructor.
			{"b-class-assigned", hingeRaw + "const v: B | E = new Raw(K.B, 1);\nconsole.log(" + read("v") + ");\n"},
			{"b-class-either", hingeRaw + "const v: B | E = yes() ? new Raw(K.B, 1) : honest;\nconsole.log(" + read("v") + ");\n"},
			// c, typed E through a checked predicate whose body checks the fields: as the brief writes it,
			// over object, and over a union of shapes, in case object's `in` doesn't lower.
			{"c-predicate-object", "function isE(o: object): o is E { return 'kind' in o && 'e' in o && typeof o.kind === 'number' && typeof o.e === 'number'; }\nconst raw: object = { kind: kb(), e: 1 };\nconst v: B | E = isE(raw) ? raw : honest;\nconsole.log(" + read("v") + ");\n"},
			{"c-predicate-union", "interface O { readonly o: string }\nfunction isE(n: E | O): n is E { return 'e' in n; }\nconst raw: E | O = yes() ? { kind: kb(), e: 1 } : { o: 'o' };\nconst v: B | E = isE(raw) ? raw : honest;\nconsole.log(" + read("v") + ");\n"},
			// d, the narrowing in another function from the build: a parameter, a field, and b's class.
			{"d-parameter", show + lie + "console.log(show(lie));\n"},
			{"d-field", "interface Holder { readonly v: B | E }\nfunction showHeld(h: Holder): string { return " + read("h.v") + "; }\n" + lie + "const held: Holder = { v: lie };\nconsole.log(showHeld(held));\n"},
			{"d-parameter-class", hingeRaw + show + "console.log(show(new Raw(K.B, 1)));\n"},
			// e, the honest control: Node prints honest, then other.
			{"e-honest", show + "console.log(show(honest));\nconsole.log(show({ kind: K.A, e: 1 }));\n"},
		} {
			programs = append(programs, hingeProgram{road.name + "/" + form, hingePrelude + road.source, []string{"b"}})
		}
		// f, beyond the brief's list: readonly doesn't hold assignability, so a B seen as { kind: K }
		// takes kind A, and the false side of isB, or `v.kind === K.A`, hands the B back as E. Node
		// prints undefined; a raw read takes e from b's place. The other side stays B | E for the
		// comparison, so neither side reads b there.
		widen := "(!isB(v) ? String(v.e) : 'b')"
		if form == "comparison" {
			widen = "(v.kind === K.A ? String(v.e) : 'b')"
		}
		programs = append(programs, hingeProgram{"f-alias-write/" + form, hingePrelude +
			"const b0: B = { kind: K.B, b: 'bee' };\nconst w: { kind: K } = b0;\nw.kind = K.A;\nfunction showE(v: B | E): string { return " + widen + "; }\nconsole.log(showE(b0));\n", []string{"e"}})
	}
	programs = append(programs, hingeJsxText()...)
	return programs
}

// hingeJsxText is tsc's JsxText, shaped small. JsxText extends LiteralLikeNode, not
// LiteralExpression, and LiteralExpression declares no kind of its own (Node's whole SyntaxKind),
// yet isLiteralExpression is isLiteralKind(node.kind), FirstLiteralToken..LastLiteralToken, which
// takes JsxText in (types.ts:477, utilitiesPublic.ts:1494 and :1498 in cohere's fixtures). So Lit's
// kind is the whole SK, no declared type assignable to Lit has kind Jsx (J lacks brand), and isLit
// still hands a J, J's layout and kind Jsx, to code that holds a Lit.
//
// Expected under the rule: its condition holds for Lit | J by kind Jsx, so the narrowing is admitted
// (as under the mutant), and the program is sound only because R-T80 checks isLit at its call: its
// body is a range of kinds, which proves no layout, so in .ts the call checks j against Lit, finds
// no brand and stops with exit 70 before any narrowing, while Node prints true. In "twin" the
// union's Jsx member has another layout than the value: if isLit were trusted, the narrowing would
// read attributes off a J raw. In .a isLit is unproven, so it's refused there.
func hingeJsxText() []hingeProgram {
	const types = `enum SK { Num = 1, Str = 2, Jsx = 3, Id = 4 }
interface Syntax { readonly kind: SK }
interface Lit extends Syntax { readonly text: string; readonly brand: number }
interface J extends Syntax { readonly kind: SK.Jsx; readonly text: string; readonly only: boolean }
interface JA extends Syntax { readonly kind: SK.Jsx; readonly attributes: number }
function isLit(n: Syntax): n is Lit { return n.kind >= SK.Num && n.kind <= SK.Jsx; }
const j: J = { kind: SK.Jsx, text: 'jsx', only: true };
const n: Syntax = j;
`
	programs := []hingeProgram{}
	for _, form := range []string{"predicate", "comparison"} {
		same := "(isJ(x) ? String(x.only) : String(x.brand))"
		twin := "(isJA(x) ? String(x.attributes) : String(x.brand))"
		if form == "comparison" {
			same = "(x.kind === SK.Jsx ? String(x.only) : String(x.brand))"
			twin = "(x.kind === SK.Jsx ? String(x.attributes) : String(x.brand))"
		}
		programs = append(programs,
			hingeProgram{"jsx-text/" + form, types + "function isJ(x: Lit | J): x is J { return x.kind === SK.Jsx; }\nfunction show(x: Lit | J): string { return " + same + "; }\nconsole.log(isLit(n) ? show(n) : 'not a literal');\n", []string{"only", "brand"}},
			hingeProgram{"jsx-twin/" + form, types + "function isJA(x: Lit | JA): x is JA { return x.kind === SK.Jsx; }\nfunction show(x: Lit | JA): string { return " + twin + "; }\nconsole.log(isLit(n) ? show(n) : 'not a literal');\n", []string{"attributes", "brand"}},
		)
	}
	return programs
}

func TestEnumTagHinge(t *testing.T) {
	t.Parallel()
	for _, probe := range hingePrograms() {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, "main.ts", probe.source)
			program, lowered := hingeLower(path)
			_, loweredA := hingeLower(writeAgreementSource(t, "main.a", probe.source))
			source := runAgreementNode(t, path)
			script, binary, reads := "-", "-", "-"
			if program != nil {
				generated := filepath.Join(t.TempDir(), "generated.mjs")
				if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
					t.Fatal(err)
				}
				ran := runAgreementNode(t, generated)
				script = hingeObserved(ran.code, string(ran.stdout), string(ran.stderr))
				binary = hingeNative(t, program)
				reads = hingeReads(program, probe.fields)
			}
			t.Logf("HINGE %s lower=%s a=%s node=%s js=%s native=%s reads(%s)=%s", probe.name, lowered, loweredA,
				hingeObserved(source.code, string(source.stdout), string(source.stderr)), script, binary, strings.Join(probe.fields, ","), reads)
		})
	}
}

// hingeLower loads and lowers the file at path, saying what happened: ok, the checker's first
// diagnostic, or the lowering's refusal. Either failure is an outcome here, not the harness's.
func hingeLower(path string) (*ir.Program, string) {
	checked, err := load.Load([]string{path})
	if err != nil {
		return nil, "load: " + hingeFirstLine(err.Error())
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		return nil, "refused: " + hingeFirstLine(err.Error())
	}
	return program, "ok"
}

// hingeNative is brandedNative, its build failure and a hang reported rather than fatal.
func hingeNative(t *testing.T, program *ir.Program) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		return "build: " + hingeFirstLine(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, errs bytes.Buffer
	command.Stdout, command.Stderr = &out, &errs
	code := 0
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return "timeout"
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return hingeObserved(code, out.String(), errs.String())
}

// hingeReads counts the property reads of fields, outside a callee: all of them, those through a view,
// and those CheckedRead says are checked.
func hingeReads(program *ir.Program, fields []string) string {
	total, viewed, checked := 0, 0, 0
	inspect := func(node any) bool {
		if read, ok := node.(ir.Property); ok && !read.Method && slices.Contains(fields, read.Name) {
			total++
			if read.View != "" {
				viewed++
			}
			if program.CheckedRead(read) {
				checked++
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return fmt.Sprintf("%d/%d/%d", total, viewed, checked)
}

func hingeObserved(code int, stdout, stderr string) string {
	return fmt.Sprintf("%d,%q,%q", code, stdout, hingeFirstLine(stderr))
}

func hingeFirstLine(text string) string {
	return strings.SplitN(text, "\n", 2)[0]
}
