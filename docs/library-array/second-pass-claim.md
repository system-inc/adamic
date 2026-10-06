# Array second pass claim

This is the first task commit, before fixture or compiler edits. Base: `34da3d8`,
merged with origin/main `50045bd`. The independent validity runner is built from
origin/codex/test262-ts-validity `af12899` in a separate checkout. The exact
adapted inputs come from that runner, test262 `c8c798898646638cd0c24879f8e0374e847e7d74`,
adaptation on. Stock TypeScript 6.0.3 independently checks every checker refusal.

The per-input compiler and real tsc ledger leaves 437 refused and 1844
not-typescript, alongside the previous 191 passes and 610 skips. The completed runner baseline confirms those counts; its full table is
archived in second-pass-before.json.
The 495 figure belongs to pre-first-pass main, not this second pass's base.
First-blocker families below sum to all 437 remaining refusals. Later blockers
may be masked. No promised pass count is inferred from a family's size.

## Library claim, largest first

| Count | Feature | Owner |
|---:|---|---|
| 33 | Array constructors | array-library |
| 12 | Reductions | array-library |
| 7 | Other library constructors: Boolean | other-library |
| 7 | Other library constructors: String | other-library |
| 6 | Explicit receiver calls and observations: toString | array-library |
| 5 | Array constructor calls and aliases | array-library |
| 4 | Explicit receiver calls and observations: copyWithin | array-library |
| 4 | Explicit receiver calls and observations: findIndex | array-library |
| 4 | Explicit receiver calls and observations: findLastIndex | array-library |
| 4 | Explicit receiver calls and observations: toReversed | array-library |
| 4 | Explicit receiver calls and observations: toSorted | array-library |
| 4 | Other library constructors: Date | other-library |
| 4 | Other library constructors: Number | other-library |
| 3 | Default sort | array-library |
| 3 | Explicit receiver calls and observations: find | array-library |
| 3 | Explicit receiver calls and observations: findLast | array-library |
| 3 | Explicit receiver calls and observations: pop | array-library |
| 3 | Explicit receiver calls and observations: slice | array-library |
| 3 | Explicit receiver calls and observations: toSpliced | array-library |
| 3 | Other library constructors: EvalError | other-library |
| 2 | Explicit receiver calls and observations: concat | array-library |
| 2 | Explicit receiver calls and observations: every | array-library |
| 2 | Explicit receiver calls and observations: forEach | array-library |
| 2 | Explicit receiver calls and observations: join | array-library |
| 2 | Explicit receiver calls and observations: of | array-library |
| 2 | Explicit receiver calls and observations: push | array-library |
| 2 | Explicit receiver calls and observations: shift | array-library |
| 2 | Explicit receiver calls and observations: unshift | array-library |
| 1 | Array.from iterable overload | array-library |
| 1 | Array.prototype length observation | array-library |
| 1 | Explicit receiver calls and observations: entries | array-library |
| 1 | Explicit receiver calls and observations: fill | array-library |
| 1 | Explicit receiver calls and observations: filter | array-library |
| 1 | Explicit receiver calls and observations: flatMap | array-library |
| 1 | Explicit receiver calls and observations: includes | array-library |
| 1 | Explicit receiver calls and observations: isArray | array-library |
| 1 | Explicit receiver calls and observations: keys | array-library |
| 1 | Explicit receiver calls and observations: map | array-library |
| 1 | Explicit receiver calls and observations: reduce | array-library |
| 1 | Explicit receiver calls and observations: reduceRight | array-library |
| 1 | Explicit receiver calls and observations: reverse | array-library |
| 1 | Explicit receiver calls and observations: some | array-library |
| 1 | Explicit receiver calls and observations: splice | array-library |
| 1 | Explicit receiver calls and observations: values | array-library |
| 1 | Explicit receiver calls and observations: with | array-library |
| 1 | Object.freeze on arrays | other-library |
| 1 | Other library constructors: SyntaxError | other-library |

Build order: dense Array constructor overloads first, then explicit-receiver
calls and supported intrinsic observations, then reductions including
reduceRight and omitted initial accumulators, then remaining overloads that
existing representations can hold. Port V8's exact algorithms where available.

Array length construction with nonzero length needs sparse presence; boxed
Boolean/String/Number, Date and native Error constructors need their own
library representations. These are library features with language foundations,
not permission to simulate holes with undefined or erase boxed identity.
Other-library work belongs to its library owners. Unsupported Array overloads
will remain explicit refusals naming the representation they need.

## Language handoff to @system_adamic

Do not build these in this slice. Each row names the current first blocker and
has a one-line reproducer; the ledger supplies every complete adapted test.

| Count | Feature | One-line reproducer |
|---:|---|---|
| 49 | Property descriptors | `Object.defineProperty({x: 1}, "x", {value: 2});` |
| 46 | Empty or heterogeneous array slots | ``const empty: never[] = []; console.log(`${empty.length}`);`` |
| 37 | Truthiness | ``console.log(`${[1].some(() => 1)}`);`` |
| 26 | Property deletion | `const a = [1]; delete a[0];` |
| 20 | Prototype mutation | `Array.prototype[0] = 1;` |
| 18 | Arguments object | ``function f(): number { return arguments.length; } console.log(`${f()}`);`` |
| 18 | Sparse array literals | ``const a = [1, , 3]; console.log(`${a.length}`);`` |
| 14 | Unproven any, undefined, void or never values | ``const a = [1].map(() => {}); console.log(`${a.length}`);`` |
| 13 | Closure registration and hoisting | ``function f(): number { return a.length; } const a = [1]; console.log(`${f()}`);`` |
| 8 | Object coercion | ``console.log(`${Array.prototype.indexOf.call({0: 1, length: {valueOf: (): number => 1}}, 1)}`);`` |
| 8 | Var declarations | ``var n = 1; console.log(`${n}`);`` |
| 6 | Unrepresented intrinsic values | ``const f = Math; console.log(`${f.PI}`);`` |
| 4 | Getters | ``const o = {get length(): number {return 1;}}; console.log(`${o.length}`);`` |
| 3 | Statements with value expressions | `new Array(1);` |
| 3 | Structural views hiding primitive slots | ``const value = {}; console.log(`${Array.prototype.indexOf.call({0: value, length: 1}, value)}`);`` |
| 1 | Void operator | `void 1;` |
| 2 | Assignment through a temporary | `({x: 1}).x = 2;` |
| 2 | Mixed string and number operators | `console.log("x" + 1);` |
| 1 | Checker diagnostic mismatch | `function f() { return [].concat(arguments); } f();` |
| 1 | Generic array element slots | `function f<T>(a: T[]): T[] { return a.toSorted(); }` |
| 1 | In operator | ``console.log(`${"x" in {x: 1}}`);`` |
| 1 | Non-numeric comparator result | `[1].sort(() => "0");` |
| 1 | Top-level this | `const target = this;` |

The diagnostic-mismatch row is accepted by neither compiler: Adamic reports
TS2740 and stock tsc reports a different code. The requested runner retains it
in refused, so it stays visible as checker compatibility work, not an Array
implementation obligation. Prototype reads used as writes are language
mutation, rather than missing metadata. The pure prototype.length observation
is library work.

## Shrinking search fixture claim

Add the requested findLastIndex fixture and a findLast twin, with strings built
at runtime. Raw Node continues at missing indexes with undefined and finishes;
the existing backends insert a panic to protect the callback's proven element
type. Preserve that documented inserted check and record both Node's complete
output and the backend prefix/panic. Mutate each backend to skip the missing
index, show this fixture catches each, and restore. No change to
internal/oracle/library_array_mutant_test.go's environment line.

## Toolchain

bash cloud/setup.sh succeeded: Go ready 0s; clang ready 0s; Node ready 0s;
submodules ready 0s; build cache warm 84s; done 84s. nproc is 5;
cgroup cpu.max is 400000 100000; reported memory 17.6 GB.
Source /workspace/adamic-tools/env.sh for every build/test shell.
