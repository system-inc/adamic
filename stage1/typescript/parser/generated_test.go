package parser

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Operator pairs hold every associativity and precedence boundary, not just
// a snapshot of the current precedence table. Spellings are independent of it.
func generatedExpressions() []string {
	operators := []string{"??", "||", "&&", "|", "^", "&", "==", "!=", "===", "!==", "<", ">", "<=", ">=", "instanceof", "in", "<<", ">>", ">>>", "+", "-", "*", "/", "%", "**"}
	cases := []string{
		"delete obj.x; void x; typeof x; +x; -x; ~x; !x; ++x; --x; x++; x--;",
		"const f = async (x = await (f())) => await x;",
		"const f = function* () { yield; yield 1; yield* xs; };",
		"const f = async function() { return await (f()); };",
		"f<T>?.(); f<T>?.[x]; f<T> ? x : y; ({x: f<T>});",
		"class C { #x = 1; m(obj) { return #x in obj && this.#x; } }",
		"const f = (x: T = a in b) => x; for(x = (a in b); x; x++) f();",
		"for(x = [(a in b)]; x; x++) f(); for(x = {a: b in c}; x; x++) f();",
		"const f = () => { label: while(x) { break label; } debugger; };",
		"const f = <T extends {x: U} = {x: V}>(x: T) => x;",
		"tag`a\\n`; `a\\u{1f600}`;",

		"class C<T> extends Base<T[]> { field = 1; method(x = 2) { return x; } }",
		"const enum E { A = 1, B = 2 }",
		"type A = false | true | null | -1; interface B extends A { readonly [Symbol.iterator]: typeof ns.value; pos: -1; }",
		"type I<U> = (U extends any ? (k: U) => void : never) extends ((k: infer I) => void) ? I : never;",
		"const f = (member): member is A & { body: B; } => test(member);",
		"function f(readonly = false, type = 1) { type = readonly ? 1 : type < 2 ? 3 : 4; }",
		"const f = (a: abstract new (...args: any[]) => object) => a;",
		"const f = (x: { [K in keyof T as `get${K & string}`]?: T[K] }) => x;",
		"const f = (x: [a: T, b?: U, ...rest: V[]]): typeof import('x') => x;",
		"const f = (x: typeof import('x').Foo<T>) => x;",
		"const f = ([, value]) => value;",
		"const f = () => { for(let i = 0; i < 3; i++) { if(i) continue; } for(const x of xs) f(x); for(x in xs) f(x); do {x++;} while(x < 3); switch(x) { case 1: return x; default: break; } try { f(); } catch(e) { throw e; } finally { f(); } };",
		"const c = class extends Base<T[]> implements A { declare readonly x: T; static { f(); } constructor(x = 1) { f(x); } get y() { return this.x; } set y(v) { this.x = v; } };",
		"a?.b<T>(); a?.b!<T>(); a?.b<T>`x`;",
		"const f = (a = /[)]/g) => a;",
		"const f = (a = `x${(f())})`) => a;",
		"a ? (b ? c : d) : (x): T => x;",
		"for(x = f(a in b); x; x++) f();",
		"({async = 1, get: x, set() {return x;}});",
		"async type => type;",
		"(a as T) * b; (a + b as T) * c;",
		"f<`a${T}`>();",
		"f<{x?: T}>();",
		"const f = (x = /a\\/b/gi) => x;",
	}
	for _, left := range operators {
		for _, right := range operators {
			cases = append(cases, fmt.Sprintf("a %s b %s c;", left, right))
		}
	}
	for _, operator := range []string{"=", "+=", "-=", "*=", "/=", "%=", "**=", "<<=", ">>=", ">>>=", "&=", "|=", "^=", "&&=", "||=", "??="} {
		cases = append(cases, fmt.Sprintf("a %s b %s c;", operator, operator))
	}
	random := rand.New(rand.NewSource(720))
	var expression func(int) string
	expression = func(depth int) string {
		if depth == 0 {
			values := []string{"x", "123", "0xff", "1_000n", "'héllo😀'", "true", "null", "this.x", "/[a-z]+/gi", "`a\\n`"}
			return values[random.Intn(len(values))]
		}
		switch random.Intn(12) {
		case 0:
			return "(" + expression(depth-1) + ")"
		case 1:
			return "(" + expression(depth-1) + " " + operators[random.Intn(10)] + " " + expression(depth-1) + ")"
		case 2:
			return "(" + expression(depth-1) + " ? " + expression(depth-1) + " : " + expression(depth-1) + ")"
		case 3:
			return "f(" + expression(depth-1) + ", ...xs)"
		case 4:
			return "[" + expression(depth-1) + ",, ...xs,]"
		case 5:
			return "({value: " + expression(depth-1) + ", [key]: x, ...rest})"
		case 6:
			return "((x: T) => " + expression(depth-1) + ")"
		case 7:
			return "(" + expression(depth-1) + " as T)"
		case 8:
			return "(" + expression(depth-1) + " satisfies T)"
		case 9:
			return "a?.b?.[" + expression(depth-1) + "]?.(x)"
		case 10:
			return "new Foo<T>(" + expression(depth-1) + ")"
		default:
			return "`head${" + expression(depth-1) + "}tail`"
		}
	}
	for i := 0; i < 1000; i++ {
		cases = append(cases, "const generated = "+expression(4)+";")
	}
	return cases
}

// Not parallel: native.runtimeBuilds and the runtime archive cache at os.UserCacheDir()/adamic/runtime via native.Build.
func TestGeneratedExpressionsAgree(t *testing.T) {
	cases := generatedExpressions()
	directory := t.TempDir()
	var manifest strings.Builder
	for i, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("generated-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", path)
	absolute, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	got := node(t, absolute, path, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = execute(t, "", buildPort(t, absolute, true), "--manifest", path)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}
	t.Logf("%d generated inputs, seed 720, %d identical tree bytes", len(cases), len(want.output))
}
