# Array third pass claim

First task commit, before compiler or fixture edits. Branch codex/library-array-3
starts at origin/codex/library-array-2 bcf803d, as requested. The separate validity
runner checkout is origin/codex/test262-ts-validity af12899, stock tsc 6.0.3,
test262 c8c798898646638cd0c24879f8e0374e847e7d74, adaptation enabled.
The inherited final table is 242 passing, zero disagreements, 386 refused,
zero crashes, 610 skipped and 1844 not-typescript. A fresh complete baseline
measurement is running before implementation. Only refused is work owed.

The complete 386-input claim ledger is third-pass-claim.json. Source digests and
observed compiler reasons are preserved from the previous report. Classification
uses the actual program as well as its first refusal, because a method-read
refusal can hide prototype mutation, freezing or custom construction.

## Array library work, largest first

| Inputs | Feature | Implementation scope |
|---:|---|---|
| 4 | Proven intrinsic aliases and metadata | Track effect-free unreassigned bindings, prove every use is an intrinsic observation or supported explicit call, and eliminate the unobservable binding. Preserve typeof, name, length and absence of an own prototype. Refuse escaping or mutable aliases. |
| 1 | Generic map oversized ArraySpeciesCreate | Port V8's order: evaluate arguments, get length, validate callable, and throw RangeError for a receiver proven to require an invalid Array length. No simulated sparse array. |
| 1 | Never-returning sort comparator | Preserve abrupt completion from a comparator whose proven return type is never, for sort and toSorted. A void callback has no numeric result to read. |
| 1 | Fresh function receiver shift | Port V8's generic shift failure on the immutable length of a fresh zero-argument function. Preserve exact observed error text. The same corpus input later exercises descriptors, which remains language work. |

The 16 formerly Array-owned inputs contain these six library blockers, seven
language blockers, and three other-library blockers. Array.of with custom
constructors needs callable/newable identity and construction through a runtime
value; that is a language foundation, not permission to pretend every callable
is an Array constructor. Array function freezing belongs to Object.freeze.

## Language handoff to @system_adamic

Do not build these in this slice. Counts are classified current or masked first
blockers; later blockers may still be hidden. Each feature has a one-line
reproducer. The ledger supplies all complete adapted inputs.

| Feature | One-line reproducer |
|---|---|
| Property descriptors | `Object.defineProperty([1], "0", {value: 2});` |
| Empty or heterogeneous array slots | `const empty: never[] = []; console.log(empty.length);` |
| Truthiness | `[1].some(() => 1);` |
| Property deletion | `const a = [1]; delete a[0];` |
| Prototype mutation | `Array.prototype[0] = 1;` |
| Arguments object | `function f(): number { return arguments.length; }` |
| Sparse array literals | `const a = [1, , 3]; console.log(a.length);` |
| Unproven any, undefined, void or never values | `const a = [1].map(() => {}); console.log(a.length);` |
| Closure registration and hoisting | `function f(): number { return a.length; } const a = [1]; console.log(f());` |
| Object coercion | `Array.prototype.indexOf.call({0: 1, length: {valueOf: (): number => 1}}, 1);` |
| Var declarations | `var n = 1;` |
| Unrepresented intrinsic values | `const f = Math; console.log(f.PI);` |
| Getters | `const o = {get length(): number {return 1;}};` |
| Statements with value expressions | `new Array(1);` |
| Structural views hiding primitive slots | `const value = {}; Array.prototype.indexOf.call({0: value, length: 1}, value);` |
| Void operator | `void 1;` |
| Assignment through a temporary | `({x: 1}).x = 2;` |
| Mixed string and number operators | `console.log("x" + 1);` |
| Checker diagnostic mismatch | `function f() { return [].concat(arguments); } f();` |
| Generic array element slots | `function f<T>(a: T[]): T[] { return a.toSorted(); }` |
| In operator | `const present = "x" in {x: 1};` |
| Top-level this | `const target = this;` |
| Sparse array presence | `const a = new Array<number>(3);` |
| Indexed extension | `const a = new Array(1, 2); a[2] = 3;` |
| Constructor aliases in instanceof | `const C = Array; const a = new Array(1, 2); const b = a instanceof C;` |
| Custom callable constructor identity | `function C() {} const a = Array.of.call(C);` |

The diagnostic mismatch is retained in refused by the requested runner because
its tsc diagnostic differs from Adamic's diagnostic. It is checker compatibility,
not a library algorithm obligation. The JSON ledger retains its reason.

Current classified totals: 347 language, seven Array-library, 32 other-library.

Follow-up claim correction, before sort implementation: the inherited non-numeric
comparator item actually returns never by throwing. That is an Array callback
overload, not a language feature. Its exact accepted adapted source is in the
ledger. Indexed-extension reasons are also distinguished from sparse creation.
Other-library work: Boolean, String, Number, Date, EvalError and SyntaxError
constructors, Object.freeze on arrays/functions, and Object.prototype.toString.

## Evidence plan

Before/after complete built-ins/Array tables from the independent runner; every
new pass agrees with raw Node, and zero disagreements is required. New .a fixtures
run source on Node, native under sanitizers, and emitted JavaScript. A mutant for
each built method family must compile, exit cleanly with no sanitizer/leak report,
and be caught only by stdout comparison with Node. Small dispatch hooks only;
no edits to library_array_mutant_test.go's integration-owned environment line.
Linux package, uncached filtered Array oracle, counts and vet gates; exact commands,
outputs and uncovered work in the final report, with a five-line summary first.
