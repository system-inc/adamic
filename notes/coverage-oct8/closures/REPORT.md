Built 19 .a programs, a four-way byte comparator, a leak check, and eight isolated Go-overlay mutants.
Commits: base 031a1259bc7973934792dc6cb1bd4074fc2204b9; programs/evidence c91a13818dea996a58ab1e307fdb7381eb9592f7; report commit is the next commit on codex/coverage-oct8-closures.
Commands/results: run.py passed its expectations; 15 four-way agreements (14 clean exits, one expected exit 70), four compiler boundaries; lower, native, filtered oracle and vet passed after dependency repair.
Mutants: bind-presence caught by lower; count-zero by bytes; packed-leak by LSAN; spread-snapshot and rest-retain by ASan and existing oracle; mixed-tuple survived lower then failed C; delayed-fill masked; unnamed-nested survived lower.
Not covered: full repository gate, exhaustive exit/type/guard coverage, synthetic unnamed ASTs, or executable native behavior for the four refused/type-invalid programs.

# Scope and method

Read CLAUDE.md first, then README.md, docs/0.1.md and docs/memory.md before editing. Read the five named implementation files whole and the requested expression.go/class.go/object.go step 04 hunks against current code. Reviewed `git log -p 73352e87..5a2681b1 -- <named files>` and relevant non-merge patches. Relevant method/nested/count history includes 57bf6dc1, b15216da, 990eb119, 41995767, 433b2abe, f76331d5, 7e635882, 89c8e81e, a02dd049 and d2c2dc2b; namespace admission/revert/restore includes adc45ca4, 46102ff5 and fe8ec9a9. These observations apply to the base SHA, not to every intermediate commit.

No compiler source was changed. Mutations use Go overlays; their exact patches and commands are in mutations/. Programs remain outside the registered oracle fixture set. This unit adds reproducible probes and an account, not production fixes or a main-branch merge.

`python3 notes/coverage-oct8/closures/run.py` runs source through Node's own type stripping via oracle/node.mjs, builds native with --sanitize (the oracle's ASan/UBSan configuration), builds default release at -O2, and compiles the JavaScript backend then runs it through the same Node wrapper. Comparison uses raw stdout/stderr bytes and exit codes; results/observations.json contains verbatim text and hex, including every compilation stage and command. Node here means the repository's normalized oracle wrapper, not bare Node stack traces: the uncaught throw is normalized to exit 70. Native sanitizer execution disables leak detection for oracle parity; a separate LSAN execution enables it for successful exits. Clean means agreement, exit zero, and leak execution exit zero. Refused compilation is recorded as a compiler result, never as a native execution.

# Setup and verification

`bash cloud/setup.sh > /tmp/coverage-setup.log 2>&1` succeeded; sourced /workspace/adamic-tools/env.sh. nproc=5, cpu.max=400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8. Timing lines are preserved verbatim in evidence/coverage-setup.log: tool readiness 0.071s/0.085s/0.367s; markdown installation 2.019s, ready 2.185s; submodules 23.087s; build 226.060s; deferred tests 226.152s; cache warm 226.153s; total 226.177s.

Initial `go test -count=1 -timeout 30m ./internal/lower ./internal/native` failed lower because @types/node 25.3.3 was absent; native passed in 279.971s. The full failure is in evidence/coverage-packages-baseline.log. `npm ci --prefix stage3/api` installed three packages; then `go test -count=1 -timeout 30m ./internal/lower` passed in 36.133s. Dependency installation left tracked package files unchanged.

`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(arguments_length|nested_|library_method_values|method_coverage_bound|closures_throw|class_super_closure)|TestArgumentsLengthWrongSlotMutant'` passed in 17.258s. `go vet ./internal/lower ./internal/native` passed with empty output. All test output was redirected to logs, never piped. Full gate, cohere lint and counts gate were not run; no claim is made about them. Exact package commands for mutants are in each record.json. Native mutant package checks are specifically TestNoReaderCallingConvention, not the whole native package.

# Observations on current code

The coverage matrix below describes the additional programs. The four-way result and leak result are measured observations; extrapolating them to all closure lifetimes would be an inference.

| Program | Case |
| --- | --- |
| parameter_lifetime | Reassigned captured string parameter, independent activations, returned closure |
| loop_updates | Per-iteration let capture, reassignment, continue and break, shared outer loop variable |
| nested_recursive_escape | Escaping mutually recursive captured functions, mutable prefix and counter |
| captured_recursive_function | Captured callable parameter reassigned to another recursive function |
| escape_creator_throw | Closure saved before creator throws, used after catch, then released |
| exits_finally | Normal and early return, throws, continue/break, finally mutation and overriding exits |
| rest_capture_lifetime | Captured default parameter, rest array and creator count survive repeated calls |
| nested_argument_counts | Nested defaults/rest/arguments.length, explicit undefined, spread snapshot under later mutation |
| count_default_evaluation | arguments.length inside default, function values, absent/undefined/rest/empty spread |
| namespace_nested_counts | Qualified factory and alias, namespace variable updates, outer and inner counts |
| bound_receiver_escape | Intrinsic method value binds runtime string, survives source reassignment and reattachment |
| bound_creator_throw | Retained bound method survives creator throw and is released |
| arrow_this_reattach | Lexical this survives escape and reattachment; original receiver mutation remains visible |
| detached_method_reattach | Ordinary detached class method reattached to another object, intentionally refused |
| delayed_fill_guard | Delayed Array.prototype.fill on an array with holes, intentionally refused |
| mixed_tuple_spread_guard | Mixed number/string tuple spread, intentionally refused |
| number_extra_arguments_guard | Extra toFixed apply arguments, rejected by the type checker before lower guard |
| optional_intrinsic_comment | Optional zero-argument trim contradicts the blanket optionalCall comment |
| uncaught_nested_throw | Escaping closure throw and finally, normalized expected exit 70 |

All 14 successful accepted programs agreed byte for byte and finished cleanly with no LSAN finding. The uncaught throw also agreed byte for byte but is an intentional stop, not a clean success or a leak-tested exit. There was no observed silent miscompile, current C compilation failure, sanitizer failure or leak among the accepted probes.

Four current boundary disagreements are compiler errors: delayed fill at internal/lower/library_method_values.go:251, mixed tuple representation at internal/lower/arguments_length.go:81, detached class method at internal/lower/refusals.go:205 (also guarded in object.go:388), and extra toFixed arguments at the checker, before library_method_values.go's arity handling. The last is deliberately a type-invalid boundary witness rather than an admitted Adamic program. These results do not demonstrate a regression; Node's accepted syntax is broader than Adamic's admitted subset.

# Comment and documentation audit

Observation: internal/lower/expression.go:1275-1281 says library optional calls are not lowered and gives only the declared-object method form. The current call dispatcher at :643 handles optionalIntrinsic first; optional_intrinsic_comment.a executes nullable trim successfully on all four paths. Kind: comment does not match current dispatch. Inference: update the comment's stated scope to calls that reach optionalCall.

Observation: internal/lower/object.go:659 says builtin handles Math/toFixed and isBuiltin is false for every other call. Its current dispatch also handles libraryEmptyMap/libraryMethodCall and other library forms. The intrinsic-bound probes execute such methods. Kind: comment does not match the code. Inference: broaden that description.

Observation: library_method_values.go:138 names `sequence` while the function is methodSequence. Its behavioral claim about preserving preceding evaluation/checks matches the implementation; this is a stale name, not a behavioral bug. object.go:1180's push/join description is incomplete as an inventory, without claiming those are the only array methods.

The following inventory groups every comment/doc block in the five whole-read files by starting line; static comparison found no additional contradiction. Static agreement is separate from runtime coverage.

| File | Comment/doc block starts | Checked against |
| --- | --- | --- |
| lower/library_method_values.go | 9,111,138,175,179,411,451,538 | Intrinsic selection, sequencing, adapter arguments, binding/presence, callback forwarding |
| lower/nested_functions.go | 10,22,32,59,85,105,168,193,201,219,247,292 | Predeclaration, cells, recursive capture graph, capture propagation and callable forwarding |
| lower/arguments_length.go | 65,119,157,169,206,223,229,246 | Spread representation, source count, context restoration, computed thunk captures |
| native/arguments_length.go | 10,48,69,89,168,184 | Actual count versus padded slots, ordered pack/snapshot, retained rest values, call reuse |
| lower/namespaces.go | 20,39,179,243,277,297,315,342,386,419,435,453,507 | Qualified flattening, declaration order, bindings, exported access, callable merges |

The qualified-only namespace description at :39 applies to runtime namespace containers; merged callable identities have explicit handling, so it is not evidence that every merged function identity is invisible. In the step 04 hunks, stable function adapters and actual-count forwarding in expression.go, class super/this capture comments in class.go, and hasOwn/number/string call ordering comments in object.go match the inspected paths. Only the two blanket dispatch descriptions above assert behavior contradicted by current dispatch. No comments were edited.

# Mutation observations and limits

| Mutant and location | Package test observation | Additional observation | Interpretation |
| --- | --- | --- | --- |
| bind-presence, lower/library_method_values.go:467 | Full lower fails TestLibraryMethodValueSafety/bind_absent_receiver; unsafe method value admitted | Positive bound receiver program still agrees | Existing admission check catches removed presence guard |
| mixed-tuple, lower/arguments_length.go:81 | Full lower passes, 44.963s | Mixed tuple probe passes lowering; generated native C fails compiling double as retain pointer; JS succeeds | Branch removal is not noticed by lower package; clang failure is not proof of a successful semantic mutant |
| delayed-fill, lower/library_method_values.go:251 | Full lower passes, 32.099s | Probe still refused as array of any instead of delayed fill | Another refusal masks removal; pinned diagnostic check fails, no unsafe admission demonstrated |
| unnamed-nested, lower/nested_functions.go:28 | Full lower passes, 30.097s | No public valid source witness; checker/parser requires declaration name | Defensive branch unobserved by package tests; reachability through synthetic AST remains untested |
| spread-snapshot, native/arguments_length.go:97 | Focused native test passes | nested_argument_counts ASan use-after-free; O2 wrong first spread element; existing arguments_length_spread oracle fails | Snapshot ownership/evaluation is necessary on exercised path |
| rest-retain, native/arguments_length.go:128 | Focused native test passes | nested_argument_counts ASan use-after-free and O2 SIGSEGV; existing spread oracle fails; rest_capture_lifetime still agrees | New combined spread/rest witness catches missing retain; simpler rest capture is a negative control |
| count-zero, native/arguments_length.go:27 | Focused native test passes | All paths run successfully but native reports 0:0:4\|5 instead of 3:3:4\|5 | Byte comparison independently proves able to fail, without relying on a crash |
| packed-leak, native/arguments_length.go:93 | Focused native test passes | All four outputs agree; separate LSAN exits 23, 176 bytes in four allocations | Leak check independently proves able to fail despite perfect oracle output |

`python3 notes/coverage-oct8/closures/mutants.py <name>` reproduces each isolated mutation. Exact patches, package logs, build logs, commands and verbatim probe observations are committed under mutations/. Snapshot/rest existing-oracle failure logs are under evidence/. Mixed-tuple's probe was rerun after final runner expectations and exited 1; package evidence was not repeated because source mutation was unchanged. The three lower guard removals that leave lower green are evidence of package coverage gaps, not proof they could silently miscompile an accepted program. No synthetic AST test was attempted for the unnamed guard.

Observations establish representative return, throw, catch, finally, continue and break lifetime behavior for strings, arrays, receiver objects and callable captures. Inference is limited: this does not prove retain/release on every possible exit path, nor generic/virtual callbacks, constructors, every spread tuple shape, every namespace merge or every guarded branch. A complete guard census was not performed. The proposed suspected gaps were exercised or explicitly marked masked/unreachable above.

# Verbatim four-way records

Each program is included below. Empty fenced stdout/stderr blocks mean zero bytes. A failed compile is shown verbatim as that path's result; no runtime result exists on that path. Full stage records and hexadecimal bytes remain in the adjacent JSON. Mutant records are distinct from current compiler findings.

## arrow_this_reattach

```typescript
class Counter {
    value: string;
    constructor(seed: number) { this.value = `counter${seed}`; }
    make(): () => string { return () => this.value; }
}
function create(seed: number): () => string { return new Counter(seed).make(); }
const first = create(1);
const owner = new Counter(2);
const second = owner.make();
owner.value = `changed${3}`;
const other = { value: 'other', callback: second };
console.log(`${first()}|${second()}|${other.callback()}`);
```

Source Node: exit 0

stdout:
```text
counter1|changed3|changed3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
counter1|changed3|changed3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
counter1|changed3|changed3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
counter1|changed3|changed3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
counter1|changed3|changed3
```

stderr:
```text
```

## bound_creator_throw

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- bind supplies a primitive receiver. */
let saved: (() => string) | undefined = undefined;
function create(seed: number): void {
    const trim = String.prototype.trim;
    saved = trim.bind(`  kept${seed}  `);
    throw new Error(`creator${seed}`);
}
try { create(3); }
catch (error) { console.log(error instanceof Error ? error.message : 'unexpected'); }
function invokeSaved(): void {
    if (saved !== undefined) { console.log(saved()); }
}
invokeSaved();
saved = undefined;
```

Source Node: exit 0

stdout:
```text
creator3
kept3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
creator3
kept3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
creator3
kept3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
creator3
kept3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
creator3
kept3
```

stderr:
```text
```

## bound_receiver_escape

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- Intrinsic receivers are supplied explicitly. */
const trim = String.prototype.trim;
function make(seed: number): () => string {
    let text = `  bound${seed}  `;
    const bound = trim.bind(text);
    text = `  replaced${seed}  `;
    console.log(trim.call(text));
    return bound;
}
const first = make(1);
const second = make(2);
const holder = { callback: first };
console.log(`${first()}|${holder.callback()}|${second()}`);
```

Source Node: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

## captured_recursive_function

```typescript
function sum(n: number): number { return n === 0 ? 0 : n + sum(n - 1); }
function product(n: number): number { return n === 0 ? 1 : n * product(n - 1); }
function make(compute: (n: number) => number): () => string {
    function read(): string { return `computed:${compute(5)}`; }
    compute = product;
    return read;
}
const held = make(sum);
console.log(held());
console.log(held());
```

Source Node: exit 0

stdout:
```text
computed:120
computed:120
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
computed:120
computed:120
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
computed:120
computed:120
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
computed:120
computed:120
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
computed:120
computed:120
```

stderr:
```text
```

## count_default_evaluation

```typescript
function measure(first: number = arguments.length, ...rest: number[]): string {
    return `${arguments.length}:${first}:${rest.join('|')}`;
}
const run: (first?: number, ...rest: number[]) => string = measure;
console.log(measure());
console.log(measure(undefined));
console.log(measure(undefined, 2, 3));
console.log(run());
console.log(run(undefined, 4, 5));
const empty: number[] = [];
console.log(run(...empty));
```

Source Node: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

## delayed_fill_guard

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- call supplies the receiver. */
const fill = Array.prototype.fill;
const result: number[] = fill.call(new Array<number>(2), 7);
console.log(result.join('|'));
```

Source Node: exit 0

stdout:
```text
7|7
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower delayed fill of an array with holes; write the direct filled construction yet
```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower delayed fill of an array with holes; write the direct filled construction yet
```

JavaScript backend Node: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower delayed fill of an array with holes; write the direct filled construction yet
```

## detached_method_reattach

```typescript
class Counter {
    value: string;
    constructor(seed: number) { this.value = `counter${seed}`; }
    read(): string { return this.value; }
}
const owner = new Counter(1);
const detached = owner.read;
const other = { value: `other${2}`, read: detached };
console.log(other.read());
```

Source Node: exit 0

stdout:
```text
other2
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/detached_method_reattach.a:7:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => owner.read() (unbound-method)
```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/detached_method_reattach.a:7:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => owner.read() (unbound-method)
```

JavaScript backend Node: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/detached_method_reattach.a:7:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => owner.read() (unbound-method)
```

## escape_creator_throw

```typescript
let saved: (() => string) | undefined = undefined;
function creator(text: string): void {
    function read(): string { return text; }
    saved = read;
    text = `${text}:after`;
    throw new Error(`leave:${text}`);
}
try { creator(`local${1}`); }
catch (error) { console.log(error instanceof Error ? error.message : 'unexpected'); }
function invokeSaved(): void {
    if (saved !== undefined) { console.log(saved()); }
}
invokeSaved();
saved = undefined;
console.log('released');
```

Source Node: exit 0

stdout:
```text
leave:local1:after
local1:after
released
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
leave:local1:after
local1:after
released
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
leave:local1:after
local1:after
released
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
leave:local1:after
local1:after
released
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
leave:local1:after
local1:after
released
```

stderr:
```text
```

## exits_finally

```typescript
function trial(mode: number): string {
    let text = `local${mode}`;
    function read(): string { return text; }
    try {
        if (mode === 0) { return read(); }
        if (mode === 1) { throw new Error(read()); }
        for (let index = 0; index < 3; index += 1) {
            const temporary = `${read()}:${index}`;
            const capture = () => temporary;
            if (index === 0) { console.log(capture()); continue; }
            console.log(capture());
            break;
        }
        return read();
    } finally {
        text = `${text}:finally`;
        console.log(read());
        if (mode === 3) { return read(); }
        if (mode === 4) { throw new Error(read()); }
    }
}
for (let mode = 0; mode < 5; mode += 1) {
    try { console.log(`result:${trial(mode)}`); }
    catch (error) { console.log(error instanceof Error ? `caught:${error.message}` : 'unexpected'); }
}
```

Source Node: exit 0

stdout:
```text
local0:finally
result:local0
local1:finally
caught:local1
local2:0
local2:1
local2:finally
result:local2
local3:0
local3:1
local3:finally
result:local3:finally
local4:0
local4:1
local4:finally
caught:local4:finally
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
local0:finally
result:local0
local1:finally
caught:local1
local2:0
local2:1
local2:finally
result:local2
local3:0
local3:1
local3:finally
result:local3:finally
local4:0
local4:1
local4:finally
caught:local4:finally
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
local0:finally
result:local0
local1:finally
caught:local1
local2:0
local2:1
local2:finally
result:local2
local3:0
local3:1
local3:finally
result:local3:finally
local4:0
local4:1
local4:finally
caught:local4:finally
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
local0:finally
result:local0
local1:finally
caught:local1
local2:0
local2:1
local2:finally
result:local2
local3:0
local3:1
local3:finally
result:local3:finally
local4:0
local4:1
local4:finally
caught:local4:finally
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
local0:finally
result:local0
local1:finally
caught:local1
local2:0
local2:1
local2:finally
result:local2
local3:0
local3:1
local3:finally
result:local3:finally
local4:0
local4:1
local4:finally
caught:local4:finally
```

stderr:
```text
```

## loop_updates

```typescript
const saved: (() => string)[] = [];
for (let index = 0; index < 4; index += 1) {
    let label = `item${index}`;
    saved.push(() => `${index}:${label}`);
    label = `${label}!`;
    if (index === 1) { continue; }
    if (index === 2) { break; }
}
console.log(saved.map((read) => read()).join('|'));
const outer: (() => number)[] = [];
let shared = 0;
for (; shared < 3; shared += 1) { outer.push(() => shared); }
console.log(outer.map((read) => read()).join('|'));
```

Source Node: exit 0

stdout:
```text
0:item0!|1:item1!|2:item2!
3|3|3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:item0!|1:item1!|2:item2!
3|3|3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:item0!|1:item1!|2:item2!
3|3|3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:item0!|1:item1!|2:item2!
3|3|3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:item0!|1:item1!|2:item2!
3|3|3
```

stderr:
```text
```

## mixed_tuple_spread_guard

```typescript
function read(number: number, text: string): string {
    return `${arguments.length}:${number}:${text}`;
}
console.log(read(...[7, `word${8}`]));
```

Source Node: exit 0

stdout:
```text
2:7:word8
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/mixed_tuple_spread_guard.a:4:18: stage 0 can't lower a call spreading a tuple with differently represented elements yet
```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/mixed_tuple_spread_guard.a:4:18: stage 0 can't lower a call spreading a tuple with differently represented elements yet
```

JavaScript backend Node: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/mixed_tuple_spread_guard.a:4:18: stage 0 can't lower a call spreading a tuple with differently represented elements yet
```

## namespace_nested_counts

```typescript
namespace Store {
    let label = 'initial';
    export function set(value: string): void { label = value; }
    export function make(base: number = 10): (...values: number[]) => string {
        const outerCount = arguments.length;
        function read(...values: number[]): string {
            return `${label}:${base}:${outerCount}:${arguments.length}:${values.join('|')}`;
        }
        base += 1;
        return read;
    }
}
const factory = Store.make;
const first = factory();
const second = Store.make(20);
Store.set(`changed${3}`);
console.log(first());
console.log(first(...[1, 2]));
console.log(second(3, ...[4, 5]));
```

Source Node: exit 0

stdout:
```text
changed3:11:0:0:
changed3:11:0:2:1|2
changed3:21:1:3:3|4|5
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
changed3:11:0:0:
changed3:11:0:2:1|2
changed3:21:1:3:3|4|5
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
changed3:11:0:0:
changed3:11:0:2:1|2
changed3:21:1:3:3|4|5
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
changed3:11:0:0:
changed3:11:0:2:1|2
changed3:21:1:3:3|4|5
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
changed3:11:0:0:
changed3:11:0:2:1|2
changed3:21:1:3:3|4|5
```

stderr:
```text
```

## nested_argument_counts

```typescript
function make(label: string): (head?: string, ...tail: string[]) => string {
    function read(head: string = `default:${label}`, ...tail: string[]): string {
        const count = arguments.length;
        const captured = () => `${count}:${head}:${tail.join('|')}:${label}`;
        return captured();
    }
    label = `${label}!`;
    return read;
}
const run = make(`label${1}`);
console.log(run());
console.log(run(undefined));
console.log(run(`head${2}`, `tail${3}`, `tail${4}`));
const items = [`x${5}`, `y${6}`];
function replace(): string { items[0] = `z${7}`; items.push(`q${8}`); return `last${9}`; }
console.log(run(...items, replace()));
console.log(run(...items));
```

Source Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

## nested_recursive_escape

```typescript
function make(prefix: string): (n: number) => string {
    let calls = 0;
    function left(n: number): string {
        calls += 1;
        return n === 0 ? `${prefix}:${calls}` : right(n - 1);
    }
    function right(n: number): string {
        calls += 1;
        const forward = (value: number): string => left(value);
        return n === 0 ? `${prefix}:${calls}` : forward(n - 1);
    }
    prefix = `${prefix}!`;
    return right;
}
const first = make(`seed${7}`);
console.log(first(4));
console.log(first(3));
console.log(make(`seed${8}`)(2));
```

Source Node: exit 0

stdout:
```text
seed7!:5
seed7!:9
seed8!:3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
seed7!:5
seed7!:9
seed8!:3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
seed7!:5
seed7!:9
seed8!:3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
seed7!:5
seed7!:9
seed8!:3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
seed7!:5
seed7!:9
seed8!:3
```

stderr:
```text
```

## number_extra_arguments_guard

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- call supplies a number. */
const format = Number.prototype.toFixed;
function extra(): number { console.log('extra evaluated'); return 99; }
console.log(format.apply(1.25, [1, extra()]));
```

Source Node: exit 0

stdout:
```text
extra evaluated
1.3
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
notes/coverage-oct8/closures/number_extra_arguments_guard.a:4:32: error TS2345: Argument of type '[number, number]' is not assignable to parameter of type '[fractionDigits?: number | undefined]'.
  Source has 2 element(s) but target allows only 1.
```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
notes/coverage-oct8/closures/number_extra_arguments_guard.a:4:32: error TS2345: Argument of type '[number, number]' is not assignable to parameter of type '[fractionDigits?: number | undefined]'.
  Source has 2 element(s) but target allows only 1.
```

JavaScript backend Node: exit 1

stdout:
```text
```

stderr:
```text
notes/coverage-oct8/closures/number_extra_arguments_guard.a:4:32: error TS2345: Argument of type '[number, number]' is not assignable to parameter of type '[fractionDigits?: number | undefined]'.
  Source has 2 element(s) but target allows only 1.
```

## optional_intrinsic_comment

```typescript
function read(text?: string): string { return text?.trim() ?? 'missing'; }
console.log(read());
console.log(read(`  present${2}  `));
```

Source Node: exit 0

stdout:
```text
missing
present2
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
missing
present2
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
missing
present2
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
missing
present2
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
missing
present2
```

stderr:
```text
```

## parameter_lifetime

```typescript
function make(text: string): () => string {
    function read(): string { return text; }
    text = `${text}:changed`;
    return read;
}
function outer(seed: number): () => string {
    const built = `owned${seed}`;
    return make(built);
}
const first = outer(1);
const second = outer(2);
console.log(`${first()}|${second()}|${first()}`);
```

Source Node: exit 0

stdout:
```text
owned1:changed|owned2:changed|owned1:changed
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
owned1:changed|owned2:changed|owned1:changed
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
owned1:changed|owned2:changed|owned1:changed
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
owned1:changed|owned2:changed|owned1:changed
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
owned1:changed|owned2:changed|owned1:changed
```

stderr:
```text
```

## rest_capture_lifetime

```typescript
function make(head: string = `default${1}`, ...tail: string[]): () => string {
    const count = arguments.length;
    function read(): string {
        tail.push(`${head}:${tail.length}`);
        return `${count}:${head}:${tail.join('|')}`;
    }
    head = `${head}!`;
    return read;
}
const first = make();
const second = make(undefined, `tail${2}`, `tail${3}`);
console.log(first());
console.log(second());
console.log(first());
console.log(second());
```

Source Node: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

## uncaught_nested_throw

```typescript
function make(text: string): () => string {
    function read(): string { throw new Error(`${text}:throw`); }
    text = `${text}!`;
    return read;
}
const fail = make(`local${4}`);
try { console.log(fail()); }
finally { console.log('finally ran'); }
```

Source Node: exit 70

stdout:
```text
finally ran
```

stderr:
```text
adamic: panic: Error: local4!:throw
```

Native sanitizer: exit 70

stdout:
```text
finally ran
```

stderr:
```text
adamic: panic: Error: local4!:throw
```

Native -O2: exit 70

stdout:
```text
finally ran
```

stderr:
```text
adamic: panic: Error: local4!:throw
```

JavaScript backend Node: exit 70

stdout:
```text
finally ran
```

stderr:
```text
adamic: panic: Error: local4!:throw
```

# Mutant bind-presence

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## bound_receiver_escape

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- Intrinsic receivers are supplied explicitly. */
const trim = String.prototype.trim;
function make(seed: number): () => string {
    let text = `  bound${seed}  `;
    const bound = trim.bind(text);
    text = `  replaced${seed}  `;
    console.log(trim.call(text));
    return bound;
}
const first = make(1);
const second = make(2);
const holder = { callback: first };
console.log(`${first()}|${holder.callback()}|${second()}`);
```

Source Node: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
replaced1
replaced2
bound1|bound1|bound2
```

stderr:
```text
```

# Mutant count-zero

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## count_default_evaluation

```typescript
function measure(first: number = arguments.length, ...rest: number[]): string {
    return `${arguments.length}:${first}:${rest.join('|')}`;
}
const run: (first?: number, ...rest: number[]) => string = measure;
console.log(measure());
console.log(measure(undefined));
console.log(measure(undefined, 2, 3));
console.log(run());
console.log(run(undefined, 4, 5));
const empty: number[] = [];
console.log(run(...empty));
```

Source Node: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
0:0:4|5
0:0:
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
0:0:4|5
0:0:
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
3:3:4|5
0:0:
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:0:
1:1:
3:3:2|3
0:0:
0:0:4|5
0:0:
```

stderr:
```text
```

# Mutant delayed-fill

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## delayed_fill_guard

```typescript
/* eslint-disable @typescript-eslint/unbound-method -- call supplies the receiver. */
const fill = Array.prototype.fill;
const result: number[] = fill.call(new Array<number>(2), 7);
console.log(result.join('|'));
```

Source Node: exit 0

stdout:
```text
7|7
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower an array of any yet
```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower an array of any yet
```

JavaScript backend Node: exit 1

stdout:
```text
```

stderr:
```text
adamic: /workspace/adamic/notes/coverage-oct8/closures/delayed_fill_guard.a:3:26: stage 0 can't lower an array of any yet
```

# Mutant mixed-tuple

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## mixed_tuple_spread_guard

```typescript
function read(number: number, text: string): string {
    return `${arguments.length}:${number}:${text}`;
}
console.log(read(...[7, `word${8}`]));
```

Source Node: exit 0

stdout:
```text
2:7:word8
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-1360962882/main.c:33:82: error: passing 'double' to parameter of incompatible type 'void *'
   33 |         adamic_array_push(adamic_temporary_8, (adamic_value){.reference = adamic_retain((0x1.cp+02))});
      |                                                                                         ^~~~~~~~~~~
/home/agent/.cache/adamic/runtime/b9a72cc2b2a30b6cfb931643a543124cfa7c9f9ad81d9b2529292fd7856d17f4/adamic.h:48:27: note: passing argument to parameter 'value' here
   48 | void *adamic_retain(void *value);
      |                           ^
1 error generated.

```

Native -O2: exit 1

stdout:
```text
```

stderr:
```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-452493640/main.c:33:82: error: passing 'double' to parameter of incompatible type 'void *'
   33 |         adamic_array_push(adamic_temporary_8, (adamic_value){.reference = adamic_retain((0x1.cp+02))});
      |                                                                                         ^~~~~~~~~~~
/home/agent/.cache/adamic/runtime/9d21974a1fa526de12c8839b47ec1842c6e86b5c7f057464a8998252416960c8/adamic.h:48:27: note: passing argument to parameter 'value' here
   48 | void *adamic_retain(void *value);
      |                           ^
1 error generated.

```

JavaScript backend Node: exit 0

stdout:
```text
2:7:word8
```

stderr:
```text
```

# Mutant packed-leak

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## nested_argument_counts

```typescript
function make(label: string): (head?: string, ...tail: string[]) => string {
    function read(head: string = `default:${label}`, ...tail: string[]): string {
        const count = arguments.length;
        const captured = () => `${count}:${head}:${tail.join('|')}:${label}`;
        return captured();
    }
    label = `${label}!`;
    return read;
}
const run = make(`label${1}`);
console.log(run());
console.log(run(undefined));
console.log(run(`head${2}`, `tail${3}`, `tail${4}`));
const items = [`x${5}`, `y${6}`];
function replace(): string { items[0] = `z${7}`; items.push(`q${8}`); return `last${9}`; }
console.log(run(...items, replace()));
console.log(run(...items));
```

Source Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Separate LSAN: exit 23

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text

=================================================================
==17292==ERROR: LeakSanitizer: detected memory leaks

Direct leak of 112 byte(s) in 2 object(s) allocated from:
    #0 0x55fa4a249274 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
    #1 0x55fa4a2a1dbf in adamic_allocate /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:196:10

Indirect leak of 64 byte(s) in 2 object(s) allocated from:
    #0 0x55fa4a24966c in realloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:81:3
    #1 0x55fa4a28faa5 in adamic_array_push /home/agent/.cache/adamic/runtime/.build-3904243615/array.c:31:25

SUMMARY: AddressSanitizer: 176 byte(s) leaked in 4 allocation(s).
```

# Mutant rest-retain

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## nested_argument_counts

```typescript
function make(label: string): (head?: string, ...tail: string[]) => string {
    function read(head: string = `default:${label}`, ...tail: string[]): string {
        const count = arguments.length;
        const captured = () => `${count}:${head}:${tail.join('|')}:${label}`;
        return captured();
    }
    label = `${label}!`;
    return read;
}
const run = make(`label${1}`);
console.log(run());
console.log(run(undefined));
console.log(run(`head${2}`, `tail${3}`, `tail${4}`));
const items = [`x${5}`, `y${6}`];
function replace(): string { items[0] = `z${7}`; items.push(`q${8}`); return `last${9}`; }
console.log(run(...items, replace()));
console.log(run(...items));
```

Source Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
=================================================================
==14300==ERROR: AddressSanitizer: heap-use-after-free on address 0x7bd677fe0640 at pc 0x558e2c84ffd0 bp 0x7ffe361841c0 sp 0x7ffe361841b8
READ of size 8 at 0x7bd677fe0640 thread T0
    #0 0x558e2c84ffcf in adamic_release /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:372:28
    #1 0x558e2c83b115 in main /tmp/adamic-gate/adamic-build-3589264455/main.c:210:2
    #2 0x7f6678f89ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #3 0x7f6678f89d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #4 0x558e2c754460 in _start (/tmp/coverage-closures/74e55a7600d2/nested_argument_counts.sanitized+0x58460)

0x7bd677fe0640 is located 0 bytes inside of 69-byte region [0x7bd677fe0640,0x7bd677fe0685)
freed by thread T0 here:
    #0 0x558e2c7f6fd6 in free /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:51:3
    #1 0x558e2c850ba3 in deallocate /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:212:3
    #2 0x558e2c850ba3 in free_one /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:351:2
    #3 0x558e2c850ba3 in release_last /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:364:3

previously allocated by thread T0 here:
    #0 0x558e2c7f7274 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
    #1 0x558e2c84fdbf in adamic_allocate /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:196:10

SUMMARY: AddressSanitizer: heap-use-after-free /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:372:28 in adamic_release
Shadow bytes around the buggy address:
  0x7bd677fe0380: fa fa fa fa fd fd fd fd fd fd fd fd fd fa fa fa
  0x7bd677fe0400: fa fa 00 00 00 00 00 00 00 00 01 fa fa fa fa fa
  0x7bd677fe0480: 00 00 00 00 00 00 00 00 05 fa fa fa fa fa 00 00
  0x7bd677fe0500: 00 00 00 00 00 00 01 fa fa fa fa fa fd fd fd fd
  0x7bd677fe0580: fd fd fd fd fd fa fa fa fa fa 00 00 00 00 00 00
=>0x7bd677fe0600: 00 00 01 fa fa fa fa fa[fd]fd fd fd fd fd fd fd
  0x7bd677fe0680: fd fa fa fa fa fa fd fd fd fd fd fd fd fd fd fa
  0x7bd677fe0700: fa fa fa fa fd fd fd fd fd fd fd fd fd fd fa fa
  0x7bd677fe0780: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa
  0x7bd677fe0800: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa
  0x7bd677fe0880: fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa fa
Shadow byte legend (one shadow byte represents 8 application bytes):
  Addressable:           00
  Partially addressable: 01 02 03 04 05 06 07 
  Heap left redzone:       fa
  Freed heap region:       fd
  Stack left redzone:      f1
  Stack mid redzone:       f2
  Stack right redzone:     f3
  Stack after return:      f5
  Stack use after scope:   f8
  Global redzone:          f9
  Global init order:       f6
  Poisoned by user:        f7
  Container overflow:      fc
  Array cookie:            ac
  Intra object redzone:    bb
  ASan internal:           fe
  Left alloca redzone:     ca
  Right alloca redzone:    cb
==14300==ABORTING
```

Native -O2: exit -11

stdout:
```text
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

## rest_capture_lifetime

```typescript
function make(head: string = `default${1}`, ...tail: string[]): () => string {
    const count = arguments.length;
    function read(): string {
        tail.push(`${head}:${tail.length}`);
        return `${count}:${head}:${tail.join('|')}`;
    }
    head = `${head}!`;
    return read;
}
const first = make();
const second = make(undefined, `tail${2}`, `tail${3}`);
console.log(first());
console.log(second());
console.log(first());
console.log(second());
```

Source Node: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Native sanitizer: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Native -O2: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

Separate LSAN: exit 0

stdout:
```text
0:default1!:default1!:0
3:default1!:tail2|tail3|default1!:2
0:default1!:default1!:0|default1!:1
3:default1!:tail2|tail3|default1!:2|default1!:3
```

stderr:
```text
```

# Mutant spread-snapshot

Location and classification are in the mutation table above; these outputs are from the mutant compiler.

## nested_argument_counts

```typescript
function make(label: string): (head?: string, ...tail: string[]) => string {
    function read(head: string = `default:${label}`, ...tail: string[]): string {
        const count = arguments.length;
        const captured = () => `${count}:${head}:${tail.join('|')}:${label}`;
        return captured();
    }
    label = `${label}!`;
    return read;
}
const run = make(`label${1}`);
console.log(run());
console.log(run(undefined));
console.log(run(`head${2}`, `tail${3}`, `tail${4}`));
const items = [`x${5}`, `y${6}`];
function replace(): string { items[0] = `z${7}`; items.push(`q${8}`); return `last${9}`; }
console.log(run(...items, replace()));
console.log(run(...items));
```

Source Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

Native sanitizer: exit 1

stdout:
```text
```

stderr:
```text
=================================================================
==13987==ERROR: AddressSanitizer: heap-use-after-free on address 0x7b935a9e0808 at pc 0x55ae21251f0e bp 0x7ffec6648e60 sp 0x7ffec6648e58
READ of size 4 at 0x7b935a9e0808 thread T0
    #0 0x55ae21251f0d in adamic_retain /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:222:28
    #1 0x55ae2123e01f in adamic_function_2_read /tmp/adamic-gate/adamic-build-3668804346/main.c:103:2
    #2 0x55ae2123d6e6 in adamic_closure_call /home/agent/.cache/adamic/runtime/d9668874c6a962a7e84226ffd52e9a2f0f8566254cb636e71c3ab6c532057304/adamic.h:143:33
    #3 0x55ae2123d6e6 in main /tmp/adamic-gate/adamic-build-3668804346/main.c:244:37
    #4 0x7f235b818ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #5 0x7f235b818d64 in __libc_start_main (/lib/x86_64-linux-gnu/libc.so.6+0x29d64) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)
    #6 0x55ae21156460 in _start (/tmp/coverage-closures/675dcd68415e/nested_argument_counts.sanitized+0x58460)

0x7b935a9e0808 is located 8 bytes inside of 66-byte region [0x7b935a9e0800,0x7b935a9e0842)
freed by thread T0 here:
    #0 0x55ae211f8fd6 in free /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:51:3
    #1 0x55ae21252b63 in deallocate /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:212:3
    #2 0x55ae21252b63 in free_one /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:351:2
    #3 0x55ae21252b63 in release_last /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:364:3
    #4 0x55ae2123d3c8 in adamic_function_1_replace /tmp/adamic-gate/adamic-build-3668804346/main.c:69:2
    #5 0x55ae2123d3c8 in main /tmp/adamic-gate/adamic-build-3668804346/main.c:238:40
    #6 0x7f235b818ca7  (/lib/x86_64-linux-gnu/libc.so.6+0x29ca7) (BuildId: c495b62edadd6c356265942ec1282d98058a7b41)

previously allocated by thread T0 here:
    #0 0x55ae211f9274 in malloc /home/runner/work/llvm-project/llvm-project/compiler-rt/lib/asan/asan_malloc_linux.cpp:67:3
    #1 0x55ae21251d7f in adamic_allocate /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:196:10

SUMMARY: AddressSanitizer: heap-use-after-free /home/agent/.cache/adamic/runtime/.build-3904243615/heap.c:222:28 in adamic_retain
Shadow bytes around the buggy address:
  0x7b935a9e0580: fd fd fd fd fd fa fa fa fa fa fd fd fd fd fd fd
  0x7b935a9e0600: fd fd fd fa fa fa fa fa fd fd fd fd fd fd fd fd
  0x7b935a9e0680: fd fa fa fa fa fa fd fd fd fd fd fd fd fd fd fa
  0x7b935a9e0700: fa fa fa fa fd fd fd fd fd fd fd fd fd fd fa fa
  0x7b935a9e0780: fa fa fd fd fd fd fd fd fd fd fd fa fa fa fa fa
=>0x7b935a9e0800: fd[fd]fd fd fd fd fd fd fd fa fa fa fa fa fd fd
  0x7b935a9e0880: fd fd fd fd fd fd fd fa fa fa fa fa 00 00 00 00
  0x7b935a9e0900: 00 00 00 00 02 fa fa fa fa fa fd fd fd fd fd fd
  0x7b935a9e0980: fd fd fd fa fa fa fa fa 00 00 00 00 00 00 00 00
  0x7b935a9e0a00: 02 fa fa fa fa fa fd fd fd fd fd fd fd fd fd fa
  0x7b935a9e0a80: fa fa fa fa 00 00 00 00 00 00 00 00 02 fa fa fa
Shadow byte legend (one shadow byte represents 8 application bytes):
  Addressable:           00
  Partially addressable: 01 02 03 04 05 06 07 
  Heap left redzone:       fa
  Freed heap region:       fd
  Stack left redzone:      f1
  Stack mid redzone:       f2
  Stack right redzone:     f3
  Stack after return:      f5
  Stack use after scope:   f8
  Global redzone:          f9
  Global init order:       f6
  Poisoned by user:        f7
  Container overflow:      fc
  Array cookie:            ac
  Intra object redzone:    bb
  ASan internal:           fe
  Left alloca redzone:     ca
  Right alloca redzone:    cb
==13987==ABORTING
```

Native -O2: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:q8:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```

JavaScript backend Node: exit 0

stdout:
```text
0:default:label1!::label1!
1:default:label1!::label1!
3:head2:tail3|tail4:label1!
3:x5:y6|last9:label1!
3:z7:y6|q8:label1!
```

stderr:
```text
```
