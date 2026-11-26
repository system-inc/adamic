package lower

import (
	"errors"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
	"testing"
)

const kindHelperLeaf = "function leaf(k:K):boolean { switch(k) { case K.B:return true; default:return false; } }\n"
const kindHelperChain = kindHelperLeaf + "function kinds(k:K):boolean { switch(k) { case K.A:return true; default:return leaf(k); } }\n"

func kindHelperRecognized(t *testing.T, source string) (bool, int) {
	t.Helper()
	l, statements := enumGuardProof(t, source)
	for _, n := range statements {
		if n.Kind == ast.KindFunctionDeclaration && n.Name().Text() == "isAB" {
			_, members, ok := l.kindPredicateBody(n.Body(), n.Parameters()[0].Name())
			return ok, len(members)
		}
	}
	t.Fatal("missing helper predicate")
	return false, 0
}
func TestKindHelperShapesAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, p := range []struct{ name, helpers, body string }{
		{"default_chain", kindHelperChain, "return kinds(n.kind);"},
		{"guard", kindHelperChain, "return n && kinds(n.kind);"},
		{"double_guard", kindHelperChain, "return !!n && kinds(n.kind);"},
		{"disjunction", kindHelperLeaf, "return n.kind===K.A || leaf(n.kind);"},
		{"arrow", "const leaf=(k:K):boolean=>k===K.B;\n", "return n.kind===K.A || leaf(n.kind);"},
		{"return_chain", kindHelperLeaf + "function kinds(k:K):boolean { return k===K.A || leaf(k); }\n", "return kinds(n.kind);"},
		{"switch_tail", kindHelperLeaf + "function kinds(k:K):boolean { switch(k) {case K.A:return true;} return leaf(k); }\n", "return kinds(n.kind);"},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			parameter := "N"
			if p.name == "double_guard" {
				parameter = "N|undefined"
			}
			source := kindSetTypes + p.helpers + "function isAB(n:" + parameter + "):n is A|B {" + p.body + "}\n" + kindSetShow + kindSetValues
			if ok, count := kindHelperRecognized(t, source); !ok || count != 2 {
				t.Fatalf("kind-helper grammar %s: recognized=%v members=%d", p.name, ok, count)
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, out, errs := brandedNative(t, program); code != 0 || out != "alpha\n7\nfalse\n" {
				t.Fatalf("kind-helper %s: exit=%d stdout=%q stderr=%q", p.name, code, out, errs)
			}
			_, err := lowerSourceAs(t, source, ".a")
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf(".a helper verifier: %v", err)
			}
			if refused.What != "a type predicate whose return is not proven (return expression is not a trusted check on n)" {
				t.Fatalf(".a verifier changed: %s", refused.What)
			}
		})
	}
}

func TestKindHelperDefaultExcludesExplicitFalse(t *testing.T) {
	t.Parallel()
	const helpers = `function leaf(k:K):boolean { return k===K.A || k===K.B; }
function kinds(k:K):boolean { switch(k) { case K.A:return false; default:return leaf(k); } }
function isAB(n:N):n is B {return kinds(n.kind);}
function show(n:N):string { if(isAB(n))return String(n.b); if(n.kind===K.A)return n.a;return String(n.c); }
`
	source := kindSetTypes + helpers + kindSetValues
	if ok, count := kindHelperRecognized(t, source); !ok || count != 1 {
		t.Fatalf("kind-helper default exclusion: recognized=%v members=%d", ok, count)
	}
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, out, errs := brandedNative(t, program); code != 0 || out != "alpha\n7\nfalse\n" {
		t.Fatalf("default exclusion: exit=%d stdout=%q stderr=%q", code, out, errs)
	}
}

func kindHelperDepthChain(count int) string {
	source := ""
	for i := 0; i < count; i++ {
		body := "return k===K.A;"
		if i+1 < count {
			body = fmt.Sprintf("return h%d(k);", i+1)
		}
		source += fmt.Sprintf("function h%d(k:K):boolean {%s}\n", i, body)
	}
	return source
}
func TestKindHelperFallbacksRejectWholeBody(t *testing.T) {
	t.Parallel()
	for _, p := range []struct{ name, helpers, body string }{
		{"h05", "function kinds(k:K):boolean {switch(k) {case K.A:return true;case K.D:return false;default:return false;}}\n", "return n && kinds(n.kind);"},
		{"other_read", "const flag=true; function kinds(k:K):boolean {return k===K.A || flag;}\n", "return kinds(n.kind);"},
		{"non_helper", "function other(k:K):boolean {return k>0;} function kinds(k:K):boolean {return k===K.A || other(k);}\n", "return kinds(n.kind);"},
		{"written_parameter", "function kinds(k:K):boolean {k=K.A;return k===K.A;}\n", "return kinds(n.kind);"},
		{"default_parameter", "function kinds(k:K=K.A):boolean {return k===K.A;}\n", "return kinds(n.kind);"},
		{"wrong_field", kindHelperLeaf, "return n.kind===K.A || leaf(n.other);"},
		{"arithmetic_argument", kindHelperLeaf, "return n.kind===K.A || leaf(n.kind+0);"},
		{"two_enums", "enum L { A=1,B=2 } function kinds(k:K|L):boolean {return k===L.B;}\n", "return n.kind===K.A || kinds(n.kind);"},
		{"depth", kindHelperDepthChain(kindHelperMaxDepth + 1), "return h0(n.kind);"},
		{"cycle", "function a(k:K):boolean {return k===K.A || b(k);} function b(k:K):boolean {return a(k);}\n", "return a(n.kind);"},
		{"overload", "function kinds(k:K):boolean; function kinds(k:K):boolean {return k===K.A;}\n", "return kinds(n.kind);"},
		{"let_helper", "let kinds=(k:K):boolean=>k===K.A;\n", "return kinds(n.kind);"},
		{"alias_argument", kindHelperLeaf, "const kind=n.kind;return kind===K.A || leaf(kind);"},
		{"mixed_disjunction", kindHelperLeaf + "function block(n:N):boolean {return n.kind===K.A;}\n", "return leaf(n.kind) || block(n);"},
		{"isExpression", kindHelperLeaf + "function skip(n:N):N {return n;}\n", "return leaf(skip(n).kind);"},
		{"isStatement", kindHelperLeaf + "function block(n:N):boolean {return n.kind===K.A;}\n", "const kind=n.kind;return leaf(kind) || block(n);"},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			types := kindSetTypes
			if p.name == "h05" {
				types = strings.Replace(types, "C=3", "D=1", 1)
				types = strings.Replace(types, "readonly kind: K.C", "readonly kind: K.B", 1)
			}
			if p.name == "wrong_field" {
				types = strings.ReplaceAll(types, "readonly kind:", "readonly other:K; readonly kind:")
			}
			source := types + p.helpers + "function isAB(n:N):n is A|B {" + p.body + "}\n" + kindSetShow
			if ok, _ := kindHelperRecognized(t, source); ok {
				t.Fatalf("kind-helper fallback %s was recognized", p.name)
			}
			// isExpression: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestKindHelperFallbacksRejectWholeBody/isExpression".
			// isStatement: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestKindHelperFallbacksRejectWholeBody/isStatement".
		})
	}
}
func TestKindHelperLieStopsAtCheckedRead(t *testing.T) {
	t.Parallel()
	const source = `enum Kind { A,B,C }
interface Base {readonly kind:Kind}
interface XA extends Base {readonly kind:Kind.A;readonly a:string}
interface ZA extends Base {readonly kind:Kind.B;readonly a:string}
interface XB extends Base {readonly kind:Kind.C;readonly b:number}
interface YA extends Base {readonly kind:Kind.A;readonly a:number}
function selected(k:Kind):boolean {switch(k) {case Kind.A:case Kind.B:return true;default:return false;}}
function isText(n:Base):n is XA|ZA {return selected(n.kind);}
function guard(run:()=>string):string {try{return run();}catch(e){return 'threw';}}
type Inner=YA|XB;type Outer=Inner|XA|ZA;
function show(n:Outer):string {return isText(n)?guard(()=>String(n.a).toUpperCase()+typeof n.a):'other';}
const ya:YA={kind:Kind.A,a:3}; console.log(show(ya));
`
	observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
	if observed.code != 0 || string(observed.stdout) != "3number\n" {
		t.Fatalf("kind-helper lie Node: %+v", observed)
	}
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	brandedStopsProgram(t, program, "adamic: panic: field read failed: n.a is not a string")
}

// An unannotated variable may hold a boolean arrow. Without the arrow's return
// annotation TS infers a type predicate, which keeps its separate lowering wall.
func TestKindHelperInferredArrowKeepsPredicateBoundary(t *testing.T) {
	t.Parallel()
	source := kindSetTypes + "const leaf=(k:K)=>k===K.B;\nfunction isAB(n:N):n is A|B {return n.kind===K.A || leaf(n.kind); }\n" + kindSetShow + kindSetValues
	if ok, _ := kindHelperRecognized(t, source); !ok {
		t.Fatal("inferred arrow grammar lost recognition")
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "a call through a predicate signature nothing checks or proves" {
		t.Fatalf("inferred helper arrow boundary: %v", err)
	}
}
func TestKindHelperDepthBoundaryRecognizes(t *testing.T) {
	t.Parallel()
	source := kindSetTypes + kindHelperDepthChain(kindHelperMaxDepth) + "function isAB(n:N):n is A {return h0(n.kind); }\n"
	if ok, count := kindHelperRecognized(t, source); !ok || count != 1 {
		t.Fatalf("kind-helper depth boundary: recognized=%v members=%d", ok, count)
	}
}

// An empty default participates in fallthrough just like a nonempty one. It
// must replace the tail's fallback even when its inherited outcome is false.
func TestKindHelperEmptyDefaultTailRoutes(t *testing.T) {
	t.Parallel()
	for _, p := range []struct {
		name, body, node string
		recognized       bool
	}{
		{"empty_false", "switch(k) {default:case K.A:return false;} return leaf(k);", "undefined\n", false},
		{"empty_true", "switch(k) {default:case K.A:return true;} return leaf(k);", "true\n", false},
		{"default_last", "switch(k) {case K.A:return true;default:} return leaf(k);", "alpha\n7\nfalse\n", true},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			helpers := "function leaf(k:K):boolean {return k===K.B;}\nfunction kinds(k:K):boolean {" + p.body + "}\n"
			source := kindSetTypes + helpers + "function isAB(n:N):n is A|B {return n && kinds(n.kind); }\n" + kindSetShow + kindSetValues
			if !p.recognized {
				types := strings.Replace(kindSetTypes, "readonly a: string", "readonly a: string; readonly c: boolean", 1)
				source = types + helpers + "function isAB(n:N):n is B {return n && kinds(n.kind); }\nfunction show(n:N):string {if(isAB(n))return 'true';return String(n.c);}\nconst b:B={kind:K.B,b:7}; console.log(show(b));\n"
			}
			observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
			if observed.code != 0 || string(observed.stdout) != p.node {
				t.Fatalf("empty default %s Node: %+v", p.name, observed)
			}
			ok, count := kindHelperRecognized(t, source)
			if !p.recognized {
				if ok {
					t.Fatalf("kind-helper empty default resurrected an unreachable tail: %s members=%d", p.name, count)
				}
				// empty_false: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestKindHelperEmptyDefaultTailRoutes/empty_false".
				// empty_true: .a stops earlier at a type predicate whose return is not proven; its .ts side is enumTagSwingRows' "TestKindHelperEmptyDefaultTailRoutes/empty_true".
				return
			}
			if !ok || count != 2 {
				t.Fatalf("empty default reaching tail lost its kind set: recognized=%v members=%d", ok, count)
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, out, errs := brandedNative(t, program); code != 0 || out != p.node {
				t.Fatalf("empty default tail: exit=%d stdout=%q stderr=%q", code, out, errs)
			}
		})
	}
}
