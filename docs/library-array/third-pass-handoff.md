# Array third pass language handoff

For @system_adamic: 349 remaining first blockers are language work. This lists
every family with one-line reproducer; full accepted adapted sources, SHA256s,
observed reasons and classifications are in third-pass-final-refusals.json.
No maintainer message was sent.

| Inputs | Language feature | One-line reproducer |
|---:|---|---|
| 62 | Property descriptors | `Object.defineProperty([1], "0", {value: 2});` |
| 56 | Empty or heterogeneous array slots | `const empty: never[] = []; console.log(empty.length);` |
| 37 | Truthiness | `[1].some(() => 1);` |
| 26 | Property deletion | `const a = [1]; delete a[0];` |
| 24 | Prototype mutation | `Array.prototype[0] = 1;` |
| 18 | Sparse array literals | `const a = [1, , 3]; console.log(a.length);` |
| 18 | Arguments object | `function f(): number { return arguments.length; }` |
| 14 | Unproven any, undefined, void or never values | `const a = [1].map(() => {}); console.log(a.length);` |
| 13 | Closure registration and hoisting | `function f(): number { return a.length; } const a = [1]; console.log(f());` |
| 12 | Sparse array presence | `const a = new Array<number>(3);` |
| 10 | Var declarations | `var n = 1;` |
| 10 | Indexed extension | `const a = new Array(1, 2); a[2] = 3;` |
| 9 | Object coercion | `Array.prototype.indexOf.call({0: 1, length: {valueOf: (): number => 1}}, 1);` |
| 8 | Generic array element slots | `function f<T>(a: T[]): T[] { return a.toSorted(); }` |
| 6 | Unrepresented intrinsic values | `const f = Math; console.log(f.PI);` |
| 4 | Getters | `const o = {get length(): number {return 1;}};` |
| 3 | Statements with value expressions | `new Array(1);` |
| 3 | Structural views hiding primitive slots | `const value = {}; Array.prototype.indexOf.call({0: value, length: 1}, value);` |
| 3 | Mixed string and value operators | `console.log("x" + undefined);` |
| 3 | In operator | `const present = "x" in {x: 1};` |
| 2 | Assignment through a temporary | `({x: 1}).x = 2;` |
| 2 | Custom callable constructor identity | `function C() {} const a = Array.of.call(C);` |
| 2 | Void operator | `void 1;` |
| 2 | Constructor aliases in instanceof | `const C = Array; const a = new Array(1, 2); const b = a instanceof C;` |
| 1 | Top-level this | `const target = this;` |
| 1 | Checker diagnostic mismatch | `function f() { return [].concat(arguments); } f();` |

The checker diagnostic mismatch is classified as refused by the requested runner
because tsc rejects it with a different diagnostic. It is checker compatibility,
not an Array algorithm. Reproducers describe the foundation, not necessarily the
whole corpus input. Mixed string/value concatenation and descriptors become
visible after the new join metadata and function-shift paths lower.

Other-library work remains 32 inputs:

| Inputs | Feature |
|---:|---|
| 7 | Other library constructors: Boolean |
| 7 | Other library constructors: String |
| 4 | Other library constructors: Number |
| 4 | Other library constructors: Date |
| 3 | Other library constructors: EvalError |
| 3 | Object.freeze on arrays |
| 2 | Object.freeze on intrinsic functions |
| 1 | Other library constructors: SyntaxError |
| 1 | Object.prototype.toString tag |

All seven claimed Array first blockers were implemented. Five entire inputs now
compile; two expose the language blockers above. These counts concern the first
blocker in this pinned corpus, and do not claim general intrinsic escape, sparse
generic receivers or optional reference sorting are supported.
