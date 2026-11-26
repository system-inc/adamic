package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/javascript"
)

const kindSetTypes = `enum K { A=1, B=2, C=3 }
enum F { X=1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: string }
interface B { readonly kind: K.B; readonly b: number }
interface C { readonly kind: K.C; readonly c: boolean }
type N=A|B|C;
`
const kindSetValues = `const a:A={kind:K.A,flag:F.X,a:'alpha'};
const b:B={kind:K.B,b:7};
const c:C={kind:K.C,c:false};
console.log(show(a)); console.log(show(b)); console.log(show(c));
`
const kindSetShow = `function show(n:N):string {
 if (isAB(n)) { if(n.kind===K.A) return n.a; return String(n.b); }
 return String(n.c);
}
`

var kindSetBodies = []struct{ name, body, refused string }{
	{"disjunction", "return (n.kind===K.A) || (K.B===(n.kind));", ""},
	{"switch", "switch(n.kind) { case K.A: case K.B: return true; case K.C: return false; default: return false; }", "return paths through KindSwitchStatement are not verified"},
	{"switch_tail", "switch(n.kind) { case K.A: return true; case K.B: return true; } return false;", "return paths through KindSwitchStatement are not verified"},
	{"alias", "const k=n.kind; return k===K.A || k===K.B;", "return expression is not a trusted check on n"},
	{"guard", "return n && (n.kind===K.A || n.kind===K.B);", "return expression is not a trusted check on n"},
	{"double_guard", "return !!n && (n.kind===K.A || n.kind===K.B);", "return expression is not a trusted check on n"},
	{"if_guard_alias", "if(n) { const kind=n.kind; return kind===K.A || kind===K.B; } return false;", "branch is not a trusted parameter check"},
}

// U1 judges calls in .ts; rulings.md:141 leaves .a's independent predicate
// verifier intact. Each admitted shape reads both branches and all three kinds.
func TestKindSetPredicateShapesAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, p := range kindSetBodies {
		for _, extension := range []string{".ts", ".a"} {
			if extension == ".a" && p.refused != "" {
				continue
			}
			t.Run(p.name+extension, func(t *testing.T) {
				t.Parallel()
				source := kindSetTypes + "function isAB(n:N):n is A|B { " + p.body + " }\n" + kindSetShow + kindSetValues
				program := lowersAndAgreesWithNodeAs(t, source, extension)
				if code, out, errs := brandedNative(t, program); code != 0 || out != "alpha\n7\nfalse\n" {
					t.Fatalf("kind-set %s: exit=%d stdout=%q stderr=%q", p.name, code, out, errs)
				}
			})
		}
	}
}
func TestKindSetDotAKeepsPredicateVerifierRefusals(t *testing.T) {
	t.Parallel()
	for _, p := range kindSetBodies {
		if p.refused == "" {
			continue
		}
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			source := kindSetTypes + "function isAB(n:N):n is A|B { " + p.body + " }\n" + kindSetShow + kindSetValues
			_, err := lowerSourceAs(t, source, ".a")
			var refused *Refused
			want := "a type predicate whose return is not proven (" + p.refused + ")"
			if !errors.As(err, &refused) || refused.What != want {
				t.Fatalf(".a verifier: want %q, got %v", want, err)
			}
		})
	}
}

// These direct controls isolate the grammar from .a's verifier and any later
// lowering stop. Every member of the complete body must be read into its set.
func TestKindSetBodyGrammar(t *testing.T) {
	t.Parallel()
	for _, p := range kindSetBodies {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			l, statements := enumGuardProof(t, kindSetTypes+"function isAB(n:N):n is A|B { "+p.body+" }\n")
			for _, n := range statements {
				if n.Kind == ast.KindFunctionDeclaration && n.Name().Text() == "isAB" {
					field, members, ok := l.kindPredicateBody(n.Body(), n.Parameters()[0].Name())
					if !ok || field != "kind" || len(members) != 2 {
						t.Fatalf("kind-set grammar %s: field=%q members=%d recognized=%v", p.name, field, len(members), ok)
					}
					return
				}
			}
			t.Fatal("missing predicate")
		})
	}
}
func TestKindSetNullableGuardsKeepMembershipBoundary(t *testing.T) {
	t.Parallel()
	for _, p := range kindSetBodies {
		if p.name != "double_guard" && p.name != "if_guard_alias" {
			continue
		}
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			source := kindSetTypes + "function isAB(n:N|undefined):n is A|B { " + p.body + " }\n" +
				"function show(n:N|undefined):string { if(isAB(n)) { if(n.kind===K.A) return n.a; return String(n.b); } if(!n) return 'missing'; return String(n.c); }\n" + kindSetValues + "console.log(show(undefined));\n"
			_, err := lowerSourceAs(t, source, ".ts")
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != "checked predicate membership over an open tag domain" {
				t.Fatalf("nullable membership boundary: %v", err)
			}
		})
	}
}
func TestKindSetConstArrowKeepsBothBoundaries(t *testing.T) {
	t.Parallel()
	const predicate = "const isAB=(n:N):n is A|B => n.kind===K.A || n.kind===K.B;\n"
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			trueSource := kindSetTypes + predicate + strings.Replace(kindSetShow, "return String(n.c);", "return 'other';", 1) + kindSetValues
			program := lowersAndAgreesWithNodeAs(t, trueSource, extension)
			if code, out, errs := brandedNative(t, program); code != 0 || out != "alpha\n7\nother\n" {
				t.Fatalf("arrow true side: exit=%d stdout=%q stderr=%q", code, out, errs)
			}
			falseSource := kindSetTypes + predicate + kindSetShow + kindSetValues
			if extension == ".a" {
				// Its verified body proves the false answer; .ts keeps the checked closure.
				program := lowersAndAgreesWithNodeAs(t, falseSource, extension)
				if code, out, errs := brandedNative(t, program); code != 0 || out != "alpha\n7\nfalse\n" {
					t.Fatalf("verified arrow false side: exit=%d stdout=%q stderr=%q", code, out, errs)
				}
				return
			}
			_, err := lowerSourceAs(t, falseSource, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || !strings.Contains(stop.What, "a call through a predicate value whose false answer isn't checked") {
				t.Fatalf("arrow false-side boundary: %v", err)
			}
		})
	}
}

func TestKindSetPredicateFallbacksKeepFamilyOne(t *testing.T) {
	t.Parallel()
	for _, p := range []struct{ name, body, extra string }{
		{"extra_disjunct", "return n.kind===K.A || n.kind===K.B || n.kind>0;", ""},
		{"call", "return n.kind===K.A || more(n);", "function more(n:N):boolean { return n.kind===K.B; }\n"},
		{"written_alias", "let k=n.kind; return k===K.A || k===K.B;", ""},
		{"alias_other_read", "const k=n.kind; return k===K.A || n.kind===K.B;", ""},
		{"nonliteral_case", "switch(n.kind) { case K.A: case K.B: return 1>0; default: return false; }", ""},
		{"true_default", "switch(n.kind) { case K.A: case K.B: return true; default: return true; }", ""},
		{"two_fields", "return n.kind===K.A || n.other===K.B;", ""},
		{"two_enums", "return n.kind===K.A || n.kind===L.B;", "enum L { B=2, D=4 }\n"},
		{"duplicate_case", "switch(n.kind) { case K.A: return false; case K.A: return true; default: return false; }", ""},
		{"k04", "switch(n.kind) { case K.A: return false; case K.D: return true; default: return false; }", ""},
		{"duplicate_case_true_first", "switch(n.kind) { case K.A: return true; case K.A: return false; default: return false; }", ""},
		{"primitive_guard", "return !!n && (n.length===K.A || n.length===K.B);", ""},
		{"not_equal", "return n.kind!==K.C;", ""},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			types := kindSetTypes
			if p.name == "k04" {
				types = strings.Replace(types, "C=3", "C=3, D=1", 1)
			}
			if p.name == "two_fields" {
				types = strings.ReplaceAll(types, "readonly kind:", "readonly other:K; readonly kind:")
			}
			if p.name == "two_enums" {
				types = strings.Replace(types, "readonly kind: K.B", "readonly kind: L.B", 1)
			}
			source := p.extra + types + "function isAB(n:N):n is A|B { " + p.body + " }\n" + kindSetShow
			if p.name == "primitive_guard" {
				source = "enum K { A=1, B=2 } function isAB(n:string):n is string { " + p.body + " }"
			}
			l, statements := enumGuardProof(t, source)
			for _, declaration := range statements {
				if declaration.Kind == ast.KindFunctionDeclaration && declaration.Name().Text() == "isAB" {
					if _, _, recognized := l.kindPredicateBody(declaration.Body(), declaration.Parameters()[0].Name()); recognized {
						t.Fatalf("kind-set body fallback %s was partially recognized", p.name)
					}
				}
			}
			if p.name == "primitive_guard" || p.name == "k04" {
				return
			} // These pins isolate the recognizer from unrelated later fences.
			switch p.name {
			// extra_disjunct: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/extra_disjunct".
			// call: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/call".
			// written_alias: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/written_alias".
			// alias_other_read: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/alias_other_read".
			// nonliteral_case: .a stops earlier at a type predicate whose return is not proven (its paths through a kind switch are not verified); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/nonliteral_case".
			// true_default: .a stops earlier at a type predicate whose return is not proven (its paths through a kind switch are not verified); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/true_default".
			// two_fields: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/two_fields".
			// duplicate_case: .a stops earlier at a type predicate whose return is not proven (its paths through a kind switch are not verified); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case".
			// duplicate_case_true_first: .a stops earlier at a type predicate whose return is not proven (its paths through a kind switch are not verified); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/duplicate_case_true_first".
			// not_equal: .a stops earlier at a type predicate whose return is not proven (its return is not a trusted check); its .ts side is enumTagSwingRows' "TestKindSetPredicateFallbacksKeepFamilyOne/not_equal".
			case "extra_disjunct", "call", "written_alias", "alias_other_read", "nonliteral_case", "true_default", "two_fields", "duplicate_case", "duplicate_case_true_first", "not_equal":
				return
			}
			if _, err := lowerSourceAs(t, source, ".a"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
				t.Fatalf("kind-set fallback %s: want family 1, got %v", p.name, err)
			}
		})
	}
}

// R-T80 still refuses membership over this open tag domain before lowering the
// read. Ask the flow proof directly too, so the false-side set rule is exercised
// and cannot silently change when that earlier fence lifts.
func TestKindSetFalseSideKeepsPartialTag(t *testing.T) {
	t.Parallel()
	source := strings.Replace(kindSetTypes, "readonly kind: K.C", "readonly kind: K.B|K.C", 1) +
		"function isAB(n:N):n is A|B { return n.kind===K.A || n.kind===K.B; }\n" +
		"function show(n:N):string { if(isAB(n)) return 'selected'; return String(n.c); }\n" +
		"const c:C={kind:K.C,c:false}; console.log(show(c));\n"
	l, statements := enumGuardProof(t, source)
	var read *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindPropertyAccessExpression && n.Name().Text() == "c" {
			read = n.Expression()
			return true
		}
		return n.ForEachChild(visit)
	}
	for _, n := range statements {
		if visit(n) {
			break
		}
	}
	if read == nil {
		t.Fatal("missing false-side read")
	}
	declared := l.checker.GetTypeOfSymbol(l.checker.GetSymbolAtLocation(read))
	if !l.enumTagRefinementProven(read, declared) {
		t.Fatal("kind-set partial tag was dropped from the false side")
	}
	if _, err := lowerSourceAs(t, source, ".ts"); err != nil && !strings.Contains(err.Error(), "membership over an open tag domain") {
		t.Fatalf("partial-tag fail-closed stop: %v", err)
	}
}

func TestKindSetPredicateLieStopsAtCheckedRead(t *testing.T) {
	t.Parallel()
	const source = `enum Kind { A, B, C }
interface Base { readonly kind:Kind }
interface XA extends Base { readonly kind:Kind.A; readonly a:string }
interface ZA extends Base { readonly kind:Kind.B; readonly a:string }
interface XB extends Base { readonly kind:Kind.C; readonly b:number }
interface YA extends Base { readonly kind:Kind.A; readonly a:number }
function isText(n:Base):n is XA|ZA { return n.kind===Kind.A || n.kind===Kind.B; }
function guard(run:()=>string):string { try { return run(); } catch(error) { return 'threw'; } }
type Inner=YA|XB;
type Outer=Inner|XA|ZA;
function show(s:Outer):string { return isText(s) ? guard(()=>String(s.a).toUpperCase()+typeof s.a) : 'other'; }
const ya:YA={kind:Kind.A,a:3}; console.log(show(ya));
`
	if observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); observed.code != 0 || string(observed.stdout) != "3number\n" {
		t.Fatalf("lying Node twin: %+v", observed)
	}
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("kind-set checked lie: %v", err)
	}
	brandedStopsProgram(t, program, "adamic: panic: field read failed: s.a is not a string")
}

// Optimizer's source-fit seam: Ident's kind fits the resolved claim, while a
// second source of Num's kind may lack NumLit's fields. A successful first kind
// never licenses replacing the closed Ex by the whole Ident | NumLit claim.
func TestKindSetClosedSourcesFitEveryKind(t *testing.T) {
	t.Parallel()
	const predicate = "function isSelected(n:Node):n is Ident|NumLit { return n.kind===K.Ident || n.kind===K.Num; }\n" +
		"function show(n:Body|Ex):string { if(isSelected(n)) { if(n.kind===K.Ident) return n.name; return n.text; } return 'other'; }\n"
	for _, p := range []struct {
		name, source string
		fit          bool
	}{
		{"all_fit", "interface ExtraNum extends NumLit { readonly extra:number }\nconst numbered:ExtraNum={kind:K.Num,pos:2,extra:9,text:'7',_expressionBrand:undefined,_literalBrand:undefined};\n", true},
		{"one_does_not_fit", "interface ExtraNum extends Ex { readonly kind:K.Num; readonly extra:number }\nconst numbered:ExtraNum={kind:K.Num,pos:2,extra:9,_expressionBrand:undefined};\n", false},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + p.source + predicate + "console.log(show(body)); console.log(show(ident)); console.log(show(numbered));\n"
			if !p.fit {
				// .a stops earlier at the predicate under rulings.md:142, so the fit is judged in .ts: a source that
				// doesn't fit leaves the refinement to the swing's checked view (enumTagSwingSites), and its
				// agreement is enumTagSwingRows' "TestKindSetClosedSourcesFitEveryKind/one_does_not_fit".
				if _, sites, err := lowerSourceSwung(t, source); len(sites) == 0 {
					t.Fatalf("kind-set all-kinds fit: a closed source with only its first kind fitting narrowed with no swing (%v)", err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, out, errs := brandedNative(t, program); code != 0 || out != "other\nx\n7\n" {
				t.Fatalf("closed kind-set sources: exit=%d stdout=%q stderr=%q", code, out, errs)
			}
		})
	}
}

// Body is in the predicate set but outside K(Ex): it has no Ex source to fit.
// The Ident source still licenses replacing Ex by the resolved whole claim.
func TestKindSetClosedSourcesIgnoreKindsOutsideDomain(t *testing.T) {
	t.Parallel()
	const predicate = "function selected(n:Node):n is Ident|Body { return n.kind===K.Ident || n.kind===K.Body; }\n" +
		"function show(n:Body|Ex):string { if(selected(n)) { if(n.kind===K.Body) return n.b; return n.name; } return 'other'; }\n"
	source := closedKindTypes + predicate + "console.log(show(body)); console.log(show(ident));\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, out, errs := brandedNative(t, program); code != 0 || out != "block\nx\n" {
		t.Fatalf("outside closed kind: exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

// Equal-valued tags make isD true for A at runtime. The existing checked-view
// boundary stops the missing string read; neither singleton nor set reads garbage.
func TestKindSetEqualValueTagsStopAtCheckedRead(t *testing.T) {
	t.Parallel()
	const common = `enum K { A=1, B=2, D=1 }
interface A { readonly kind: K.A; readonly a: number }
interface D { readonly kind: K.D; readonly d: string }
`
	for _, p := range []struct{ name, source string }{
		{"singleton", "function isD(n:A|D):n is D { return n.kind===K.D; }\nfunction show(n:A|D):void { console.log(String(isD(n))); if(isD(n)) console.log(n.d); }\n"},
		{"set", "interface B { readonly kind:K.B; readonly d:string }\nfunction isD(n:A|D|B):n is D|B { return n.kind===K.D || n.kind===K.B; }\nfunction show(n:A|D|B):void { console.log(String(isD(n))); if(isD(n)) console.log(n.d); }\n"},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			source := common + p.source + "const a:A={kind:K.A,a:7}; show(a);\n"
			observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
			if observed.code != 0 || string(observed.stdout) != "true\nundefined\n" {
				t.Fatalf("equal-value Node: %+v", observed)
			}
			program, err := lowerSourceAs(t, source, ".ts")
			if err != nil {
				t.Fatalf("equal-value checked read no longer lowers: %v", err)
			}
			wantError := "adamic: panic: field read failed: n.d is not initialized; expected string, found missing\n"
			generated := writeAgreementSource(t, "generated.mjs", javascript.JavaScript(program))
			js := runAgreementNode(t, generated)
			if js.code != 70 || string(js.stdout) != "true\n" || string(js.stderr) != wantError {
				t.Fatalf("equal-value JavaScript: exit=%d stdout=%q stderr=%q", js.code, js.stdout, js.stderr)
			}
			code, out, errs := brandedNative(t, program)
			if code != 70 || out != "true\n" || errs != wantError {
				t.Fatalf("equal-value native: exit=%d stdout=%q stderr=%q", code, out, errs)
			}
		})
	}
}

// A union claim must fit every source: being a union never bypasses source fit.
// Odd carries A's kind but B's layout and relates to neither member.
func TestKindSetUnionClaimFitsEverySource(t *testing.T) {
	t.Parallel()
	const base = `enum K { Body=1, A=2, B=3 }
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
	for _, p := range []struct{ name, source string }{
		{"control", ""},
		{"odd_declared", "interface Odd extends Ex { readonly kind:K.A; readonly b:string }\n"},
		{"odd_built", "interface Odd extends Ex { readonly kind:K.A; readonly b:string }\nconst odd:Odd={kind:K.A,pos:2,_exBrand:undefined,b:'y'}; console.log(show(odd));\n"},
		{"odd_anonymous", "const raw={kind:K.A as const,pos:3,_exBrand:undefined,b:'w'}; const anon:Ex=raw; console.log(show(anon));\n"},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			source := base + p.source
			if p.name == "control" {
				program := lowersAndAgreesWithNodeAs(t, source, ".ts")
				if code, out, errs := brandedNative(t, program); code != 0 || out != "other\nx\n" {
					t.Fatalf("union-claim control: exit=%d stdout=%q stderr=%q", code, out, errs)
				}
				return
			}
			// .a stops earlier at the predicate under rulings.md:142, so the fit is judged in .ts: a source fitting
			// neither member leaves the refinement to the swing's checked view (enumTagSwingSites), and each row's
			// agreement is enumTagSwingRows' "TestKindSetUnionClaimFitsEverySource/<row>".
			if _, sites, err := lowerSourceSwung(t, source); len(sites) == 0 {
				t.Fatalf("kind-set union claim %s: a source fitting neither claimed member narrowed with no swing (%v)", p.name, err)
			}
		})
	}
}
