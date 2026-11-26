package lower

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
)

// The swing (#x2cjvy9, Adamic's ruling of Nov 25) moved 134 rows of lower's verbose run, each a .ts
// program its test held refused as an object refinement using an open numeric enum as a literal tag.
// Their premise moved to .a, where rulings.md:258 still refuses; this file holds their .ts side. Each
// row here is the program its test lowered in .ts, built from the same shared constants and the same
// concatenation, so it's byte for byte that program. The constants below are the tests' own locals,
// copied, since a test's local can't be named from here.

func init() { enumTagSwingSites = &enumTagSwingRecord{} }

// enumTagSwungAt is every site in the program whose entry file is path where a .ts refinement took the
// swing's checked view in place of family 1's refusal (enumTagSwingSites), sorted. Each test writes its
// program in a directory of its own, so the sites are that program's alone.
func enumTagSwungAt(path string) []string {
	enumTagSwingSites.mutex.Lock()
	defer enumTagSwingSites.mutex.Unlock()
	sites := slices.Clone(enumTagSwingSites.sites[filepath.Dir(path)])
	slices.Sort(sites)
	return sites
}

// lowerSourceSwung lowers source as main.ts and gives the sites where the swing took a refinement.
func lowerSourceSwung(t *testing.T, source string) (*ir.Program, []string, error) {
	t.Helper()
	path := writeAgreementSource(t, "main.ts", source)
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := Lower(context.Background(), loaded)
	return program, enumTagSwungAt(path), err
}

// swungAtSuffix says whether one of sites ends with the given file, line and column, as a Where does.
func swungAtSuffix(sites []string, suffix string) bool {
	return slices.ContainsFunc(sites, func(site string) bool { return strings.HasSuffix(site, suffix) })
}

// TestEnumTagAccessRefusals' reset.
const swingAccessReset = "function reset(m: M): void { m.n = b; }\n"

// TestEnumTagRuntimeRefinementPosition's read and wide.
const swingRuntimeRead = "function read(value: X): Tagged { if (value.kind === Kind.B) { return value.literal; } return value.literal; }\n"
const swingRuntimeWide = "function numeric(): number { return 1; }\nconst wide: X = { kind: numeric(), literal: { kind: Kind.A, value: 1 }, name: { kind: Kind.A, value: 2 } };\n"

// TestKindPredicateMatchesTheWrittenComparison's unionTag.
const swingUnionTag = `enum K { A = 1, B = 2, C = 3 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
interface C { readonly kind: K.B | K.C; readonly c: string }
type N = A | B | C;
`

// TestKindPredicateWalkRefusals' show.
const swingWalkShow = "function show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"

// TestKindAliasFallbacksKeepFamilyOne's getters and cases.
const swingAliasGetters = `enum K { A = 1, B = 2 }
enum F { X = 1 }
class GA { get kind(): K.A { return K.A; } readonly flag: F = F.X; readonly a: number = 1 }
class GB { get kind(): K.B { return K.B; } readonly b: string = 'x' }
type N = GA | GB;
const a = new GA();
const b = new GB();
`
const swingAliasCases = "switch (kind) { case K.A: return `a ${n.a}`; default: return 'b'; }"

// TestClosedKindLiteralKindBuildsStayOpen's types and annotated.
const swingLiteralBuildTypes = "enum Kind { Block, Ident }\ninterface Node { readonly kind: Kind }\ninterface Block extends Node { readonly kind: Kind.Block; count: number }\ninterface Expr extends Node { text: string }\ninterface Ident extends Expr { readonly kind: Kind.Ident }\n" +
	"function isBlock(node: Node): node is Block { return node.kind === Kind.Block; }\n" +
	"function measure(body: Block | Expr): string { try { return String(isBlock(body) ? body.count.toFixed(1).length : body.text.length); } catch (error) { return 'threw'; } }\n" +
	"const block: Block = { kind: Kind.Block, count: 7 };\nconst ident: Ident = { kind: Kind.Ident, text: 'ab' };\n"
const swingLiteralBuildAnnotated = "{ readonly kind: Kind.Block; readonly text: string }"

// TestClosedKindRunTimeKindStaysOpen's open.
const swingRunTimeOpen = "interface Open extends Node { readonly n: number }\n" +
	"function isOpenBody(x: Body | Open): x is Body { return x.kind === K.Body; }\n" +
	"function show(x: Body | Open): string { return isOpenBody(x) ? x.b : 'open'; }\n" +
	"function kb(): K { return K.Ident; }\n"

// TestClosedKindHoldingTheComparedKindRefuses' show.
const swingHoldingShow = closedKindIsBody + "function show(x: Body | Ex): string { return isBody(x) ? x.b : 'expression'; }\nconsole.log(show(body));\n"

// TestClosedKindFunctionConstructorNarrowing's narrowing.
const swingConstructorNarrowing = "function isOpenBody(x: Body | Open): x is Body { return x.kind === K.Body; }\n" +
	"function show(x: Body | Open): string { return isOpenBody(x) ? x.b : 'open'; }\n" +
	"console.log(show(body));\nconsole.log(show(open));\n"

// TestKindSetClosedSourcesFitEveryKind's predicate.
const swingFitPredicate = "function isSelected(n:Node):n is Ident|NumLit { return n.kind===K.Ident || n.kind===K.Num; }\n" +
	"function show(n:Body|Ex):string { if(isSelected(n)) { if(n.kind===K.Ident) return n.name; return n.text; } return 'other'; }\n"

// TestKindSetUnionClaimFitsEverySource's base.
const swingUnionClaimBase = `enum K { Body=1, A=2, B=3 }
interface Node { readonly kind: K; readonly pos: number }
interface Body extends Node { readonly kind: K.Body; readonly b0: string }
interface Ex extends Node { readonly _exBrand: undefined }
interface A extends Ex { readonly kind: K.A; readonly a: string }
interface B extends Ex { readonly kind: K.B; readonly b: string }
function isAOrB(n:Node):n is A|B { return n.kind===K.A || n.kind===K.B; }
function read(y:A|B):string { return y.kind===K.A ? y.a : y.b; }
function show(x:Body|Ex):string { return isAOrB(x) ? read(x) : 'other'; }
const body:Body={kind:K.Body,pos:0,b0:'z'};
const a:A={kind:K.A,pos:1,_exBrand:undefined,a:'x'};
console.log(show(body)); console.log(show(a));
`

// swingKindHelperShapeSource is TestKindHelperShapesAgreeWithNode's source for a row whose
// parameter is N, which is every row but double_guard.
func swingKindHelperShapeSource(helpers, body string) string {
	return kindSetTypes + helpers + "function isAB(n:N):n is A|B {" + body + "}\n" + kindSetShow + kindSetValues
}

// swingKindHelperFallbackSource is TestKindHelperFallbacksRejectWholeBody's source for a row.
func swingKindHelperFallbackSource(name, helpers, body string) string {
	types := kindSetTypes
	if name == "h05" {
		types = strings.Replace(types, "C=3", "D=1", 1)
		types = strings.Replace(types, "readonly kind: K.C", "readonly kind: K.B", 1)
	}
	if name == "wrong_field" {
		types = strings.ReplaceAll(types, "readonly kind:", "readonly other:K; readonly kind:")
	}
	return types + helpers + "function isAB(n:N):n is A|B {" + body + "}\n" + kindSetShow
}

// swingKindHelperEmptyDefaultSource is TestKindHelperEmptyDefaultTailRoutes' source for a row it
// holds unrecognized.
func swingKindHelperEmptyDefaultSource(body string) string {
	helpers := "function leaf(k:K):boolean {return k===K.B;}\nfunction kinds(k:K):boolean {" + body + "}\n"
	types := strings.Replace(kindSetTypes, "readonly a: string", "readonly a: string; readonly c: boolean", 1)
	return types + helpers + "function isAB(n:N):n is B {return n && kinds(n.kind); }\nfunction show(n:N):string {if(isAB(n))return 'true';return String(n.c);}\nconst b:B={kind:K.B,b:7}; console.log(show(b));\n"
}

// swingKindSetGrammarSource is TestKindSetBodyGrammar's source for a body of kindSetBodies.
func swingKindSetGrammarSource(body string) string {
	return kindSetTypes + "function isAB(n:N):n is A|B { " + body + " }\n"
}

// swingKindSetFallbackSource is TestKindSetPredicateFallbacksKeepFamilyOne's source for a row.
func swingKindSetFallbackSource(name, body, extra string) string {
	types := kindSetTypes
	if name == "k04" {
		types = strings.Replace(types, "C=3", "C=3, D=1", 1)
	}
	if name == "two_fields" {
		types = strings.ReplaceAll(types, "readonly kind:", "readonly other:K; readonly kind:")
	}
	if name == "two_enums" {
		types = strings.Replace(types, "readonly kind: K.B", "readonly kind: L.B", 1)
	}
	if name == "primitive_guard" {
		return "enum K { A=1, B=2 } function isAB(n:string):n is string { " + body + " }"
	}
	return extra + types + "function isAB(n:N):n is A|B { " + body + " }\n" + kindSetShow
}

// enumTagSwingRows is every row of the 134 that one .ts source reproduces, named as Go's verbose run
// names it.
var enumTagSwingRows = []struct {
	name   string
	source string
}{
	// TestEnumTagAccessRefusals: kindPredicateTypes + accessHolder + the row's source.
	{"TestEnumTagAccessRefusals/mutable_field,_resetting_call_between", kindPredicateTypes + accessHolder + swingAccessReset + "function show(m: M): string { if (isA(m.n)) { reset(m); return `a ${m.n.a}`; } return 'b'; }\nconsole.log(show({ n: a }));\n"},
	{"TestEnumTagAccessRefusals/readonly_field,_resetting_call_between", kindPredicateTypes + accessHolder + "const raw: M = { n: a };\nfunction reset(): void { raw.n = b; }\nfunction show(h: H): string { if (isA(h.n)) { reset(); return `a ${h.n.a}`; } return 'b'; }\nconsole.log(show(raw));\n"},
	{"TestEnumTagAccessRefusals/alias_write_between", kindPredicateTypes + accessHolder + "function show(h: H, m: M): string { if (isA(h.n)) { m.n = b; return `a ${h.n.a}`; } return 'b'; }\nconst both: M = { n: a };\nconsole.log(show(both, both));\n"},
	{"TestEnumTagAccessRefusals/call_in_the_chain", kindPredicateTypes + accessHolder + "function log(): boolean { return true; }\nfunction show(h: H): string { return isA(h.n) && log() ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(ha));\n"},
	{"TestEnumTagAccessRefusals/early_return", kindPredicateTypes + accessHolder + "function show(h: H): string { if (!isA(h.n)) { return 'b'; } return `a ${h.n.a}`; }\nconsole.log(show(ha));\n"},
	{"TestEnumTagAccessRefusals/read_in_a_loop_under_the_test", kindPredicateTypes + accessHolder + swingAccessReset + "function show(m: M): string { let out = ''; if (isA(m.n)) { let i = 0; while (i < 2) { if (i === 1) { out = `a ${m.n.a}`; } reset(m); i = i + 1; } } return out; }\nconsole.log(show({ n: a }));\n"},
	{"TestEnumTagAccessRefusals/getter_of_that_name_in_the_program", kindPredicateTypes + accessHolder + "class Lazy implements H { get n(): N { return a; } }\nfunction show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(new Lazy()));\n"},
	{"TestEnumTagAccessRefusals/writing_predicate", kindPredicateTypes + accessHolder + "const holder: M = { n: a };\nfunction isAw(n: N): n is A { holder.n = b; return n.kind === K.A; }\nfunction show(h: H): string { return isAw(h.n) ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(holder));\n"},
	{"TestEnumTagAccessRefusals/call_inside_the_step", kindPredicateTypes + accessHolder + "function kindOf(): K.A { return K.A; }\nfunction show(h: H): string { return h.n.kind === kindOf() ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(ha));\n"},
	{"TestEnumTagAccessRefusals/getter_of_the_compared_field,_by_predicate", kindPredicateTypes + accessHolder + "class Odd { get kind(): K.A { return K.A; } }\nfunction show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }\nconst odd = new Odd();\nconsole.log(show(ha) + String(odd.kind));\n"},
	{"TestEnumTagAccessRefusals/getter_of_the_compared_field,_by_comparison", kindPredicateTypes + accessHolder + "class Odd { get kind(): K.A { return K.A; } }\nfunction show(h: H): string { return h.n.kind === K.A ? `a ${h.n.a}` : 'b'; }\nconst odd = new Odd();\nconsole.log(show(ha) + String(odd.kind));\n"},
	{"TestEnumTagAccessRefusals/constructor's_this", kindPredicateTypes + accessHolder + "class C { readonly n: N; readonly out: string; constructor(n: N) { this.n = n; this.out = isA(this.n) ? `a ${this.n.a}` : 'b'; } }\nconsole.log(new C(a).out);\n"},

	// TestEnumTagNullishStepRefusals: kindPredicateTypes + accessHolder + nullishStepHolder + the row's source.
	{"TestEnumTagNullishStepRefusals/premise,_call_then_predicate", kindPredicateTypes + accessHolder + nullishStepHolder + "function show(o: O): string {\n    if (!o.n) return 'none';\n    clear(o);\n    return isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},
	{"TestEnumTagNullishStepRefusals/premise,_call_then_if", kindPredicateTypes + accessHolder + nullishStepHolder + "function show(o: O): string {\n    if (!o.n) return 'none';\n    clear(o);\n    if (isA(o.n)) return `a ${o.n.a}`;\n    return 'other';\n}\nconsole.log(show({ n: a }));\n"},
	{"TestEnumTagNullishStepRefusals/premise,_nothing_between", kindPredicateTypes + accessHolder + nullishStepHolder + "function show(o: O): string {\n    if (!o.n) return 'none';\n    return isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},
	{"TestEnumTagNullishStepRefusals/null_removed,_undefined_kept", kindPredicateTypes + accessHolder + nullishStepHolder + "function show(q: Q): string {\n    return q.n !== null && q.n !== undefined && isA(q.n) ? `a ${q.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},

	// TestEnumTagRuntimeRefinementPosition, its .ts row.
	{"TestEnumTagRuntimeRefinementPosition/.ts", enumTagTypePositionTypes + swingRuntimeRead + swingRuntimeWide},

	// TestKindPredicateMatchesTheWrittenComparison.
	{"TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body", swingUnionTag + "function isBC(n: N): n is C { return n.kind === K.B; }\nfunction show(n: N): string { return isBC(n) ? `c ${n.c}` : 'other'; }\nconsole.log(show({ kind: K.C, c: 'y' }));\n"},

	// TestKindPredicateWalkRefusals: the row's types + its source.
	{"TestKindPredicateWalkRefusals/wide_kind_call", "enum K { A = 1, B = 2 }\nenum F { X = 1 }\ninterface A { readonly kind: K.A; readonly flag: F; readonly a: number }\ninterface B { readonly kind: K; readonly b: string }\ntype N = A | B;\nfunction numeric(): number { return 1; }\nconst wide: B = { kind: numeric(), b: 'wide' };\n" + "function isA(n: N): n is A { return n.kind === K.A; }\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/shared_value_call", "enum K { A = 1, Z = 5 }\nenum L { B = 1, W = 6 }\ninterface A { readonly kind: K.A; readonly a: number }\ninterface B { readonly kind: L.B; readonly b: string }\ntype N = A | B;\n" + "function isA(n: N): n is A { return n.kind === K.A; }\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, a: 1 }));\n"},
	{"TestKindPredicateWalkRefusals/shared_value_comparison", "enum K { A = 1, Z = 5 }\nenum L { B = 1, W = 6 }\ninterface A { readonly kind: K.A; readonly a: number }\ninterface B { readonly kind: L.B; readonly b: string }\ntype N = A | B;\n" + "function show(n: N): string { return n.kind === K.A ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, a: 1 }));\n"},
	{"TestKindPredicateWalkRefusals/lying_declared_type", kindPredicateTypes + "function isA(n: N): n is A { return n.kind === K.B; }\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/twin_layouts", "enum K { A = 1, B = 2 }\nenum F { X = 1 }\ninterface A { readonly kind: K.A; readonly flag: F; readonly a: number }\ninterface B { readonly kind: K.A; readonly b: string }\ntype N = A | B;\n" + "function isA(n: N): n is A { return n.kind === K.A; }\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/inferred_predicate", kindPredicateTypes + "function isA(n: N) { return n.kind === K.A; }\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/overloaded", kindPredicateTypes + "function isA(n: A): n is A;\nfunction isA(n: N): n is A;\nfunction isA(n: N): n is A { return n.kind === K.A; }\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/value_alias", kindPredicateTypes + "function isAlpha(n: N): n is A { return n.kind === K.A; }\nconst isA = isAlpha;\n" + swingWalkShow},
	{"TestKindPredicateWalkRefusals/parameter_with_a_default", kindPredicateTypes + "const fallback: N = { kind: K.A, flag: F.X, a: 0 };\nfunction isA(n: N = fallback): n is A { return n.kind === K.A; }\nfunction show(n: N | undefined): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"},
	{"TestKindPredicateWalkRefusals/optional_parameter", kindPredicateTypes + "function isA(n?: N): n is A { return n !== undefined && n.kind === K.A; }\nfunction show(n: N): string { return isA(n) ? `a ${n.a}` : 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"},
	{"TestKindPredicateWalkRefusals/asserts", kindPredicateTypes + "function assertA(n: N): asserts n is A { if (n.kind !== K.A) { throw new Error('not A'); } }\nfunction show(n: N): string { assertA(n); return `a ${n.a}`; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"},
	{"TestKindPredicateWalkRefusals/switch_(true)", kindPredicateTypes + "function isA(n: N): n is A { return n.kind === K.A; }\nfunction show(n: N): string { switch (true) { case isA(n): return `a ${n.a}`; } return 'b'; }\nconsole.log(show({ kind: K.A, flag: F.X, a: 1 }));\n"},

	// TestKindAliasFallbacksKeepFamilyOne: let and other_reference are rows the real compiler lowers;
	// they're among the 134 as a mutant's child run failed them.
	{"TestKindAliasFallbacksKeepFamilyOne/let", kindPredicateTypes + kindAliasValues + "function show(n: N): string { let kind = n.kind; switch (kind) { case K.A: return n.kind === K.A ? `a ${n.a}` : 'x'; default: return 'b'; } }\nconsole.log(show(a));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/after", kindPredicateTypes + kindAliasValues + "function show(n: N): string { const read = (): string => { " + swingAliasCases + " }; const kind = n.kind; return read(); }\nconsole.log(show(a));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/other_function", kindPredicateTypes + kindAliasValues + "function show(n: N): string { const kind = n.kind; const read = (): string => { " + swingAliasCases + " }; return read(); }\nconsole.log(show(a));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/loop", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; let out = ''; for (let i = 0; i < 2; i++) { switch (kind) { case K.A: out += `a ${n.a};`; } n = m; } return out; }\nconsole.log(show(a, b));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/other_reference", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = m.kind; switch (kind) { case K.A: return n.kind === K.A ? `a ${n.a}` : 'x'; default: return 'b'; } }\nconsole.log(show(a, a));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/getter", swingAliasGetters + "function show(n: N): string { const kind = n.kind; " + swingAliasCases + " }\nconsole.log(show(a)); console.log(show(b));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/assigned", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; n = m; " + swingAliasCases + " }\nconsole.log(show(a, b));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/destructured", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; [n] = [m]; " + swingAliasCases + " }\nconsole.log(show(a, b));\n"},
	{"TestKindAliasFallbacksKeepFamilyOne/shorthand", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; ({ n } = { n: m }); " + swingAliasCases + " }\nconsole.log(show(a, b));\n"},

	// TestKindHelperShapesAgreeWithNode, each among the 134 as a mutant's child run failed it.
	{"TestKindHelperShapesAgreeWithNode/default_chain", swingKindHelperShapeSource(kindHelperChain, "return kinds(n.kind);")},
	{"TestKindHelperShapesAgreeWithNode/arrow", swingKindHelperShapeSource("const leaf=(k:K):boolean=>k===K.B;\n", "return n.kind===K.A || leaf(n.kind);")},
	{"TestKindHelperShapesAgreeWithNode/return_chain", swingKindHelperShapeSource(kindHelperLeaf+"function kinds(k:K):boolean { return k===K.A || leaf(k); }\n", "return kinds(n.kind);")},
	{"TestKindHelperShapesAgreeWithNode/switch_tail", swingKindHelperShapeSource(kindHelperLeaf+"function kinds(k:K):boolean { switch(k) {case K.A:return true;} return leaf(k); }\n", "return kinds(n.kind);")},

	// TestKindHelperFallbacksRejectWholeBody: isExpression and isStatement held family 1 in .ts; the
	// rest are among the 134 as a mutant's child run failed them. None prints.
	{"TestKindHelperFallbacksRejectWholeBody/h05", swingKindHelperFallbackSource("h05", "function kinds(k:K):boolean {switch(k) {case K.A:return true;case K.D:return false;default:return false;}}\n", "return n && kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/other_read", swingKindHelperFallbackSource("other_read", "const flag=true; function kinds(k:K):boolean {return k===K.A || flag;}\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/non_helper", swingKindHelperFallbackSource("non_helper", "function other(k:K):boolean {return k>0;} function kinds(k:K):boolean {return k===K.A || other(k);}\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/written_parameter", swingKindHelperFallbackSource("written_parameter", "function kinds(k:K):boolean {k=K.A;return k===K.A;}\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/default_parameter", swingKindHelperFallbackSource("default_parameter", "function kinds(k:K=K.A):boolean {return k===K.A;}\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/wrong_field", swingKindHelperFallbackSource("wrong_field", kindHelperLeaf, "return n.kind===K.A || leaf(n.other);")},
	{"TestKindHelperFallbacksRejectWholeBody/arithmetic_argument", swingKindHelperFallbackSource("arithmetic_argument", kindHelperLeaf, "return n.kind===K.A || leaf(n.kind+0);")},
	{"TestKindHelperFallbacksRejectWholeBody/two_enums", swingKindHelperFallbackSource("two_enums", "enum L { A=1,B=2 } function kinds(k:K|L):boolean {return k===L.B;}\n", "return n.kind===K.A || kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/depth", swingKindHelperFallbackSource("depth", kindHelperDepthChain(kindHelperMaxDepth+1), "return h0(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/cycle", swingKindHelperFallbackSource("cycle", "function a(k:K):boolean {return k===K.A || b(k);} function b(k:K):boolean {return a(k);}\n", "return a(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/overload", swingKindHelperFallbackSource("overload", "function kinds(k:K):boolean; function kinds(k:K):boolean {return k===K.A;}\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/let_helper", swingKindHelperFallbackSource("let_helper", "let kinds=(k:K):boolean=>k===K.A;\n", "return kinds(n.kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/alias_argument", swingKindHelperFallbackSource("alias_argument", kindHelperLeaf, "const kind=n.kind;return kind===K.A || leaf(kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/mixed_disjunction", swingKindHelperFallbackSource("mixed_disjunction", kindHelperLeaf+"function block(n:N):boolean {return n.kind===K.A;}\n", "return leaf(n.kind) || block(n);")},
	{"TestKindHelperFallbacksRejectWholeBody/isExpression", swingKindHelperFallbackSource("isExpression", kindHelperLeaf+"function skip(n:N):N {return n;}\n", "return leaf(skip(n).kind);")},
	{"TestKindHelperFallbacksRejectWholeBody/isStatement", swingKindHelperFallbackSource("isStatement", kindHelperLeaf+"function block(n:N):boolean {return n.kind===K.A;}\n", "const kind=n.kind;return leaf(kind) || block(n);")},

	// TestKindHelperEmptyDefaultTailRoutes, its two unrecognized rows.
	{"TestKindHelperEmptyDefaultTailRoutes/empty_false", swingKindHelperEmptyDefaultSource("switch(k) {default:case K.A:return false;} return leaf(k);")},
	{"TestKindHelperEmptyDefaultTailRoutes/empty_true", swingKindHelperEmptyDefaultSource("switch(k) {default:case K.A:return true;} return leaf(k);")},

	// TestKindSetBodyGrammar, the bodies of kindSetBodies, each among the 134 as a mutant's child run
	// failed it. None prints.
	{"TestKindSetBodyGrammar/disjunction", swingKindSetGrammarSource("return (n.kind===K.A) || (K.B===(n.kind));")},
	{"TestKindSetBodyGrammar/switch", swingKindSetGrammarSource("switch(n.kind) { case K.A: case K.B: return true; case K.C: return false; default: return false; }")},
	{"TestKindSetBodyGrammar/switch_tail", swingKindSetGrammarSource("switch(n.kind) { case K.A: return true; case K.B: return true; } return false;")},
	{"TestKindSetBodyGrammar/alias", swingKindSetGrammarSource("const k=n.kind; return k===K.A || k===K.B;")},
	{"TestKindSetBodyGrammar/guard", swingKindSetGrammarSource("return n && (n.kind===K.A || n.kind===K.B);")},
	{"TestKindSetBodyGrammar/double_guard", swingKindSetGrammarSource("return !!n && (n.kind===K.A || n.kind===K.B);")},
	{"TestKindSetBodyGrammar/if_guard_alias", swingKindSetGrammarSource("if(n) { const kind=n.kind; return kind===K.A || kind===K.B; } return false;")},

	// TestKindSetPredicateFallbacksKeepFamilyOne: k04 and primitive_guard are among the 134 as a mutant's
	// child run failed them. None prints.
	{"TestKindSetPredicateFallbacksKeepFamilyOne/extra_disjunct", swingKindSetFallbackSource("extra_disjunct", "return n.kind===K.A || n.kind===K.B || n.kind>0;", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/call", swingKindSetFallbackSource("call", "return n.kind===K.A || more(n);", "function more(n:N):boolean { return n.kind===K.B; }\n")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/written_alias", swingKindSetFallbackSource("written_alias", "let k=n.kind; return k===K.A || k===K.B;", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/alias_other_read", swingKindSetFallbackSource("alias_other_read", "const k=n.kind; return k===K.A || n.kind===K.B;", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/nonliteral_case", swingKindSetFallbackSource("nonliteral_case", "switch(n.kind) { case K.A: case K.B: return 1>0; default: return false; }", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/true_default", swingKindSetFallbackSource("true_default", "switch(n.kind) { case K.A: case K.B: return true; default: return true; }", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/two_fields", swingKindSetFallbackSource("two_fields", "return n.kind===K.A || n.other===K.B;", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/two_enums", swingKindSetFallbackSource("two_enums", "return n.kind===K.A || n.kind===L.B;", "enum L { B=2, D=4 }\n")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case", swingKindSetFallbackSource("duplicate_case", "switch(n.kind) { case K.A: return false; case K.A: return true; default: return false; }", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/k04", swingKindSetFallbackSource("k04", "switch(n.kind) { case K.A: return false; case K.D: return true; default: return false; }", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case_true_first", swingKindSetFallbackSource("duplicate_case_true_first", "switch(n.kind) { case K.A: return true; case K.A: return false; default: return false; }", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/primitive_guard", swingKindSetFallbackSource("primitive_guard", "return !!n && (n.length===K.A || n.length===K.B);", "")},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/not_equal", swingKindSetFallbackSource("not_equal", "return n.kind!==K.C;", "")},

	// TestKindSetClosedSourcesFitEveryKind, its row that doesn't fit.
	{"TestKindSetClosedSourcesFitEveryKind/one_does_not_fit", closedKindTypes + "interface ExtraNum extends Ex { readonly kind:K.Num; readonly extra:number }\nconst numbered:ExtraNum={kind:K.Num,pos:2,extra:9,_expressionBrand:undefined};\n" + swingFitPredicate + "console.log(show(body)); console.log(show(ident)); console.log(show(numbered));\n"},

	// TestKindSetUnionClaimFitsEverySource: base + the row's source.
	{"TestKindSetUnionClaimFitsEverySource/odd_declared", swingUnionClaimBase + "interface Odd extends Ex { readonly kind:K.A; readonly b:string }\n"},
	{"TestKindSetUnionClaimFitsEverySource/odd_built", swingUnionClaimBase + "interface Odd extends Ex { readonly kind:K.A; readonly b:string }\nconst odd:Odd={kind:K.A,pos:2,_exBrand:undefined,b:'y'}; console.log(show(odd));\n"},
	{"TestKindSetUnionClaimFitsEverySource/odd_anonymous", swingUnionClaimBase + "const raw={kind:K.A as const,pos:3,_exBrand:undefined,b:'w'}; const anon:Ex=raw; console.log(show(anon));\n"},

	// TestClosedKindFunctionConstructorNarrowing, its called row.
	{"TestClosedKindFunctionConstructorNarrowing/called", closedKindConstructorTypes + "OpenNode.call(open);\n" + swingConstructorNarrowing},

	// TestClosedKindHoldingTheComparedKindRefuses: closedKindTypes + show + the row's source.
	{"TestClosedKindHoldingTheComparedKindRefuses/anonymous_literal", closedKindTypes + swingHoldingShow + "const e: Ex = { kind: K.Body, pos: 2, _expressionBrand: undefined };\nconsole.log(show(e));\n"},
	{"TestClosedKindHoldingTheComparedKindRefuses/twin_layout", closedKindTypes + swingHoldingShow + "interface Twin extends Ex { readonly kind: K.Body; readonly t: number }\n"},

	// TestClosedKindLiteralKindBuildsStayOpen, its .ts rows: types + the row's source.
	{"TestClosedKindLiteralKindBuildsStayOpen/annotated_literal.ts", swingLiteralBuildTypes + "const odd: " + swingLiteralBuildAnnotated + " = { kind: Kind.Block, text: 'xyz' };\nconst lie: Expr = odd;\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/as_const_kind.ts", swingLiteralBuildTypes + "const lie: Expr = { kind: Kind.Block as const, text: 'xyz' };\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/inline_intersection.ts", swingLiteralBuildTypes + "const odd: Node & " + swingLiteralBuildAnnotated + " = { kind: Kind.Block, text: 'xyz' };\nconst lie: Expr = odd;\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/return_type.ts", swingLiteralBuildTypes + "function make(): " + swingLiteralBuildAnnotated + " { return { kind: Kind.Block, text: 'xyz' }; }\nconst lie: Expr = make();\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/parameter_type.ts", swingLiteralBuildTypes + "function take(odd: " + swingLiteralBuildAnnotated + "): string { const lie: Expr = odd; return measure(lie); }\nconsole.log(measure(ident) + ' ' + take({ kind: Kind.Block, text: 'xyz' }));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/widened_kind.ts", swingLiteralBuildTypes + "const odd = { kind: Kind.Block, text: 'xyz' };\nconst lie: Expr = odd;\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},
	{"TestClosedKindLiteralKindBuildsStayOpen/widened_as.ts", swingLiteralBuildTypes + "const lie = { kind: Kind.Block, text: 'xyz' } as Expr;\nconsole.log(measure(ident) + ' ' + measure(lie));\n"},

	// TestClosedKindRunTimeKindStaysOpen: closedKindTypes + open + the row's source.
	{"TestClosedKindRunTimeKindStaysOpen/class", closedKindTypes + swingRunTimeOpen + "class N { constructor(readonly kind: K, readonly pos: number, readonly n: number) {} }\nconsole.log(show(new N(kb(), 0, 1)));\n"},
	{"TestClosedKindRunTimeKindStaysOpen/function", closedKindTypes + swingRunTimeOpen + "function make(kind: K): Open { return { kind, pos: 0, n: 1 }; }\nconsole.log(show(make(kb())));\n"},

	// TestClosedKindSourcesNarrowByThePredicate: closedKindTypes + closedKindSourcesNarrowing + the row's source.
	{"TestClosedKindSourcesNarrowByThePredicate/anonymous_literal,_no_name", closedKindTypes + closedKindSourcesNarrowing + "const raw = { kind: K.Ident as const, pos: 9, _expressionBrand: undefined };\nconst anon: Ex = raw;\nconsole.log(show(anon));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/class,_no_name", closedKindTypes + closedKindSourcesNarrowing + "class OwnIdent { readonly kind = K.Ident as const; readonly pos = 0; readonly _expressionBrand = undefined; }\nconst made: Ex = new OwnIdent();\nconsole.log(show(made));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/literal_cast_as_Ex,_no_name", closedKindTypes + closedKindSourcesNarrowing + "const cast = { kind: K.Ident as const, pos: 1, _expressionBrand: undefined } as Ex;\nconsole.log(show(cast));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/declared_union_with_an_anonymous_member,_no_name", closedKindTypes + closedKindSourcesNarrowing + "type Odd = Ident | { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined };\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/inline_intersection,_no_name", closedKindTypes + closedKindSourcesNarrowing + "const inter: Node & { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined } = { kind: K.Ident, pos: 3, _expressionBrand: undefined };\nconst viaInter: Ex = inter;\nconsole.log(show(viaInter));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/return_type,_no_name", closedKindTypes + closedKindSourcesNarrowing + "function make(): { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined } { return { kind: K.Ident, pos: 4, _expressionBrand: undefined }; }\nconst made: Ex = make();\nconsole.log(show(made));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/parameter_type,_no_name", closedKindTypes + closedKindSourcesNarrowing + "function take(odd: { readonly kind: K.Ident; readonly pos: number; readonly _expressionBrand: undefined }): string { const viaParam: Ex = odd; return show(viaParam); }\nconsole.log(take({ kind: K.Ident, pos: 5, _expressionBrand: undefined }));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/widened_kind", closedKindTypes + closedKindSourcesNarrowing + "const wide = { kind: K.Ident, pos: 6, _expressionBrand: undefined, name: 'w' };\nconst viaWide: Ex = wide;\nconsole.log(show(viaWide));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/widened_kind,_no_name", closedKindTypes + closedKindSourcesNarrowing + "const wide = { kind: K.Ident, pos: 6, _expressionBrand: undefined };\nconst viaWide: Ex = wide;\nconsole.log(show(viaWide));\n"},
	{"TestClosedKindSourcesNarrowByThePredicate/declared_source_with_another_field_type", closedKindTypes + closedKindSourcesNarrowing + "interface NumIdent extends Ex { readonly kind: K.Ident; readonly name: number }\nconst numbered: NumIdent = { kind: K.Ident, pos: 8, _expressionBrand: undefined, name: 3 };\nconsole.log(show(numbered));\n"},

	// TestCheckedPredicatePendingContracts, its three open-tag rows, which it lowered as main.ts.
	{"TestCheckedPredicatePendingContracts/open_enum_union_tag", `enum Kind { A = 1, B = 2 } interface A { readonly kind: Kind.A; readonly value: number } interface B { readonly kind: Kind; readonly value: string } type Node = A | B; function isA(n: Node): n is A { return n.kind === Kind.A; } function numeric(): number { return 1; } const b: B = { kind: numeric(), value: 'text' }; function use(n: Node): void { if (isA(n)) { console.log(` + "`${n.value + 1}`" + `); } } use(b);`},
	{"TestCheckedPredicatePendingContracts/open_tag_false_read", `enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function f(n: N): void { if (!isA(n)) console.log(n.b); } function numeric(): number { return 1; } const wide: A = { kind: numeric(), a: 0 }; f({ kind: Pair.B, b: 'text' });`},
	{"TestCheckedPredicatePendingContracts/open_tag_false_pass", `enum Lone { A = 1 } enum Pair { B = 2, C = 3 } interface A { readonly kind: Lone.A; readonly a: number } interface B { readonly kind: Pair.B; readonly b: string } type N = A | B; function isA(n: N): n is A { return n.kind === Lone.A; } function g(b: B): void { console.log(b.b); } function f(n: N): void { if (!isA(n)) g(n); } function numeric(): number { return 1; } const wide: A = { kind: numeric(), a: 0 }; f({ kind: Pair.B, b: 'text' });`},
}

// swingRowsNotReproduced is every row of the 134 that one .ts source can't reproduce, each with why.
var swingRowsNotReproduced = []struct {
	name   string
	reason string
}{
	{"TestKindPredicateThroughModules/namespace_import", "three files, library.ts, relay.ts and main.ts, loaded together"},
	{"TestClosedKindSourcesReadTypeOnlyImports/no_name", "two files, main.ts and the lib.d.ts it imports WideIdent from by type"},
	{"TestEnumerationReachDeepGetterRefused/field", "two .a files, held.a and main.a, and no enum tag"},
	{"TestEnumerationReachDeepGetterRefused/element", "two .a files, held.a and main.a, and no enum tag"},
	{"TestEnumerationReachReachingGetterRefused/discriminated", "two .a files, held.a and main.a, and no enum tag"},
	{"TestEnumerationReachReachingGetterRefused/generic", "two .a files, held.a and main.a, and no enum tag"},
	{"TestEnumerationReachReachingGetterRefused/held", "two .a files, held.a and main.a, and no enum tag"},
	{"TestEnumerationReachViewRefused/sibling", "two files, held.ts and main.ts, and no enum tag"},
	// The verbose run listed these 16 twice: once from the row itself, and again, indented deeper, from a
	// mutant test's child run echoing its failures. The table holds each row once, under its own name.
	{"TestKindAliasFallbacksKeepFamilyOne/after (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/after, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/assigned (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/assigned, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/destructured (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/destructured, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/getter (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/getter, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/loop (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/loop, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/other_function (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/other_function, which the table holds"},
	{"TestKindAliasFallbacksKeepFamilyOne/shorthand (mutant child run)", "the same row as TestKindAliasFallbacksKeepFamilyOne/shorthand, which the table holds"},
	{"TestKindHelperEmptyDefaultTailRoutes/empty_false (mutant child run)", "the same row as TestKindHelperEmptyDefaultTailRoutes/empty_false, which the table holds"},
	{"TestKindSetClosedSourcesFitEveryKind/one_does_not_fit (mutant child run)", "the same row as TestKindSetClosedSourcesFitEveryKind/one_does_not_fit, which the table holds"},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case (mutant child run)", "the same row as TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case, which the table holds"},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/not_equal (mutant child run)", "the same row as TestKindSetPredicateFallbacksKeepFamilyOne/not_equal, which the table holds"},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/true_default (mutant child run)", "the same row as TestKindSetPredicateFallbacksKeepFamilyOne/true_default, which the table holds"},
	{"TestKindSetPredicateFallbacksKeepFamilyOne/two_enums (mutant child run)", "the same row as TestKindSetPredicateFallbacksKeepFamilyOne/two_enums, which the table holds"},
	{"TestKindSetUnionClaimFitsEverySource/odd_anonymous (mutant child run)", "the same row as TestKindSetUnionClaimFitsEverySource/odd_anonymous, which the table holds"},
	{"TestKindSetUnionClaimFitsEverySource/odd_built (mutant child run)", "the same row as TestKindSetUnionClaimFitsEverySource/odd_built, which the table holds"},
	{"TestKindSetUnionClaimFitsEverySource/odd_declared (mutant child run)", "the same row as TestKindSetUnionClaimFitsEverySource/odd_declared, which the table holds"},
}

// Pins that every one of the 134 rows the swing moved (#x2cjvy9, Adamic's ruling of Nov 25) is here,
// reproduced in enumTagSwingRows or named in swingRowsNotReproduced with why, each once.
func TestEnumTagSwingRowsAreCounted(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, row := range enumTagSwingRows {
		if seen[row.name] {
			t.Errorf("SWING row %q is listed twice", row.name)
		}
		seen[row.name] = true
	}
	for _, row := range swingRowsNotReproduced {
		if seen[row.name] {
			t.Errorf("SWING row %q is listed twice", row.name)
		}
		seen[row.name] = true
		if row.reason == "" {
			t.Errorf("SWING row %q is not reproduced and says not why", row.name)
		}
	}
	if count := len(enumTagSwingRows) + len(swingRowsNotReproduced); count != 134 {
		t.Fatalf("SWING rows: %d reproduced and %d not, %d in all; want the 134 the swing moved", len(enumTagSwingRows), len(swingRowsNotReproduced), count)
	}
}

// enumTagSwingClass names one backend's run against Node's: "STOP" when it exits 70 at a checked read
// ("field read failed") having printed only a prefix of what Node printed, "AGREE" when it prints what
// Node prints and exits as Node does (any nonzero exit where Node throws), and "WRONG" otherwise. A
// sanitizer's report is "WRONG" whatever else the run did.
func enumTagSwingClass(code int, stdout, stderr string, node nodeObservation) string {
	if strings.Contains(stderr, "Sanitizer:") {
		return "WRONG"
	}
	// A loud stop is the runtime's own deliberate panic at exit 70 with nothing printed past what Node
	// printed: the checked view's "field read failed" at the read, or an earlier check that catches the
	// same lie, such as a checked predicate's claim at its call ("predicate isAB ... false branch").
	if code == 70 && strings.HasPrefix(stderr, "adamic: panic: ") && strings.HasPrefix(string(node.stdout), stdout) {
		return "STOP"
	}
	if stdout == string(node.stdout) && (code == node.code || code != 0 && node.code != 0) {
		return "AGREE"
	}
	return "WRONG"
}

// Pins the .ts side of rulings.md:258 after the swing (#x2cjvy9, Adamic's ruling of Nov 25): an
// open-enum refinement the walk can't prove is read through a checked view, so a lie stops at its read
// and never runs silently wrong. Each row either stays refused or not yet, or lowers, and then sanitized
// native (ASan and UBSan) and the JavaScript backend each agree with Node or stop at the checked read.
func TestEnumTagSwingTsRowsStopOrAgree(t *testing.T) {
	t.Parallel()
	for _, row := range enumTagSwingRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				var refused *Refused
				var notYet *NotYet
				if !errors.As(err, &refused) && !errors.As(err, &notYet) {
					t.Fatalf("SWING %s compile: want a *Refused or a *NotYet where it doesn't lower, got %v", row.name, err)
				}
				t.Logf("SWING %s compile %v", row.name, err)
				return
			}
			node := runAgreementNode(t, writeAgreementSource(t, "main.ts", row.source))
			nativeCode, nativeStdout, nativeStderr := brandedNative(t, program)
			generated := runAgreementNode(t, writeAgreementSource(t, "generated.mjs", javascript.JavaScript(program)))
			nativeClass := enumTagSwingClass(nativeCode, nativeStdout, nativeStderr, node)
			javascriptClass := enumTagSwingClass(generated.code, string(generated.stdout), string(generated.stderr), node)
			t.Logf("SWING %s %s %s", row.name, nativeClass, javascriptClass)
			if nativeClass == "WRONG" {
				t.Errorf("SWING %s native ran wrong: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", row.name, node.code, node.stdout, node.stderr, nativeCode, nativeStdout, nativeStderr)
			}
			if javascriptClass == "WRONG" {
				t.Errorf("SWING %s JavaScript ran wrong: Node exit %d stdout %q stderr %q; JavaScript exit %d stdout %q stderr %q", row.name, node.code, node.stdout, node.stderr, generated.code, generated.stdout, generated.stderr)
			}
		})
	}
}
