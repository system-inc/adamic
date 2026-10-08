// Package refusalprobe writes programs which must be refused, and checks their sound neighbors.
package refusalprobe

// Entry names a construct and its neighbor. Accepted pairs require both inputs to compile;
// other executable pairs require Bad to have Diagnostic and Good to compile.
// Boundary records tested NotYet facilities, constructs whose refusal cannot be reached through load.Load, or which opened
// after 0.1, and ruled refusals still awaiting compiler work. They remain visible rather than becoming successes.
type Entry struct {
	Name, Diagnostic, Bad, Good, Placement, Boundary string
	Accepted                                         bool
}

// Catalog is derived from refusals.go and the soundness/refusal tables in docs/0.1.md.
// Diagnostic is specific text, not the common "Adamic 0.1 refuses" prefix.
func Catalog() []Entry {
	return []Entry{
		{Name: "any", Diagnostic: "any", Bad: "let value: any = 1;", Good: "let value: number = 1;", Boundary: "Compiler work: explicit any is permanently Refused in .a; main returns NotYet. See RULINGS.md."},
		{Name: "cast", Diagnostic: "a cast the runtime can't check", Bad: "const value = 1; const cast = value as unknown as {readonly n: number};", Good: "const value = 1; const cast = {n: 1};"},
		{Name: "non-null", Diagnostic: "the non-null assertion !", Bad: "const value = new Map<string, number>().get('x')!;", Good: "const value = new Map<string, number>().get('x') ?? 0;"},
		{Name: "type-guard", Diagnostic: "a type predicate whose return is not proven (true return narrows to number, not number)", Bad: "function guard(x: number): x is number { return true; }", Good: "function guard(x: number): boolean { return typeof x === 'number'; }", Placement: "module"},
		{Name: "assertion-guard", Diagnostic: "a type predicate whose return is not proven (normal return has not narrowed x to number)", Bad: "function guard(x: number): asserts x is number {}", Good: "function guard(x: number): void {}", Placement: "module"},
		{Name: "definite-local", Diagnostic: "a definite assignment assertion !", Bad: "let value!: number; value = 1;", Good: "let value: number; value = 1;", Accepted: true},
		{Name: "definite-field", Diagnostic: "a definite assignment assertion !", Bad: "class Box { value!: number; constructor() { this.value = 1; } } const box = new Box();", Good: "class Box { value: number; constructor() { this.value = 1; } } const box = new Box();", Placement: "module", Accepted: true},
		{Name: "ts-ignore", Diagnostic: "@ts-ignore suppression directive", Bad: "// @ts-ignore\nconst value: number = 1;", Good: "const value: number = 1;"},
		{Name: "ts-expect-error", Diagnostic: "@ts-expect-error suppression directive", Bad: "// @ts-expect-error\nconst value: number = 'wrong';", Good: "const value: number = 1;"},
		{Name: "ts-nocheck", Diagnostic: "@ts-nocheck checking pragma", Bad: "// @ts-nocheck\nconst value: number = 1;", Good: "const value: number = 1;", Placement: "prefix"},
		{Name: "ts-check", Diagnostic: "@ts-check checking pragma", Bad: "// @ts-check\nconst value: number = 1;", Good: "const value: number = 1;", Placement: "prefix"},
		{Name: "async", Diagnostic: "an async function", Bad: "async function wait(): Promise<void> {}", Good: "function wait(): void {}", Placement: "module"},
		{Name: "generator", Diagnostic: "a generator function", Bad: "function* sequence() {}", Good: "function sequence(): void {}", Placement: "module"},
		{Name: "await", Diagnostic: "await", Bad: "export const value = await 1;", Good: "export const value = 1;", Placement: "module"},
		{Name: "yield", Diagnostic: "yield (generators)", Boundary: "yield requires a generator, which is refused first"},
		{Name: "decorator", Diagnostic: "a decorator", Bad: "function decorate<T>(value: T): void {} @decorate class Box { n = 1; } const box = new Box();", Good: "function decorate<T>(value: T): void {} class Box { n = 1; } const box = new Box();", Placement: "module"},
		{Name: "label", Boundary: "labeled break on loops and switches is supported on main (docs/0.1.md); labeled continue and other labeled statements are NotYet"},
		{Name: "with", Diagnostic: "with", Boundary: "loader rejects with in strict modules (TS1101)"},
		{Name: "delete", Diagnostic: "delete", Bad: "const box: {n?: number} = {n: 1}; delete box.n;", Good: "const box: {n?: number} = {n: 1};"},
		{Name: "debugger", Diagnostic: "debugger", Bad: "debugger;", Good: "const value = 1;"},
		// Numeric enums are open numbers on main. Member tags and runtime objects still
		// require proof; a supported enum declaration is never itself a refusal.
		{Name: "enum-tag", Diagnostic: "an object refinement using an open numeric enum as a literal tag", Bad: "enum AKind { A } enum BKind { B = 1 } interface A { readonly kind: AKind; readonly value: string } interface B { readonly kind: BKind; readonly value: number } function show(value: A | B): void { if (value.kind === BKind.B) { console.log(`${value.value + 1}`); } }", Good: "enum Kind { A, B } interface A { readonly kind: Kind.A; readonly value: string } interface B { readonly kind: Kind.B; readonly value: number } function show(value: A | B): void { if (value.kind === Kind.B) { console.log(`${value.value + 1}`); } }", Placement: "module"},
		{Name: "enum-object-view", Diagnostic: "reverse properties and the complete enum shape are unproven", Bad: "enum E { A, B } const copy: typeof E = { A: E.A, B: E.B };", Good: "enum E { A, B } const copy: typeof E = E; console.log(`${copy.A}`);", Placement: "module"},
		{Name: "enum-string-object-view", Diagnostic: "reverse properties and the complete enum shape are unproven", Bad: "enum E { A = 'a', B = 'b' } const copy = { A: E.A, B: E.B, hidden: 99 } as const; const view: typeof E = copy;", Good: "enum E { A = 'a', B = 'b' } const view: typeof E = E; console.log(view.A);", Placement: "module"},
		{Name: "enum-object-write", Diagnostic: "a write into an enum runtime object", Bad: "enum E { A, B } Object.assign(E, { A: E.A });", Good: "enum E { A, B } const copy = {A: E.A + 0, B: E.B + 0}; Object.assign(copy, {A: 3, B: 4});", Placement: "module"},
		{Name: "enum-prototype-name", Diagnostic: "an enum member name that changes the prototype or contains NUL", Bad: "enum E { '__proto__' = 'bad' }", Good: "enum E { Ordinary = 'good' } console.log(E.Ordinary);", Placement: "module"},
		{Name: "enum-nonfinite-name", Diagnostic: "a numeric enum member named like a non-finite number", Bad: "enum E { NaN = 0 }", Good: "enum E { Zero = 0 } console.log(`${E[0]}`);", Placement: "module"},
		{Name: "enum-nested", Diagnostic: "an enum inside a function or block; declare it at module scope", Bad: "function run(): void { enum E { A } }", Good: "enum E { A } function run(): void { console.log(`${E.A}`); } run();", Placement: "module", Boundary: "NotYet: enums inside functions or blocks need module scope on main"},
		{Name: "enum-merged", Diagnostic: "merged enum declarations; put the members in one declaration", Bad: "enum E { A } enum E { B = 1 }", Good: "enum E { A, B = 1 } console.log(`${E.B}`);", Placement: "module", Boundary: "NotYet: merge members into one runtime declaration"},
		{Name: "enum-ambient", Diagnostic: "an ambient enum without a runtime definition", Bad: "declare enum E { A = 0 }", Good: "enum E { A = 0 } console.log(`${E.A}`);", Placement: "module", Boundary: "NotYet: ambient enums have no runtime definition"},
		{Name: "namespace", Diagnostic: "this in a namespace function; a qualified call and a detached call have different receivers", Bad: "namespace N { export function read(this: {readonly x: number}): number { return this.x; } }", Good: "namespace N { export function read(state: {readonly x: number}): number { return state.x; } }", Placement: "module-first"},
		{Name: "void", Diagnostic: "the void operator", Bad: "void 1;", Good: "const value = 1;", Accepted: true},
		{Name: "index-signature", Diagnostic: "an index signature", Bad: "interface Indexed { [key: string]: number; }", Good: "interface Indexed { readonly n: number; }", Placement: "module"},
		{Name: "export-default", Diagnostic: "export default", Bad: "export default 1;", Good: "export const value = 1;", Placement: "module"},
		{Name: "equal", Diagnostic: "==", Bad: "const value = 1 == 1;", Good: "const value = 1 === 1;"},
		{Name: "unequal", Diagnostic: "!=", Bad: "const value = 1 != 1;", Good: "const value = 1 !== 1;"},
		{Name: "in", Diagnostic: "in", Bad: "const value = 'n' in {n: 1};", Good: "const value = {n: 1}.n === 1;", Accepted: true},
		{Name: "comma", Diagnostic: "the comma operator", Bad: "let n = 1; const value = (n += 1, n);", Good: "let n = 1; n += 1; const value = n;", Accepted: true},
		{Name: "and-assign", Diagnostic: "&&=", Bad: "let value: boolean = true; value &&= false;", Good: "let value: boolean = true; if (value) { value = false; }", Accepted: true},
		{Name: "or-assign", Diagnostic: "||=", Bad: "let value: boolean = false; value ||= true;", Good: "let value: boolean = false; if (!value) { value = true; }", Accepted: true},
		{Name: "arguments", Diagnostic: "arguments", Bad: "function count(): number { return arguments.length; } console.log(`${count()}`);", Good: "function count(): number { return 0; } console.log(`${count()}`);", Placement: "module", Accepted: true},
		{Name: "unbound-method", Diagnostic: "a method read as a value", Bad: "const items: number[] = []; const method = items.push;", Good: "const items: number[] = []; const method = (n: number): number => items.push(n);"},
		{Name: "mutable-variance", Diagnostic: "adamic/invariant-mutable", Bad: "const narrow: {n: number; s: string}[] = []; const wide: {n: number}[] = narrow;", Good: "const narrow: {n: number; s: string}[] = []; const wide: readonly {n: number}[] = narrow;"},
		{Name: "bivariant-method", Diagnostic: "method-signature-style", Bad: "interface A { n: number; } interface D extends A { s: string; } interface Narrow { handle(x: D): void; } interface Wide { handle(x: A): void; } function widen(x: Narrow): Wide { return x; }", Good: "interface A { n: number; } interface D extends A { s: string; } interface Narrow { handle(x: A): void; } interface Wide { handle(x: A): void; } function widen(x: Narrow): Wide { return x; }", Placement: "module"},
		{Name: "nominal-class", Diagnostic: "adamic/nominal-class", Bad: "class Box { n = 1; } const box: Box = {n: 1};", Good: "class Box { n = 1; } const box: Box = new Box();", Placement: "module"},
		{Name: "expando", Diagnostic: "properties added after creation", Bad: "function value() {} value.extra = 1;", Good: "const value = {call: (): void => {}, extra: 1};", Placement: "module", Boundary: "Compiler work: function expandos are permanently Refused; main returns NotYet. See RULINGS.md."},
		{Name: "spread", Diagnostic: "a spread after the first field", Bad: "const source = {n: 1}; const view: {} = source; const value = {n: 2, ...view};", Good: "const source = {n: 1}; const view: {} = source; const value = {...source, n: 2};"},
		{Name: "prototype-literal", Diagnostic: "__proto__ in an object literal", Bad: "const value = {__proto__: {n: 1}};", Good: "const value = {n: 1};"},
		{Name: "define-property", Diagnostic: "Object.defineProperty", Bad: "Object.defineProperty({n: 1}, 'n', {value: 'wrong'});", Good: "const value = {n: 1};"},
		{Name: "prototype-mutation", Diagnostic: "Object.setPrototypeOf", Bad: "Object.setPrototypeOf({n: 1}, {});", Good: "const value = {n: 1};"},
		{Name: "var", Diagnostic: "var", Bad: "var value = 1;", Good: "let value = 1;"},
		{Name: "truthiness", Diagnostic: "a number as a condition", Bad: "if (1) { console.log('ok'); }", Good: "if (1 === 1) { console.log('ok'); }", Accepted: true},
		{Name: "random", Diagnostic: "Math.random", Bad: "const value = Math.random();", Good: "const value = 0.5;"},
		{Name: "eval", Diagnostic: "eval", Bad: "eval('1');", Good: "const value = 1;", Boundary: "Compiler work: eval is permanently Refused; main returns NotYet. See RULINGS.md."},
		{Name: "function-type", Diagnostic: "Function", Bad: "function take(value: Function): void {}", Good: "function take(value: () => void): void {}", Placement: "module", Boundary: "Compiler work: Function annotations are Refused in .a, including unused parameters; main accepts. See RULINGS.md."},
		{Name: "new-function", Diagnostic: "Function", Bad: "const value = new Function('return 1');", Good: "const value = (): number => 1;", Boundary: "Compiler work: new Function is permanently Refused; main returns NotYet. See RULINGS.md."},
		{Name: "record", Bad: "const value: Record<string, number> = {};", Good: "const value = new Map<string, number>();", Accepted: true},
		{Name: "optional-widening", Diagnostic: "adamic/no-optional-widening", Bad: "const original = {x: 1, y: 'wrong'}; const view: {x: number} = original; const wider: {x: number; y?: number} = view;", Good: "const original = {x: 1, y: 'wrong'}; const view: {x: number} = original; const wider: {x: number; y?: number} = {x: view.x, y: 2};"},
		{Name: "merging", Diagnostic: "declaration merging", Bad: "class Box { n = 1; } interface Box { extra: number; } const box = new Box();", Good: "class Box { n = 1; } const box = new Box();", Placement: "module", Boundary: "Compiler work: uninitialized fields claimed by class/interface merging are Refused at the declaration in .a; main accepts. See RULINGS.md."},
		{Name: "constructor-escape", Diagnostic: "this escaping a constructor before every field is set", Bad: "class Box { n: number; constructor() { this.read(); this.n = 1; } read(): number { return this.n; } } const box = new Box();", Good: "class Box { n: number; constructor() { this.n = 1; this.read(); } read(): number { return this.n; } } const box = new Box();", Placement: "module"},
		{Name: "prototype-read", Diagnostic: "isPrototypeOf", Bad: "const value = {}.isPrototypeOf;", Good: "const value = (n: number): boolean => n === 1;"},
		// provePredicate refuses a type predicate whose body does not prove it. An arrow predicate is an
		// expression, so the writer can place it in every surrounding; the neighbor is the proven predicate.
		{Name: "unproven-predicate", Diagnostic: "a type predicate whose return is not proven (true return narrows to string | undefined, not string)", Bad: "const isText = (x: string | undefined): x is string => true; const word: string | undefined = 'a'; if (isText(word)) { console.log(word); }", Good: "const isText = (x: string | undefined): x is string => x !== undefined; const word: string | undefined = 'a'; if (isText(word)) { console.log(word); }"},
		{Name: "predicate-argument", Diagnostic: "an unproven predicate argument for parameter callback", Bad: "function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => true);", Good: "function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => typeof value === 'number');", Placement: "module"},
		{Name: "node-library", Diagnostic: "node:os.userInfo", Bad: "import {userInfo} from 'node:os'; userInfo();", Good: "console.log('user');", Placement: "module", Boundary: "NotYet: node:os.userInfo has no runtime implementation"},
		{Name: "typed-array-unsupported", Diagnostic: "typed array element type Int8Array", Bad: "const value = new Int8Array(4);", Good: "const value = new Uint8Array(4);", Boundary: "NotYet: Int8Array storage is not implemented"},
		{Name: "arguments-index", Diagnostic: "arguments other than a read of arguments.length", Bad: "function count(): void { console.log(`${arguments[0]}`); } count();", Good: "function count(): void { console.log(`${arguments.length}`); } count();", Placement: "module"},
		{Name: "node-buffer-read", Diagnostic: "Buffer.byteOffset outside the census value reads", Bad: "import {Buffer} from 'node:buffer'; console.log(`${Buffer.from('abc').byteOffset}`);", Good: "import {Buffer} from 'node:buffer'; console.log(`${Buffer.from('abc').length}`);", Placement: "module", Boundary: "NotYet: Buffer byteOffset value reads are outside the host census"},
		{Name: "library-method-object", Diagnostic: "a library method value outside a const alias (an object field erases its receiver and callable ABI); use an arrow", Bad: "const method = Number.parseInt; const object = {method};", Good: "const method = (text: string): number => Number.parseInt(text); const object = {method};", Boundary: "NotYet: opaque intrinsic aliases cannot escape into object fields"},
		{Name: "method-override", Diagnostic: "adamic/contravariant-override", Boundary: "checkOverrides also guards inheritance member kinds, accessor descriptors and ABI representation; not generated in this 0.1 corpus"},
		{Name: "parameter-properties", Diagnostic: "a parameter property", Bad: "class Box { constructor(public value: number) {} }", Good: "class Box { value: number; constructor(value: number) { this.value = value; } }", Placement: "module", Accepted: true},
		{Name: "overloads", Boundary: "implementation signatures are not lowered as overloads; needs a separate diagnostic design"},
		{Name: "reopened", Boundary: "throw Error/try/catch/finally, inheritance/super/abstract/protected, accessors, for-in on proven plain objects, RegExp/Set/JSON.stringify, Object helpers, arguments/files input opened after 0.1; not promised Refused by current compiler"},
		{Name: "library-boundaries", Boundary: "symbol/bigint/Date/process/globalThis/Intl and stdin/environment need individual loader/NotYet contracts; not generated"},
		{Name: "polymorphic-recursion", Boundary: "requires recursive generic instantiation; not generated"},
		{Name: "module-boundaries", Boundary: "namespace imports, export-star, default declarations and import cycles need multi-module probes; not generated"},
	}
}
