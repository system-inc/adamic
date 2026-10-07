// Package refusalprobe writes programs which must be refused, and checks their sound neighbors.
package refusalprobe

// Entry names one construct, the diagnostic text it owns, and a single replacement which repairs it.
// Boundary records constructs whose refusal cannot be reached through load.Load, or which opened
// after 0.1. They remain visible in the catalog rather than masquerading as successful probes.
type Entry struct {
	Name, Diagnostic, Bad, Good, Placement, Boundary string
}

// Catalog is derived from refusals.go and the soundness/refusal tables in docs/0.1.md.
// Diagnostic is specific text, not the common "Adamic 0.1 refuses" prefix.
func Catalog() []Entry {
	return []Entry{
		{Name: "any", Diagnostic: "any", Bad: "let value: any = 1;", Good: "let value: number = 1;"},
		{Name: "cast", Diagnostic: "a cast the runtime can't check", Bad: "const value = 1; const cast = value as unknown as {readonly n: number};", Good: "const value = 1; const cast = {n: 1};"},
		{Name: "non-null", Diagnostic: "the non-null assertion !", Bad: "const value = new Map<string, number>().get('x')!;", Good: "const value = new Map<string, number>().get('x') ?? 0;"},
		{Name: "type-guard", Diagnostic: "a type predicate whose return is not proven (true return narrows to number, not number)", Bad: "function guard(x: number): x is number { return true; }", Good: "function guard(x: number): boolean { return typeof x === 'number'; }", Placement: "module"},
		{Name: "assertion-guard", Diagnostic: "a type predicate whose return is not proven (normal return has not narrowed x to number)", Bad: "function guard(x: number): asserts x is number {}", Good: "function guard(x: number): void {}", Placement: "module"},
		{Name: "definite-local", Diagnostic: "a definite assignment assertion !", Bad: "let value!: number; value = 1;", Good: "let value: number; value = 1;"},
		{Name: "definite-field", Diagnostic: "a definite assignment assertion !", Bad: "class Box { value!: number; constructor() { this.value = 1; } } const box = new Box();", Good: "class Box { value: number; constructor() { this.value = 1; } } const box = new Box();", Placement: "module"},
		{Name: "ts-ignore", Diagnostic: "@ts-ignore suppression directive", Bad: "// @ts-ignore\nconst value: number = 1;", Good: "const value: number = 1;"},
		{Name: "ts-expect-error", Diagnostic: "@ts-expect-error suppression directive", Bad: "// @ts-expect-error\nconst value: number = 'wrong';", Good: "const value: number = 1;"},
		{Name: "ts-nocheck", Diagnostic: "@ts-nocheck checking pragma", Bad: "// @ts-nocheck\nconst value: number = 1;", Good: "const value: number = 1;", Placement: "prefix"},
		{Name: "ts-check", Diagnostic: "@ts-check checking pragma", Bad: "// @ts-check\nconst value: number = 1;", Good: "const value: number = 1;", Placement: "prefix"},
		{Name: "async", Diagnostic: "an async function", Bad: "async function wait(): Promise<void> {}", Good: "function wait(): void {}", Placement: "module"},
		{Name: "generator", Diagnostic: "a generator function", Bad: "function* sequence() {}", Good: "function sequence(): void {}", Placement: "module"},
		{Name: "await", Diagnostic: "await", Bad: "export const value = await 1;", Good: "export const value = 1;", Placement: "module"},
		{Name: "yield", Diagnostic: "yield (generators)", Boundary: "yield requires a generator, which is refused first"},
		{Name: "decorator", Diagnostic: "a decorator", Bad: "function decorate<T>(value: T): void {} @decorate class Box { n = 1; } const box = new Box();", Good: "function decorate<T>(value: T): void {} class Box { n = 1; } const box = new Box();", Placement: "module"},
		{Name: "label", Diagnostic: "a label", Bad: "outer: while (false) { break outer; }", Good: "while (false) { break; }"},
		{Name: "with", Diagnostic: "with", Boundary: "loader rejects with in strict modules (TS1101)"},
		{Name: "delete", Diagnostic: "delete", Bad: "const box: {n?: number} = {n: 1}; delete box.n;", Good: "const box: {n?: number} = {n: 1};"},
		{Name: "debugger", Diagnostic: "debugger", Bad: "debugger;", Good: "const value = 1;"},
		{Name: "enum", Diagnostic: "enum", Boundary: "loader erasableSyntaxOnly rejects enum (TS1294) before lowering"},
		{Name: "namespace", Diagnostic: "a namespace", Bad: "namespace Types { export type N = number; }", Good: "type N = number;", Placement: "module"},
		{Name: "void", Diagnostic: "the void operator", Bad: "void 1;", Good: "const value = 1;"},
		{Name: "index-signature", Diagnostic: "an index signature", Bad: "interface Indexed { [key: string]: number; }", Good: "interface Indexed { readonly n: number; }", Placement: "module"},
		{Name: "export-default", Diagnostic: "export default", Bad: "export default 1;", Good: "export const value = 1;", Placement: "module"},
		{Name: "equal", Diagnostic: "==", Bad: "const value = 1 == 1;", Good: "const value = 1 === 1;"},
		{Name: "unequal", Diagnostic: "!=", Bad: "const value = 1 != 1;", Good: "const value = 1 !== 1;"},
		{Name: "in", Diagnostic: "in", Bad: "const value = 'n' in {n: 1};", Good: "const value = {n: 1}.n === 1;"},
		{Name: "comma", Diagnostic: "the comma operator", Bad: "let n = 1; const value = (n += 1, n);", Good: "let n = 1; n += 1; const value = n;"},
		{Name: "and-assign", Diagnostic: "&&=", Bad: "let value: boolean = true; value &&= false;", Good: "let value: boolean = true; if (value) { value = false; }"},
		{Name: "or-assign", Diagnostic: "||=", Bad: "let value: boolean = false; value ||= true;", Good: "let value: boolean = false; if (!value) { value = true; }"},
		{Name: "arguments", Diagnostic: "arguments", Bad: "function count(): number { return arguments.length; } console.log(`${count()}`);", Good: "function count(): number { return 0; } console.log(`${count()}`);", Placement: "module"},
		{Name: "unbound-method", Diagnostic: "a method read as a value", Bad: "const items: number[] = []; const method = items.push;", Good: "const items: number[] = []; const method = (n: number): number => items.push(n);"},
		{Name: "mutable-variance", Diagnostic: "adamic/invariant-mutable", Bad: "const narrow: {n: number; s: string}[] = []; const wide: {n: number}[] = narrow;", Good: "const narrow: {n: number; s: string}[] = []; const wide: readonly {n: number}[] = narrow;"},
		{Name: "bivariant-method", Diagnostic: "method-signature-style", Bad: "interface A { n: number; } interface D extends A { s: string; } interface Narrow { handle(x: D): void; } interface Wide { handle(x: A): void; } function widen(x: Narrow): Wide { return x; }", Good: "interface A { n: number; } interface D extends A { s: string; } interface Narrow { handle(x: A): void; } interface Wide { handle(x: A): void; } function widen(x: Narrow): Wide { return x; }", Placement: "module"},
		{Name: "nominal-class", Diagnostic: "adamic/nominal-class", Bad: "class Box { n = 1; } const box: Box = {n: 1};", Good: "class Box { n = 1; } const box: Box = new Box();", Placement: "module"},
		{Name: "expando", Diagnostic: "properties added after creation", Bad: "function value() {} value.extra = 1;", Good: "function value() {}", Placement: "module"},
		{Name: "spread", Diagnostic: "a spread after the first field", Bad: "const source = {n: 1}; const view: {} = source; const value = {n: 2, ...view};", Good: "const source = {n: 1}; const view: {} = source; const value = {...source, n: 2};"},
		{Name: "prototype-literal", Diagnostic: "__proto__ in an object literal", Bad: "const value = {__proto__: {n: 1}};", Good: "const value = {n: 1};"},
		{Name: "define-property", Diagnostic: "Object.defineProperty", Bad: "Object.defineProperty({n: 1}, 'n', {value: 'wrong'});", Good: "const value = {n: 1};"},
		{Name: "prototype-mutation", Diagnostic: "Object.setPrototypeOf", Bad: "Object.setPrototypeOf({n: 1}, {});", Good: "const value = {n: 1};"},
		{Name: "var", Diagnostic: "var", Bad: "var value = 1;", Good: "let value = 1;"},
		{Name: "truthiness", Diagnostic: "a number as a condition", Bad: "if (1) { console.log('ok'); }", Good: "if (1 === 1) { console.log('ok'); }"},
		{Name: "random", Diagnostic: "Math.random", Bad: "const value = Math.random();", Good: "const value = 0.5;"},
		{Name: "eval", Diagnostic: "eval", Bad: "eval('1');", Good: "const value = 1;"},
		{Name: "function-type", Diagnostic: "Function", Bad: "function take(value: Function): void {}", Good: "function take(value: () => void): void {}", Placement: "module"},
		{Name: "new-function", Diagnostic: "Function", Bad: "const value = new Function('return 1');", Good: "const value = (): number => 1;"},
		{Name: "record", Diagnostic: "index signature", Bad: "const value: Record<string, number> = {};", Good: "const value = new Map<string, number>();"},
		{Name: "optional-widening", Diagnostic: "adamic/no-optional-widening", Bad: "const original = {x: 1, y: 'wrong'}; const view: {x: number} = original; const wider: {x: number; y?: number} = view;", Good: "const original = {x: 1}; const view: {x: number} = original; const wider: {x: number} = view;"},
		{Name: "merging", Diagnostic: "declaration merging", Bad: "class Box { n = 1; } interface Box { extra: number; } const box = new Box();", Good: "class Box { n = 1; } const box = new Box();", Placement: "module"},
		{Name: "constructor-escape", Diagnostic: "this escaping a constructor before every field is set", Bad: "class Box { n: number; constructor() { this.read(); this.n = 1; } read(): number { return this.n; } } const box = new Box();", Good: "class Box { n: number; constructor() { this.n = 1; this.read(); } read(): number { return this.n; } } const box = new Box();", Placement: "module"},
		{Name: "prototype-read", Diagnostic: "isPrototypeOf", Bad: "const value = {}.isPrototypeOf;", Good: "const value = (n: number): boolean => n === 1;"},
		// provePredicate refuses a type predicate whose body does not prove it. An arrow predicate is an
		// expression, so the writer can place it in every surrounding; the neighbor is the proven predicate.
		{Name: "unproven-predicate", Diagnostic: "a type predicate whose return is not proven (true return narrows to string | undefined, not string)", Bad: "const isText = (x: string | undefined): x is string => true; const word: string | undefined = 'a'; if (isText(word)) { console.log(word); }", Good: "const isText = (x: string | undefined): x is string => x !== undefined; const word: string | undefined = 'a'; if (isText(word)) { console.log(word); }"},
		{Name: "method-override", Diagnostic: "adamic/contravariant-override", Boundary: "checkOverrides also guards inheritance member kinds, accessor descriptors and ABI representation; not generated in this 0.1 corpus"},
		{Name: "parameter-properties", Boundary: "loader erasableSyntaxOnly rejects parameter properties (TS1294)"},
		{Name: "overloads", Boundary: "implementation signatures are not lowered as overloads; needs a separate diagnostic design"},
		{Name: "reopened", Boundary: "throw Error/try/catch/finally, inheritance/super/abstract/protected, accessors, for-in on proven plain objects, RegExp/Set/JSON.stringify, Object helpers, arguments/files input opened after 0.1; not promised Refused by current compiler"},
		{Name: "library-boundaries", Boundary: "symbol/bigint/Date/process/globalThis/Intl and stdin/environment need individual loader/NotYet contracts; not generated"},
		{Name: "polymorphic-recursion", Boundary: "requires recursive generic instantiation; not generated"},
		{Name: "module-boundaries", Boundary: "namespace imports, export-star, default declarations and import cycles need multi-module probes; not generated"},
	}
}
