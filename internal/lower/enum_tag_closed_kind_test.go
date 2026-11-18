package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsc's node shapes, small: Node's kind is the whole open enum K, and each branded supertype (Ex for
// Expression, Lit for LiteralExpression, PropLike for JSDocPropertyLikeTag, Stmt for Statement) carries
// a phantom brand typed undefined, so a node of another family isn't assignable to it, and leaves with
// literal kinds. Every refinement of a union holding one of them is the walk's: its kind is open, so
// enumTagHazardOf calls it a hazard, and closedKinds decides (enum_tag_closed_kind.go, #x2cjvy9, Adamic's
// ruling of Nov 17). The predicates live in each row, since one .a can't prove refuses its program.
const closedKindTypes = `enum K { Body = 1, Ident = 2, Num = 3, Str = 4, Big = 5, Return = 6, Prop = 7, Param = 8, If = 9, Ret = 10, Computed = 11, Bin = 12 }
interface Node { readonly kind: K; readonly pos: number }
interface Body extends Node { readonly kind: K.Body; readonly b: string }
interface Ex extends Node { readonly _expressionBrand: undefined }
interface Ident extends Ex { readonly kind: K.Ident; readonly name: string }
interface Lit extends Ex { readonly _literalBrand: undefined; readonly text: string }
interface NumLit extends Lit { readonly kind: K.Num }
interface StrLit extends Lit { readonly kind: K.Str }
interface BigLit extends Lit { readonly kind: K.Big }
interface Tag extends Node { readonly tagName: string }
interface ReturnTag extends Tag { readonly kind: K.Return; readonly typeText: string }
interface PropLike extends Tag { readonly _declarationBrand: undefined; readonly bracketed: boolean }
interface PropTag extends PropLike { readonly kind: K.Prop }
interface ParamTag extends PropLike { readonly kind: K.Param }
interface Stmt extends Node { readonly _statementBrand: undefined; readonly line: number }
interface IfStmt extends Stmt { readonly kind: K.If; readonly test: string }
interface RetStmt extends Stmt { readonly kind: K.Ret }
const body: Body = { kind: K.Body, pos: 0, b: 'block' };
const ident: Ident = { kind: K.Ident, pos: 1, name: 'x', _expressionBrand: undefined };
`

// isBody is tsc's isBlock over ConciseBody, FunctionBody | Expression, as Body | Ex.
const closedKindIsBody = "function isBody(n: Body | Ex): n is Body { return n.kind === K.Body; }\n"

// The ConciseBody rows, which TestClosedKindDotA holds in .a too. Each predicate row reads only its
// true side: R-T80 keeps the hazard (predicates_checked.go), so a read on a predicate's false side over
// an open S is NotYet there, whatever the walk proves.
var closedKindConciseBody = []struct{ name, source, stdout string }{
	{"concise body predicate", closedKindIsBody + "function show(x: Body | Ex): string { return isBody(x) ? x.b : 'expression'; }\nconsole.log(show(body));\nconsole.log(show(ident));\n", "block\nexpression\n"},
	{"concise body comparison", "function show(x: Body | Ex): string { return x.kind !== K.Body ? `expression ${x.pos}` : 'body'; }\nconsole.log(show(body));\nconsole.log(show(ident));\n", "body\nexpression 1\n"},
}

// A closed supertype's narrowing lowers and agrees with Node, on the JavaScript backend and sanitized
// natively, one row per shape tsc's refusals met: ConciseBody's, LiteralExpression's with three kinds,
// JSDocPropertyLikeTag's with two leaves, and Statement's in a switch, whose default clause drops the
// one case's member. Mutants (scratchpad m3): K(S) taking every declared type's kind, assignable or not,
// puts Body, Ident and Return in the domains, and each predicate row refuses; the comparison and switch
// rows hold, since a comparison keeps S on both sides and never reads K(S).
func TestClosedKindNarrowingLowers(t *testing.T) {
	t.Parallel()
	probes := append([]struct{ name, source, stdout string }{}, closedKindConciseBody...)
	probes = append(probes, []struct{ name, source, stdout string }{
		{"literal expression", "function isIdent(n: Ident | Lit): n is Ident { return n.kind === K.Ident; }\n" +
			"function show(x: Ident | Lit): string { return isIdent(x) ? `id ${x.name}` : 'literal'; }\n" +
			"function text(x: Ident | Lit): string { return x.kind !== K.Ident ? x.text : 'id'; }\n" +
			"const num: NumLit = { kind: K.Num, pos: 2, text: '1', _expressionBrand: undefined, _literalBrand: undefined };\n" +
			"console.log(show(ident));\nconsole.log(show(num));\nconsole.log(text(num));\nconsole.log(text(ident));\n", "id x\nliteral\n1\nid\n"},
		{"property-like tag", "function isReturnTag(n: ReturnTag | PropLike): n is ReturnTag { return n.kind === K.Return; }\n" +
			"function show(x: ReturnTag | PropLike): string { return isReturnTag(x) ? x.typeText : 'tag'; }\n" +
			"function bracketed(x: ReturnTag | PropLike): string { return x.kind !== K.Return ? String(x.bracketed) : 'return'; }\n" +
			"const ret: ReturnTag = { kind: K.Return, pos: 3, tagName: 'returns', typeText: 'number' };\n" +
			"const prop: PropTag = { kind: K.Prop, pos: 4, tagName: 'property', _declarationBrand: undefined, bracketed: true };\n" +
			"console.log(show(ret));\nconsole.log(show(prop));\nconsole.log(bracketed(prop));\nconsole.log(bracketed(ret));\n", "number\ntag\ntrue\nreturn\n"},
		{"statement switch", "function show(x: Ident | Stmt): string { switch (x.kind) { case K.Ident: return 'ident'; default: return `line ${x.line}`; } }\n" +
			"const ifs: IfStmt = { kind: K.If, pos: 5, line: 7, test: 'x', _statementBrand: undefined };\n" +
			"console.log(show(ident));\nconsole.log(show(ifs));\n", "ident\nline 7\n"},
	}...)
	for _, probe := range probes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, closedKindTypes+probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// closedKindsAt loads source as main.ts and asks closedKinds of kind on the interface or type alias it
// declares as name, returning the lowering it asked through and the file's statements, so a pin can name
// a site by the program's own Where.
func closedKindsAt(t *testing.T, source, name string) (*lowering, []*ast.Node, *checker.Type, bool, string) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	t.Cleanup(release)
	l := &lowering{program: loaded, checker: checked, result: &ir.Program{}}
	statements := loaded.Files()[0].Statements.Nodes
	for _, statement := range statements {
		if (statement.Kind == ast.KindInterfaceDeclaration || statement.Kind == ast.KindTypeAliasDeclaration) && statement.Name().Text() == name {
			kinds, closed, reason := l.closedKinds(l.checker.GetDeclaredTypeOfSymbol(l.checker.GetSymbolOfDeclaration(statement)), "kind")
			return l, statements, kinds, closed, reason
		}
	}
	t.Fatalf("no interface or type alias %s", name)
	return nil, nil, nil, false, ""
}

// K(S) is exactly the literal kinds of the declared types assignable to S, read through closedKinds:
// LiteralExpression's three leaves, JSDocPropertyLikeTag's two. Mutants: m3 adds kinds to both, and m4
// (closedKinds always closed and empty) empties both.
func TestClosedKindDomains(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name string
		size int
	}{{"Lit", 3}, {"PropLike", 2}} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, _, kinds, closed, reason := closedKindsAt(t, closedKindTypes, probe.name)
			if !closed {
				t.Fatalf("want %s closed, got open: %s", probe.name, reason)
			}
			if kinds.Flags()&checker.TypeFlagsUnion == 0 || len(kinds.Types()) != probe.size {
				t.Fatalf("want %s's kinds %d, got %s", probe.name, probe.size, l.checker.TypeToString(kinds))
			}
		})
	}
}

// A generic factory builds its S at whatever kind it's handed: wrap<T>(kind: T, ...) returns
// `{ kind, text }`, whose type `{ kind: T; text: string }` isn't assignable to Expr while T is a type
// parameter, yet wrap(Kind.Block, ...) typed Expr is an Expr with Block's kind. Its permissive instantiation
// reaches Expr, so Expr stays open and isBlock's narrowing over Block | Expr stays refused, held in a
// container and a field too (Language's k01 and k02 on #x2cjvy9, which lowered at 53c2277e and stopped where
// Node ran on). decodeJson<Expr> builds an Expr whose kind its text holds, and keeps Expr open the same way.
// Mutant m6 (the permissive instantiation skipped) lowers the generic rows.
func TestClosedKindGenericAndDecodedBuildsStayOpen(t *testing.T) {
	t.Parallel()
	const types = "import { decodeJson } from 'adamic';\nenum Kind { Block = 1, Text = 2, Call = 3 }\ninterface Node { readonly kind: Kind }\ninterface Block extends Node { readonly kind: Kind.Block; readonly count: number }\ninterface Expr extends Node { readonly text: string }\nfunction isBlock(node: Node): node is Block { return node.kind === Kind.Block; }\nfunction show(body: Block | Expr): string { return isBlock(body) ? String(body.count) : body.text; }\nconst block: Block = { kind: Kind.Block, count: 2 };\nconsole.log(show(block));\n"
	const wrap = "function wrap<T>(kind: T, text: string): { kind: T; text: string } { return { kind, text }; }\n"
	for _, probe := range []struct{ name, source string }{
		{"generic literal", wrap + "const lie: Expr = wrap(Kind.Block, 'xyz');\nconsole.log(show(lie));\n"},
		{"generic in a container", wrap + "const held: Expr[] = [];\nheld.push(wrap(Kind.Block, 'xyz'));\nfor (const lie of held) { console.log(show(lie)); }\n"},
		{"generic in a field", wrap + "interface Holder { body: Expr }\nconst holder: Holder = { body: wrap(Kind.Text, 'ok') };\nholder.body = wrap(Kind.Block, 'xyz');\nconsole.log(show(holder.body));\n"},
		{"decoded", "const decoded = decodeJson<Expr>('{\"kind\":1,\"text\":\"xyz\"}');\nif (decoded.kind === 'Ok') console.log(show(decoded.value));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, types+probe.source, ".a"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
				t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
			}
		})
	}
}

// A one-member enum's member is that enum's whole type, so a kind typed One.A names no layout by its value
// alone, and a supertype it would close stays open (Adamic's row on #x2cjvy9, Nov 17, after Language's
// J-pmbsp94). Mutant m7 (closedKindLiterals reading such a member as a literal) closes it.
func TestClosedKindOneMemberEnumStaysOpen(t *testing.T) {
	t.Parallel()
	const source = "enum One { A }\ninterface Base { readonly kind: One; readonly _baseBrand: undefined; readonly pos: number }\ninterface Leaf extends Base { readonly kind: One.A; readonly leaf: string }\n"
	if _, _, _, closed, reason := closedKindsAt(t, source, "Base"); closed {
		t.Fatalf("want Base open, since its kind is a one-member enum's, got closed, reason %q", reason)
	}
}

// A Proxy of an Ident answers whatever kind its get trap chooses, so its build keeps Ex open even though its
// type says Ident (Language's should on #x2cjvy9). Today `new Proxy` doesn't lower; this holds the rule when
// it does.
func TestClosedKindProxyIsAnOpenBuild(t *testing.T) {
	t.Parallel()
	if _, _, _, closed, reason := closedKindsAt(t, closedKindTypes+"const proxied: Ident = new Proxy<Ident>(ident, {});\n", "Ex"); closed || !strings.Contains(reason, "a new Proxy") {
		t.Fatalf("want Ex open on a new Proxy, got closed %v, reason %q", closed, reason)
	}
}

// A write of kind through any, tsc's clone loop in nodeFactory.ts, leaves S closed: lowering refuses the
// cast to any, so the program never compiles and the write makes no value (closedKindReaches, J-q7nwd21).
// The same write through a mutable Lit, typed, keeps Lit open. Mutant m5 (any reaches every S) opens the first row.
func TestClosedKindWriteThroughAnyReachesNothing(t *testing.T) {
	t.Parallel()
	const clone = "function clone(n: Ex, c: Ex): Ex {\n  for (const key in n) { (c as any)[key] = (n as any)[key]; }\n  return c;\n}\n"
	if l, _, kinds, closed, reason := closedKindsAt(t, closedKindTypes+clone, "Lit"); !closed || kinds.Flags()&checker.TypeFlagsUnion == 0 || len(kinds.Types()) != 3 {
		t.Fatalf("want Lit closed with its 3 kinds beside a write through any, got closed %v, kinds %s, reason %q", closed, l.checker.TypeToString(kinds), reason)
	}
	const typed = "function retag(e: Mutable): void { e.kind = numeric(); }\ntype Mutable = { -readonly [P in keyof Lit]: Lit[P] };\nfunction numeric(): K { return K.Body; }\n"
	if _, _, _, closed, reason := closedKindsAt(t, closedKindTypes+typed, "Lit"); closed || !strings.Contains(reason, "a write of kind") {
		t.Fatalf("want Lit open on a typed write of kind, got closed %v, reason %q", closed, reason)
	}
}

// An unbranded supertype Open, with a class or a function that builds a value assignable to it with a
// kind chosen at run time, stays open: Open's value of kind Body would have Open's layout. So isBody's
// narrowing over Body | Open keeps the open-enum refusal. The control has no such build, and lowers.
// Mutants: m1 (build sites skipped) and m2 (open builds ignored) lower the class and function rows; m4
// lowers them too.
func TestClosedKindRunTimeKindStaysOpen(t *testing.T) {
	t.Parallel()
	const open = "interface Open extends Node { readonly n: number }\n" +
		"function isOpenBody(x: Body | Open): x is Body { return x.kind === K.Body; }\n" +
		"function show(x: Body | Open): string { return isOpenBody(x) ? x.b : 'open'; }\n" +
		"function kb(): K { return K.Ident; }\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"class", "class N { constructor(readonly kind: K, readonly pos: number, readonly n: number) {} }\nconsole.log(show(new N(kb(), 0, 1)));\n", ""},
		{"function", "function make(kind: K): Open { return { kind, pos: 0, n: 1 }; }\nconsole.log(show(make(kb())));\n", ""},
		{"control", "const o: Open = { kind: K.Ident, pos: 0, n: 1 };\nconsole.log(show(body));\nconsole.log(show(o));\n", "block\nopen\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + open + probe.source
			if probe.stdout != "" {
				program := lowersAndAgreesWithNodeAs(t, source, ".ts")
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
				}
				return
			}
			if _, err := lowerSourceAs(t, source, ".ts"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
				t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
			}
		})
	}
	// closedKinds' reason names the class, the first site in the program that leaves Open open. m1, m2 and
	// m4 call Open closed.
	l, statements, _, closed, reason := closedKindsAt(t, closedKindTypes+open+"class N { constructor(readonly kind: K, readonly pos: number, readonly n: number) {} }\n", "Open")
	var class *ast.Node
	for _, statement := range statements {
		if statement.Kind == ast.KindClassDeclaration {
			class = statement
		}
	}
	if closed || class == nil || !strings.Contains(reason, l.program.Where(class)) {
		t.Fatalf("want Open open with a reason naming the class, got closed %v, reason %q", closed, reason)
	}
}

// What makes K(S) hold the compared kind keeps the narrowing refused: an anonymous literal typed Ex with
// kind Body puts Body in K(Ex) (a Body-kinded value with Ex's layout), and so does a second declared type
// with Body's kind, a twin layout. The predicate's type drops Ex, the comparison its body makes keeps it,
// and the identity check refuses. Mutants: m1 lowers the anonymous literal, and m4 lowers both.
func TestClosedKindHoldingTheComparedKindRefuses(t *testing.T) {
	t.Parallel()
	const show = closedKindIsBody + "function show(x: Body | Ex): string { return isBody(x) ? x.b : 'expression'; }\nconsole.log(show(body));\n"
	for _, probe := range []struct{ name, source string }{
		{"anonymous literal", "const e: Ex = { kind: K.Body, pos: 2, _expressionBrand: undefined };\nconsole.log(show(e));\n"},
		{"twin layout", "interface Twin extends Ex { readonly kind: K.Body; readonly t: number }\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, closedKindTypes+show+probe.source, ".ts"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
				t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
			}
		})
	}
}

// tsc's LateBoundDeclaration, as the brief writes it: DynamicNamedDeclaration, whose kind is the open
// enum, intersected with a LateBoundName name. The rule closes it with K(LateBound) empty, and that's
// right. The rule asks what's assignable to S, never what S is assignable to: LateBound is assignable to
// its base DynamicNamed, but no declared type with a literal kind is assignable to LateBound, since every
// declaration's name, PropDecl's here, is wider than LateName, and the base's is too. Nor does the program
// build one. So every LateBound value tsc holds came through a hatch, isLateBound or hasLateBindableName
// (checker.ts:13809), and is a checked view by half 2. The narrowing by isBinaryExpression drops it on the
// true side, as lateBindMember's `isBinaryExpression(decl) ? decl.left : decl.name` does. Mutant: m3 puts
// Bin, a declared literal kind not assignable to LateBound, in K(LateBound), and the row refuses.
func TestClosedKindLateBoundDeclaration(t *testing.T) {
	t.Parallel()
	const lateBound = `interface Decl extends Node { readonly _declarationBrand: undefined }
interface Computed extends Node { readonly kind: K.Computed; readonly expression: Ident | NumLit }
interface LateName extends Computed { readonly expression: Ident }
interface Named extends Decl { readonly name?: Ident | Computed }
interface DynamicNamed extends Named { readonly name: Computed }
type LateBound = DynamicNamed & { readonly name: LateName };
interface PropDecl extends Named { readonly kind: K.Prop; readonly name: Ident | Computed }
interface Bin extends Node { readonly kind: K.Bin; readonly left: string }
function isBin(n: Node): n is Bin { return n.kind === K.Bin; }
function show(d: LateBound | Bin): string { return isBin(d) ? d.left : 'declaration'; }
const bin: Bin = { kind: K.Bin, pos: 6, left: 'x' };
console.log(show(bin));
`
	l, _, kinds, closed, reason := closedKindsAt(t, closedKindTypes+lateBound, "LateBound")
	if !closed {
		t.Fatalf("want LateBound closed, got open: %s", reason)
	}
	if kinds.Flags()&checker.TypeFlagsNever == 0 {
		t.Fatalf("want LateBound's kinds never, got %s", l.checker.TypeToString(kinds))
	}
	program := lowersAndAgreesWithNodeAs(t, closedKindTypes+lateBound, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "x\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "x\n")
	}
}

// The ConciseBody rows in .a. The comparison is the walk's alone, and lowers and agrees as in .ts. The
// predicate is refused before the walk is asked: .a proves isBody's body or refuses it, and R-T80's prover
// keeps the hazard by ruling (predicates_proof.go), so the comparison in its body proves no layout there.
func TestClosedKindDotA(t *testing.T) {
	t.Parallel()
	for _, probe := range closedKindConciseBody {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + probe.source
			if strings.Contains(probe.source, "isBody(") {
				if _, err := lowerSourceAs(t, source, ".a"); err == nil || !strings.Contains(err.Error(), "the comparison of kind proves nothing about the layout") {
					t.Fatalf("want R-T80's lying-tag refusal, got %v", err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".a")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// Half 2 of the premise, for Language to extend: a value built at an open-kind class Raw, which Ex's brand
// keeps from being assignable to Ex, so K(Ex) stays closed, is carried into Body | Ex by a .ts predicate
// hatch with kind Body, stored in a field or an array, and read back as Body after isBody. It's a checked
// view all the way (rulings.md:18), so sanitized native and the JavaScript backend stop at the read with
// exit 70, where Node prints undefined. Its shape is the hinge probe's either road (bc8433ad). In .a the
// hatch itself is refused unproven.
func TestClosedKindHatchStaysChecked(t *testing.T) {
	t.Parallel()
	const hatch = closedKindIsBody + `class Raw { constructor(readonly kind: K, readonly pos: number) {} }
function kb(): K { return K.Body; }
function isEx(n: Raw): n is Ex { return n.pos >= 0; }
const raw = new Raw(kb(), 1);
const v: Body | Ex = isEx(raw) ? raw : body;
`
	for _, probe := range []struct{ name, source string }{
		{"field", "interface Holder { readonly v: Body | Ex }\nfunction showHeld(h: Holder): string { return isBody(h.v) ? h.v.b : 'other'; }\nconst held: Holder = { v };\nconsole.log(showHeld(held));\n"},
		{"array", "function showAll(all: readonly (Body | Ex)[]): void { for (const n of all) { console.log(isBody(n) ? n.b : 'other'); } }\nshowAll([v]);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := closedKindTypes + hatch + probe.source
			if observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); observed.code != 0 || string(observed.stdout) != "undefined\n" {
				t.Fatalf("Node: exit %d, stdout %q; want the lie read as undefined", observed.code, observed.stdout)
			}
			program, err := lowerSourceAs(t, source, ".ts")
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			brandedStopsProgram(t, program, "adamic: panic: ")
			if _, err := lowerSourceAs(t, source, ".a"); err == nil {
				t.Fatal("want the .a hatch refused unproven")
			}
		})
	}
}
