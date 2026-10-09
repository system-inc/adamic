Built: original JSDoc parser slice and flat parser factory var bindings, with structured missing-symbol locations.
Commits: ddb51eb (JSDoc), 39c3eca (factory), 99444ae and ae5ac3c (main merges); final pushed feature tip is in the final response.
Checks: seven affected packages pass uncached; fresh matrix is 5 Compiles, 5 NotYet, 2 Refused; counts, vet and formatting pass.
Mutants: eleven runtime/equality mutations and ten diagnostic-boundary mutations caught; the complete table follows.
Not covered: full unchanged parser.ts, escaped runtime namespace objects, callable/class merges, primitive brand construction or new truthiness policy.

# Remaining original namespace fixtures and decisions

Feature branch only: no push to main or an area branch, and no merge into either.
Main e011f8f was first merged into codex/namespaces-tsc in 99444ae. A final fetch
found main e8ba3d5, merged in ae5ac3c. The sole locals.go conflict keeps both
namespace hoisted assignments and main's const-closure target metadata. The final
feature tip is recorded in the final response after verification and push.

The requested order was 08, 01, 05, 10, 02, 03. 08 gained native support in
previously pushed ddb51eb; the late-priority parser factory cut gained support
in 39c3eca before further fixture work. 01 and 05 require the boolean-only
language decision; 10 remains the explicitly permitted runtime-container
NotYet; 02 and 03 require a branded-primitive construction decision. Neither
Refused fixture was changed. These are judgments for integration review, not
new language decisions.

## Fresh matrix after merging main

The all-twelve matrices against 99444ae and ae5ac3c are five Compiles, five NotYet, two Refused,
zero checker failures. Every successful build matches fresh source Node, under
ASan/UBSan with leak checks and through checked JavaScript. The merge changes
no namespace outcome row. The factory cut is an additional oracle fixture and
therefore changes no count in this twelve-fixture matrix.

| Step | Compiles | NotYet | Refused | Changed row |
| --- | ---: | ---: | ---: | --- |
| Fixture branch merge | 4 | 6 | 2 | None |
| 08 returned singleton assignment | 5 | 5 | 2 | 08 NotYet to Compiles |
| Parser factory var bindings | 5 | 5 | 2 | Additional parser cut, outside the twelve |
| 01 BuilderState review | 5 | 5 | 2 | None, language decision |
| 05 Debug.log review | 5 | 5 | 2 | None, callable container and language decision |
| 10 tracing review | 5 | 5 | 2 | None, real container unimplemented |
| 02 JsxNames review | 5 | 5 | 2 | None, primitive-brand decision |
| 03 ReactNames review | 5 | 5 | 2 | None, primitive-brand decision |
| Current main merge and fresh matrix | 5 | 5 | 2 | None |

Below are the exact unchanged programs, the complete fresh build diagnostic,
Node's output, and the judgment. Absolute prefixes in diagnostics reflect this
worker checkout; line and column refer to the committed fixture.

## 01_builder_state.a

```typescript
// From TypeScript 6.0.3, src/compiler/builderState.ts:62
// From TypeScript 6.0.3, src/compiler/builderState.ts:72
// From TypeScript 6.0.3, src/compiler/builderState.ts:100
// From TypeScript 6.0.3, src/compiler/builderState.ts:295
// Census reason: a namespace
// Support types retain only the state contract read by canReuseOldState.
export interface BuilderState {
    readonly referencedMap?: BuilderState.ReadonlyManyToManyPathMap | undefined;
}
export namespace BuilderState {
    export interface ReadonlyManyToManyPathMap { readonly marker: number; }
    export function canReuseOldState(newReferencedMap: ReadonlyManyToManyPathMap | undefined, oldState: BuilderState | undefined): boolean | undefined {
        return oldState && !oldState.referencedMap === !newReferencedMap;
    }
}
const map: BuilderState.ReadonlyManyToManyPathMap = { marker: 1 };
for (const oldState of [undefined, {}, { referencedMap: map }]) {
    console.log(`${BuilderState.canReuseOldState(undefined, oldState)} ${BuilderState.canReuseOldState(map, oldState)}`);
}
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/01_builder_state.a:13:28: stage 0 can't lower a PrefixUnaryExpression on a value yet
```

Independent Node (exit 0, empty stderr):

```text
undefined undefined
true false
false true
```

Judgment: the interface/namespace merge itself is supported. The blocker is object truthiness in `oldState && !oldState.referencedMap === !newReferencedMap`. docs/0.1.md explicitly admits &&, || and ! only on booleans, and identifies boolean-only conditions as a language choice rather than a type-soundness rule. The current NotYet describes the unimplemented operator; accepting the unchanged body needs a language decision. The NotYet classification understates a current 0.1 policy exclusion; it is not merely a missing namespace feature. No truthiness extension was made.

Narrowest sound extension: for exactly object-or-undefined operands, prove every defined object is truthy. Evaluate the left operand once; !x is x === undefined, and x && right is undefined when x is undefined and otherwise evaluates right once. Preserve the boolean-or-undefined result and short-circuiting. Do not apply a numeric, string, any, or arbitrary-union truthiness conversion. A reviewed source port can spell the tests explicitly, but it would not be the unchanged fixture.

## 05_debug_log_merge.a

```typescript
// From TypeScript 6.0.3, src/compiler/debug.ts:99
// From TypeScript 6.0.3, src/compiler/debug.ts:113
// From TypeScript 6.0.3, src/compiler/debug.ts:123
// From TypeScript 6.0.3, src/compiler/debug.ts:137
// Census reason: a namespace
export enum LogLevel {
    Off,
    Error,
    Warning,
    Info,
    Verbose,
}

/** @internal */
export interface LoggingHost {
    log(level: LogLevel, s: string): void;
}
export namespace Debug {
    export let currentLogLevel: LogLevel = LogLevel.Warning;
    export let isDebugging = false;
    export let loggingHost: LoggingHost | undefined;
    /* eslint-enable prefer-const */


    export function shouldLog(level: LogLevel): boolean {
        return currentLogLevel <= level;
    }

    function logMessage(level: LogLevel, s: string): void {
        if (loggingHost && shouldLog(level)) {
            loggingHost.log(level, s);
        }
    }

    export function log(s: string): void {
        logMessage(LogLevel.Info, s);
    }

    export namespace log {
        export function error(s: string): void {
            logMessage(LogLevel.Error, s);
        }

        export function warn(s: string): void {
            logMessage(LogLevel.Warning, s);
        }

        export function log(s: string): void {
            logMessage(LogLevel.Info, s);
        }

        export function trace(s: string): void {
            logMessage(LogLevel.Verbose, s);
        }
    }
}
const messages: string[] = [];
Debug.loggingHost = { log: (level, s) => { messages.push(`${level}:${s}`); } };
Debug.log("info at warning");
Debug.log.error("error");
Debug.log.warn("warning");
Debug.currentLogLevel = LogLevel.Verbose;
Debug.log("info");
Debug.log.log("nested info");
Debug.log.trace("trace");
console.log(messages.join("|"));
console.log(`${Debug.log === Debug.log} ${Debug.log === Debug.log.log}`);
Debug.loggingHost = undefined;
Debug.log.error("no host");
console.log(`${messages.length}`);
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/05_debug_log_merge.a:39:5: stage 0 can't lower a namespace merged with a function; callable object properties, identity and receivers are not represented yet
```

Independent Node (exit 0, empty stderr):

```text
3:info at warning|2:warning|4:trace
true false
3
```

Judgment: this is a necessary implementation NotYet, not a type lie in the fixture. Debug.log is both callable and an object, with log.log a distinct callable property. Native closure IR and ordinary object IR do not currently represent that single merged identity. Registering the members as independent functions or treating the parent as an ordinary object would not implement both calls and equality. The original logMessage body additionally uses object-or-undefined loggingHost as the left operand of &&, a separate boolean-only language decision. Stop here without treating removal of the merge guard as support.

Narrowest future implementation: a single canonical callable container with fixed attached functions, correct call receivers and ownership; reject reopenings, property replacement and unsupported receiver use until those are represented. The body still needs the object-or-undefined short-circuit decision described for 01, or a reviewed explicit loggingHost !== undefined source port.

## 10_tracing_escape.a

```typescript
// From TypeScript 6.0.3, src/compiler/tracing.ts:37
// From TypeScript 6.0.3, src/compiler/tracing.ts:45
// From TypeScript 6.0.3, src/compiler/tracing.ts:95
// From TypeScript 6.0.3, src/compiler/tracing.ts:126
// Census reason: a namespace
interface Type { readonly id: number; }
export namespace tracingEnabled {
    type Mode = "project" | "build" | "server";
    let mode: Mode;

    const typeCatalog: Type[] = []; // NB: id is index + 1
    export function recordType(type: Type): void {
        if (mode !== "server") {
            typeCatalog.push(type);
        }
    }

    export const enum Phase {
        Parse = "parse",
        Program = "program",
        Bind = "bind",
        Check = "check", // Before we get into checking types (e.g. checkSourceFile)
        CheckTypes = "checkTypes",
        Emit = "emit",
        Session = "session",
    }
    export function run(): void {
        mode = "project";
        recordType({ id: 1 });
        mode = "server";
        recordType({ id: 2 });
        mode = "build";
        recordType({ id: 3 });
        console.log(typeCatalog.map(type => `${type.id}`).join(","));
    }
}
let tracing: typeof tracingEnabled | undefined;
tracing = tracingEnabled; // only when traceFd is properly set
tracing.run();
console.log(tracing.Phase.Parse);
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/10_tracing_escape.a:38:11: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet
```

Independent Node (exit 0, empty stderr):

```text
1,3
parse
```

Judgment: necessary implementation NotYet for the current qualified-only lowering. No language change is intrinsically needed for this fixture. Its private mutable state is already supported, but its alias needs an actual, canonical namespace container. A snapshot of public values is not general namespace support: mutable exports need live slots, and even this function-only public shape permits function replacement through a typeof alias in TypeScript. Existing qualified calls would continue to use the original flattened function. Enum container identity, const-enum observation under the source Node transform, initialization order and closure ownership also need verification. This work has not established those container semantics; the initial unit's explicit 'small and sound, else NotYet' boundary remains in force.

A potentially narrow implementation is one namespace declaration, no runtime merges or reopenings, only fixed exported ordinary functions and constant-valued enums, private singleton state, and escape after initialization. It must emit one real object with canonical function values and its enum properties, and either route qualified and alias calls through the same live properties or reject all export replacement and every writable structural view, including casts, arguments, returns and generic views. Merely rejecting direct qualified assignments is insufficient. This is a future implementation boundary, not evidence of a completed runtime container.

The additional alias-replacement probe below is checker-accepted. Fresh Node prints replacement; the compiler retains the explicit container NotYet. A copied object plus flattened calls would print original, so that shortcut would silently miscompile. No such implementation was committed.

## 02_jsx_names.a

```typescript
// From TypeScript 6.0.3, src/compiler/checker.ts:54223
// From TypeScript 6.0.3, src/compiler/types.ts:6196
// Census reason: a namespace
// Keep the string arm of the upstream escaped-name type.
type __String = string & { __escapedIdentifier: void; };
namespace JsxNames {
    export const JSX = "JSX" as __String;
    export const IntrinsicElements = "IntrinsicElements" as __String;
    export const ElementClass = "ElementClass" as __String;
    export const ElementAttributesPropertyNameContainer = "ElementAttributesProperty" as __String; // TODO: Deprecate and remove support
    export const ElementChildrenAttributeNameContainer = "ElementChildrenAttribute" as __String;
    export const Element = "Element" as __String;
    export const ElementType = "ElementType" as __String;
    export const IntrinsicAttributes = "IntrinsicAttributes" as __String;
    export const IntrinsicClassAttributes = "IntrinsicClassAttributes" as __String;
    export const LibraryManagedAttributes = "LibraryManagedAttributes" as __String;
}
console.log(`${JsxNames.JSX} ${JsxNames.IntrinsicElements} ${JsxNames.ElementClass}`);
console.log(`${JsxNames.ElementAttributesPropertyNameContainer} ${JsxNames.ElementChildrenAttributeNameContainer}`);
console.log(`${JsxNames.Element} ${JsxNames.ElementType} ${JsxNames.IntrinsicAttributes} ${JsxNames.IntrinsicClassAttributes} ${JsxNames.LibraryManagedAttributes}`);
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/02_jsx_names.a:7:18: stage 0 can't lower a value of type __String yet
```

Independent Node (exit 0, empty stderr):

```text
JSX IntrinsicElements ElementClass
ElementAttributesProperty ElementChildrenAttribute
Element ElementType IntrinsicAttributes IntrinsicClassAttributes LibraryManagedAttributes
```

Judgment: namespace constants already lower; the first blocker is the primitive/object intersection __String. There is also an independent unchecked cast from an ordinary string to that brand. It cannot be made sound by silently erasing the object arm and trusting the cast. Stop on the language decision.

Narrowest possible decision: explicitly defined primitive phantom brands with a declared construction rule, preservation in type relations, and refusal of object-style reads/writes and unchecked conversions that are not authorized brand construction. A verified constructor or a reviewed string-only source port is another option. Blanket primitive/object intersection erasure is not proposed. The direct cast probe below reaches the current Refused even without the namespace-local type representation blocker.

## 03_react_names.a

```typescript
// From TypeScript 6.0.3, src/compiler/checker.ts:54236
// From TypeScript 6.0.3, src/compiler/types.ts:6196
// Census reason: a namespace
// Keep the string arm of the upstream escaped-name type.
type __String = string & { __escapedIdentifier: void; };
namespace ReactNames {
    export const Fragment = "Fragment" as __String;
}
console.log(ReactNames.Fragment);
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/03_react_names.a:7:18: stage 0 can't lower a value of type __String yet
```

Independent Node (exit 0, empty stderr):

```text
Fragment
```

Judgment: the same __String representation and unchecked-brand-construction decision as 02. Its one exported constant is supported namespace structure. Do not turn an unchecked branded cast into an upcast by erasing its target. The narrowest acceptance and remaining obligations are those recorded for 02.

## 06_binary_expression_state.a

```typescript
// From TypeScript 6.0.3, src/compiler/factory/utilities.ts:1271
// From TypeScript 6.0.3, src/compiler/factory/utilities.ts:1365
// From TypeScript 6.0.3, src/compiler/factory/utilities.ts:1273
// Census reason: a namespace
// Minimal machine/node support; done does not read their fields.
interface BinaryExpressionStateMachine<TOuterState, TState, TResult> {}
interface BinaryExpression { readonly kind: number; }
const Debug = { assertEqual: <T>(a: T, b: T): void => { if (a !== b) { throw new Error("state mismatch"); } } };
type BinaryExpressionState = <TOuterState, TState, TResult>(machine: BinaryExpressionStateMachine<TOuterState, TState, TResult>, stackIndex: number, stateStack: BinaryExpressionState[], nodeStack: BinaryExpression[], userStateStack: TState[], resultHolder: { value: TResult; }, outerState: TOuterState) => number;

namespace BinaryExpressionState {
    export function done<TOuterState, TState, TResult>(_machine: BinaryExpressionStateMachine<TOuterState, TState, TResult>, stackIndex: number, stateStack: BinaryExpressionState[], _nodeStack: BinaryExpression[], _userStateStack: TState[], _resultHolder: { value: TResult; }, _outerState: TOuterState): number {
        Debug.assertEqual(stateStack[stackIndex], done);
        return stackIndex;
    }
}
const stateStack: BinaryExpressionState[] = [BinaryExpressionState.done];
console.log(`${BinaryExpressionState.done({}, 0, stateStack, [{ kind: 227 }], [0], { value: 0 }, 0)}`);
const detached = BinaryExpressionState.done;
console.log(`${detached === BinaryExpressionState.done}`);
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/06_binary_expression_state.a:13:27: Adamic 0.1 refuses a function taking TState[] seen as one taking TState[] (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

Independent Node (exit 0, empty stderr):

```text
0
true
```

Judgment: too strict for this exact program. The reported TState[] versus TState[] comparison comes from distinct generic signature binders: invariance.go compares the raw parameter types from the two signatures without alpha-renaming their universally bound type parameters. These callable signatures are alpha-equivalent in this fixture; there is no Animal/Dog-style narrower parameter being handed a wider argument. The refusal is preserved exactly as requested.

Narrowest sound acceptance: prove alpha-equivalence of the universal signatures under a bijection of fresh binders. Require equal binder count/order, matching constraints/defaults after substitution, matching this/optional/rest parameter structure, and equal parameter and return types under that substitution, retaining invariant mutable arrays and fields. Do not accept all generics, compare only printed names, or bypass the ordinary function-variance check. Allowing this one relation may expose a separate NotYet for generic first-class function representation; it does not itself prove this fixture will then lower.

## 09_incremental_parser.a

```typescript
// From TypeScript 6.0.3, src/compiler/parser.ts:9946
// From TypeScript 6.0.3, src/compiler/parser.ts:10130
// From TypeScript 6.0.3, src/compiler/parser.ts:10550
// Census reason: a namespace
enum SyntaxKind { EndOfFileToken = 1, NumericLiteral = 9, StringLiteral = 11, Identifier = 80 }
interface Node { readonly kind: SyntaxKind; }
namespace IncrementalParser {
    function shouldCheckNode(node: Node) {
        switch (node.kind) {
            case SyntaxKind.StringLiteral:
            case SyntaxKind.NumericLiteral:
            case SyntaxKind.Identifier:
                return true;
        }

        return false;
    }
    const enum InvalidPosition {
        Value = -1,
    }
    export function run(): void {
        for (const kind of [SyntaxKind.StringLiteral, SyntaxKind.NumericLiteral, SyntaxKind.Identifier, SyntaxKind.EndOfFileToken]) {
            console.log(`${kind} ${shouldCheckNode({ kind })} ${InvalidPosition.Value}`);
        }
    }
}
IncrementalParser.run();
```

Exact diagnostic (build exit 1):

```text
adamic: /workspace/adamic/stage3/fixtures/namespaces/09_incremental_parser.a:9:9: Adamic 0.1 refuses a non-exhaustive enum switch; missing SyntaxKind.EndOfFileToken; add the missing member's case or a default
```

Independent Node (exit 0, empty stderr):

```text
11 true -1
9 true -1
80 true -1
1 false -1
```

Judgment: necessary under the current explicit 0.1 exhaustive-switch rule, but too strict as a type-soundness requirement for this program. All three covered members return true; the uncovered EndOfFileToken falls through the switch to return false. Every function path returns a boolean, and independent Node demonstrates the fourth result. The refusal is preserved exactly as requested.

Narrowest sound language acceptance: permit an incomplete enum switch when control-flow analysis proves every uncovered path reaches a post-switch return with the declared result type, with no unchecked member read, implicit undefined return or unsafe case fallthrough. Preserve the rule for code where that proof fails. Alternatively a reviewed source port can add default: return false; that is not the unchanged fixture.

## Additional boundary probes

### Replacing an exported function through a namespace alias

```typescript
namespace tracingEnabled {
    export function run(): void { console.log('original'); }
}
let tracing: typeof tracingEnabled | undefined;
tracing = tracingEnabled;
tracing.run = (): void => { console.log('replacement'); };
tracingEnabled.run();
```

Independent Node: `replacement`, exit 0, empty stderr. Current compiler:

```text
adamic: /tmp/namespaces-decisions/alias-replacement.a:5:11: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet
exit status 1
```

### Brand cast independent of namespace representation

```typescript
type __String = string & { __escapedIdentifier: void; };
console.log('JSX' as __String);
```

Current compiler:

```text
adamic: /tmp/namespaces-decisions/brand.a:2:13: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)
exit status 1
```

Both probes use driver programs in /tmp/namespaces-decisions, not modifications
to the original twelve fixtures. Probe logs are alongside those files.

## Counts and validation evidence

The added parser-factory counts row is `8 | 8 | 7 | 17 | 6 | 0`, representing
allocations, frees, retains, releases, peak live, and in-region values, respectively. Existing rows did not change for that fix.
The current-main merge carries the borrowed-element lifetime rows from its
already committed counts table; this branch does not reinterpret those rows.
The returned-assignment row and original stages, all runtime and diagnostic
mutants, setup timing and previous package results are recorded in
[PROGRESS.md](PROGRESS.md), [REPORT.md](REPORT.md), and
[REAL_FIXTURES.md](REAL_FIXTURES.md).

The five accepted unchanged slices are 04 Debug state, 07 Parser singleton,
08 Parser.JSDocParser, 11 Status and 12 BuilderState releaseCache. These are
selected original bodies. Full unchanged parser.ts and the complete TypeScript
compiler do not run natively from this work.

The merged-main counts changes are inherited from e011f8f's borrowed-element
lifetime fixes, not namespace lowering: virtual calls now stop array-element
borrowing because the statically named method cannot prove the override's
effects, while constructing an Error without user code is admitted by the
borrow analysis and its operands remain checked. Two new oracle programs cover
throw and virtual-store lifetimes. In counts-table order:

| Row | Before | After | Reason |
| --- | --- | --- | --- |
| borrow_element_throw.a | Absent | 38, 38, 34, 54, 8, 0 | New exception-lifetime oracle |
| borrow_element_virtual_store.a | Absent | 13, 11, 9, 15, 7, 2 | New virtual-call mutation oracle |
| borrow_element_virtual_move.a | 26, 24, 10, 22, 12, 2 | 25, 23, 11, 26, 11, 2 | Virtual-call borrow rejection changes reuse and ownership accounting |

The parser-factory fixture is 8 allocations, 8 frees, 7 retains, 17 releases,
peak 6, zero in-region values. Its complete release count and the sanitized
execution hold its new singleton bindings to the existing ownership rules.

## Final main merge and boundary mutants

Main advanced to e8ba3d5 during the first broad gate. That superseded full
`ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` was stopped before
changing its source; exit 143, with no completed affected-package result claimed
from it. Its log is /tmp/namespaces-landing-full-gate.log. The replacement gate
covers every affected compiler package, including main's merged call-target and
devirtualization changes, uncached.

The first focused merged-main command selected the ten permanent mutant probes
but no ordinary fixture subtests. It passed in 10.866s and is evidence only for
those mutants. The corrected explicit path-segment filter selects all fifteen
namespace oracle fixtures and passed uncached in 10.252s: native cache hits 0,
misses 44; Node cache hits 0, misses 30. Logs:
/tmp/namespaces-merged-focused-oracle.log and
/tmp/namespaces-merged-focused-fixtures.log. Those runs used 99444ae; the final
merged code is checked by the replacement gate and another twelve-fixture row.

All three new destructuring boundaries were separately mutated in an isolated
archive of the real compiler, referring to the existing cohere submodule through
a symlink, with no cohere code copied. Removing the identifier-leaf guard makes
the nested-pattern test fail with the structured missing-symbol diagnostic.
Removing the no-default guard makes its test fail with a later generic
destructuring NotYet instead of the specific namespace reason. Removing the
no-rest guard is likewise required to fail the namespace boundary assertion.
These are diagnostic-boundary mutants, not claims of correct code generation
for any unsupported pattern. The first harness attempt omitted bridge/tsgo/spec
and failed at setup, so it is not counted as a caught mutant. The corrected
harness requires an actual TestNamespaceLimitsStayLoud assertion failure and
restores each mutation in finally. Logs:
/tmp/namespaces-binding-guard-mutants-rerun.log and
/tmp/namespaces-binding-guard-mutants/{nested,default,rest,restored}.log.

The latest-main merge adds five call_targets_* counts rows for the prerequisite's
call-target analysis and one devirtualize.a row. Its existing class_as_interface.a
row changes retains 360 to 349 and releases 525 to 514 because eleven NULL
closure lookup temporaries disappear when exact interface receivers call the
adapter directly. Its allocations 372, frees 372, peak 60 and regions 0 stay
unchanged. Exact inherited rows, in counts-table order:

| New row | Counts |
| --- | --- |
| call_targets_element.a | 45, 43, 36, 64, 7, 2 |
| call_targets_region.a | 27, 22, 13, 33, 9, 5 |
| call_targets_reuse.a | 15, 12, 9, 21, 6, 3 |
| call_targets_closure.a | 25, 25, 8, 31, 6, 0 |
| call_targets_sort.a | 11, 11, 9, 15, 7, 0 |
| devirtualize.a | 43, 39, 26, 66, 10, 4 |

These are main's changes; [its devirtualization report](../../docs/devirtualize.md)
explains their independent Node validation and mutants.

## Final verification commands

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/oracle ./internal/ir ./internal/flow ./internal/fresh > /tmp/namespaces-final-packages.log 2>&1
python3 stage3/namespaces/progress-matrix.py --label 'Latest main with closure targets and devirtualization' --scratch /tmp/namespaces-progress-final > /tmp/namespaces-progress-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/namespaces-final-counts.log 2>&1
go vet ./... > /tmp/namespaces-final-vet.log 2>&1
gofmt -l cmd internal > /tmp/namespaces-final-format.log
git diff --check
```

The fresh final matrix passed. Counts regeneration passed in 75.939s and changed
no row further. Vet exited 0 with no diagnostics; formatting output is empty,
and diff checking passes. Neither full-repository completion nor stage 1's
corpus is claimed; the superseded full gate was stopped and replaced with the
affected compiler packages permitted by this unit's slow-gate exception.

The final affected-package gate exited 0 on ae5ac3c. Package times:
load 2.348s, lower 68.914s, native 333.435s, oracle 323.371s, IR 28.308s,
flow 237.691s and freshness 83.920s. The final remote main check still reports
e8ba3d5d81de4d3773c723914fccd4c76248b965; its result is saved in
/tmp/namespaces-main-at-finish.log. All compiler code is unchanged after this
gate; the concluding commit contains documentation and matrix observations.

## Complete unit mutant table

| Mutation | What caught it | Evidence |
| --- | --- | --- |
| Wrong scoped function result | Independent Node stdout, clean native exit and sanitizers | REPORT.md |
| Wrong scoped constant | Independent Node stdout, clean native exit and sanitizers | REPORT.md |
| Parser enum SourceElements becomes 99 | Independent Node stdout, clean native exit and sanitizers | REPORT.md |
| Swap initialized namespace/module outputs | Independent Node stdout, clean native exit and sanitizers | REPORT.md |
| Outside exported state 3 becomes 30 | Independent Node stdout, clean native exit and sanitizers | REPORT.md |
| Drop parser string initialization assignment | Node comparison catches premature ready-check failure | REPORT.md |
| Remove all hoisted token declarations | Node comparison catches premature read failure | REPORT.md |
| Remove singleton ready check | Checked JavaScript exit comparison | REPORT.md |
| Parser countNode++ becomes += 2 | Fresh matrix Node equality assertion, clean sanitized native | REAL_FIXTURES.md |
| Drop returned singleton assignment | Node stdout, clean sanitized native | PROGRESS.md |
| Wrong same-signature factory method | Node stdout, clean sanitized native | PROGRESS.md |
| Remove escaped-container guard | Specific NotYet-reason assertion | REPORT.md |
| Remove function merge guard | Callable-container NotYet assertion | REPORT.md |
| Remove class merge guard | Constructor-identity NotYet assertion | REPORT.md |
| Remove namespace-class guard | Namespace constructor NotYet assertion | REPORT.md |
| Remove control-flow var guard | Namespace block-var NotYet assertion | REPORT.md |
| Remove returned-assignment scalar boundary | Reference/optional returned assignment NotYet assertion | PROGRESS.md |
| Replace located NotYet with ordinary missing-symbol error | Structured diagnostic Where assertion | PROGRESS.md |
| Remove flat-pattern identifier-leaf guard | Nested namespace var diagnostic assertion | This report |
| Remove flat-pattern no-default guard | Default namespace var diagnostic assertion | This report |
| Remove flat-pattern no-rest guard | Rest namespace var diagnostic assertion | This report |

All real compiler guard mutations were restored. Exploratory surviving or
invalid attempts are recorded in the linked reports and above, and are not
counted as caught mutants. The final isolated restored lower run passed in
6.928s. The ten permanent namespace runtime mutants also run in the final
uncached oracle package; the source matrix mutation and removed-guard probes
have their separate recorded evidence.
