# Step 18: optional calls and the rest of `?`

Branch `codex/scout-optional-calls`, based on the explicit area-next landing
candidate `8cb5e7c1`. This report serves roadmap step 18 and task #ht2nwj5.

## Where it bites in TypeScript

Observed: a fresh no-output latent census of the current stage3-adapted
TypeScript compiler directory on this base. TypeScript is pinned to 6.0.3,
upstream `050880ce59e30b356b686bd3144efe24f875ebc8`; `stage3/apply.sh`
produced the adapted source. The census measurement says **checker-rejected**:
it inventories compiler barriers without claiming those programs are accepted.

The supplied hidden ranking at `6c4fc1af` was measured with compiler
`ed6e29751ee47d86fad450cd1674139883bc0f70`, not this base. Its hidden-byte
numbers below are historical exclusive bytes revealed if that reason alone were
fixed, not measured bytes retired on this branch. Keep that provenance separate
from the fresh root counts. Ranking boundary counts are distinct boundary spans,
not the fresh diagnostic-root counts. Full normalized roots and attribution
metadata are in [inventory.json](step-18/inventory.json).

A root here is a distinct `(kind, file:line:column, reason, text)` finding across
the compiler census records. Repeated entry observations count once. Witnesses
are source positions in the pinned adapted compiler; `H` marks a historical
ranking witness when fewer than three fresh positions exist. No witness is
invented to meet a quota.

| Kind and exact reason | Base roots | Historical hidden bytes | Historical boundary spans | Three file:line witnesses |
|---|---:|---:|---:|---|
| NotYet: `a call through ?. (an optional call)` | 136 | 4053 | 136 | `builder.ts:670`; `builder.ts:710`; `builder.ts:1852` |
| NotYet: `an optional chain longer than one step` | 11 | 2106 | 11 | `checker.ts:4745`; `checker.ts:36469`; `checker.ts:53798` |
| NotYet: `?. to a number, which would be number \| undefined` | 3 | 1964 | 3 | `checker.ts:10416`; `checker.ts:10445`; `utilities.ts:8102` |
| NotYet: `optional chaining to .size on a value` | 20 | 341 | 20 | `builder.ts:705`; `builder.ts:756`; `builder.ts:778` |
| NotYet: `?.[] on a value` | 6 | 58 | 6 | `checker.ts:9101`; `checker.ts:22931`; `checker.ts:33479` |

The five syntax reasons account for 176 fresh
roots and 8,522 historical exclusive hidden bytes. All five are classified
`compiler lesson` by the supplied NotYet table; its owner column is absent.
The ranking's refusal table is pinned to `d35a81d`, and its NotYet table to
`dc6b1529`. Neither supplies an optional-call acceptance policy change.

There is no separate Refused reason for these optional-syntax shapes in the
ranking or the fresh census: zero such roots, zero attributed hidden bytes,
and no witnesses. Refused relations mentioning an *optional field* are the
optional-widening and mutable-relation soundness rules, not optional syntax.
`writing a possibly absent optional own field`, `reading optional`, and checked
view `_optionalChainBrand` representation are also separate lessons. They stay
outside this retirement count. A generic refusal can block a program that also
contains `?.`; that does not turn its refusal into a step-18 syntax root.

### Reproduction and limits

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scout-optional-setup.log 2>&1
source /workspace/adamic-tools/env.sh
STAGE3_CACHE=/workspace/scratch/scout-stage3-cache bash stage3/apply.sh /workspace/scratch/scout-optional-adapted > /tmp/scout-optional-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/scout-optional-overlay > /tmp/scout-optional-overlay.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/scout-optional-overlay/overlay.json -o /workspace/scratch/scout-optional-census ./stage3/census/latent/tool > /tmp/scout-optional-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/scout-optional-census /workspace/scratch/scout-optional-adapted/src/compiler /workspace/scratch/scout-optional-base.jsonl > /tmp/scout-optional-base-census.log 2>&1
```

The output guard returned measurement records without usable IR. This run
covers the compiler-directory census, not the tsc CLI entry census. A retired
root means this exact local barrier disappeared; later NotYet or Refused
barriers and checker errors can remain. Historical hidden bytes must not be
reported as successful compilation or additive porting progress.

## Design and policy boundary

This is a compiler lesson within 0.1's optional syntax, not permission to weaken the
checker, checked views, readiness, cycle proof, library contracts, or signature
proof. Follow [0.1](0.1.md), [escape hatches](escape-hatches.md), and the memory
model. A NotYet becoming implemented is distinct from a Refused program becoming
accepted. No refusal rule changes are proposed for implementation here.

| Source shape | Representation and lowering | Present limitation |
|---|---|---|
| `f?.(args)` | Save the closure reference once in an owned synthesized local. Test nullishness. In the present arm call that saved closure with the original signature ABI; widen only the result to the optional expression's representation. In the absent arm return undefined. Discarded void calls evaluate for effects and return undefined. | Only supported, proven closure signatures and representations; no erased callable markers or signature invention. Observed void results remain NotYet until their runtime return value is proven. |
| `obj.method?.(args)` | Evaluate the receiver and select the callable once. For an ordinary closure field, save that callable before evaluating arguments, then use the same conditional as a free callable. A receiver-bearing method needs a descriptor retaining both the selected function and receiver; existing method IR resolves its receiver only when invoked and cannot be saved as a plain closure. Optionality is on the callable, not on `obj`. | Receiver-bearing methods, including methods hidden behind structural callback views, remain NotYet. Library callable values retain their own lowering contracts. Earlier optional-chain propagation needs the general chain design below. |
| `a?.[i]` | Save `a`, guard it, and evaluate `i` only in the present arm. Keep the existing array, string or tuple access semantics, including out-of-range undefined; widen the result after access. | Arrays and strings with optional access remain NotYet in this unit. Tuple support already present is not credited to this work. |
| `get()?.field` and `get()?.field.more` | Save the call result and guard that value. A continuous chain carries an explicit short-circuit exit through all following ordinary accesses and calls. Parentheses end a continuous chain: `(get()?.field).more` must still fail on an absent field. | General multi-step propagation remains NotYet. Existing one-step reads retain their current coverage. |
| `obj?.method?.(args)` | Separate guards for receiver absence and selected callable absence, with one shared undefined exit. Each receiver, selector and argument has its source evaluation position. | Two guards and propagated optional flags remain NotYet here. |
| Chains crossing a call that reassigns a binding | Saved values are owned snapshots; later operations read the source binding anew only where JavaScript does. A saved receiver or function survives a reassignment during an argument. Checker narrowing across calls is not proof of runtime presence. | General chain lowering must retain runtime defined/readiness checks for unguarded later accesses; it must never reuse a stale narrowing or stale alias as the chain value. |

The callable's signature return type and the optional expression's type are two
different types. For example `(() => number) | undefined` is a nullable closure,
its call returns a plain number, and `f?.()` returns a maybe-number. The lowering preserves that distinction. Observed: substituting the optional
expression return descriptor survived the runtime-values mutant. Inference:
the existing packed-call adapters normalize this difference for these cases;
that mutation does not prove an ABI failure and is not credited as killed.
Use the non-nullable callable only for resolving its signature, retain the
original callable type identity for closure-target/count dispatch, and pack the
result after the call. Do the same for boolean and reference results. Unsupported
unions must give NotYet before either backend emits code.

The synthesized closure local uses existing `Declare`, `Read`, `Effects`,
`Conditional`, `IsUndefined`, `IsNull` and `CallClosure` IR. Its declaration takes
ownership under the existing emitter conventions and its lifetime includes the
argument evaluations and the call. The JavaScript backend lowers the same IR.
No backend-specific interpretation of optional syntax or new garbage collection
is needed. An optional call is a presence observation: reads feeding its guard
must not first run the check that demands a checker-narrowed value be present.
This also matters for a statically present array element that is absent at runtime.

### Places a silent miscompile could hide

- Evaluating a callee, receiver, computed key, getter or call-result twice.
- Evaluating arguments, a spread, or an index on the absent path.
- Reading the callable again after an argument replaces its variable or field.
- Losing the receiver when calling a method, or binding a free function to one.
- Treating falsy values as absent, or testing undefined while forgetting null.
- Using the optional result's ABI as the closure's return ABI, including void,
  maybe-number, maybe-boolean, null and reference results.
- Failing to fit absent and present results to the same representation.
- Propagating a chain beyond parentheses, or ending it before an ordinary access.
- Reusing checker narrowing across a call that writes the binding. A real guard
  observes runtime absence; an unguarded use still needs its defined check.
- Releasing a saved receiver, selected callable, captured cell or dynamic result
  before a sibling argument or outer expression finishes; leaking the snapshot
  on return, a loop, or an absent path.
- Losing closure-target metadata, argument-count/default/rest adaptation, checked
  view readiness, cycle accounting, or ordinary call effects in flow analyses.
- Letting library optional syntax fall through an ordinary call path without its
  intrinsic contract, or accepting a callable cast on a `typeof` check alone.

### Refused programs and proposals for @system_adamic

Optional syntax does not make a bad type claim true. These are policy examples,
not implemented escape hatches; the conservative proposal is to keep refusing
all three, and to keep optional syntax downstream of the same refusal pass.
Any future acceptance change requires @system_adamic's decision with these
programs and a sound argument/return boundary design.

```a
// A callable shape does not prove an asserted signature.
const external: unknown = (text: string): number => text.length;
const claimed = external as ((n: number) => string) | undefined;
console.log(`${claimed?.(1)}`);
```

```a
// Mutable invariance still applies to a callback slot.
interface Narrow { callback: (n: number) => number }
interface Wide { callback: ((n: number) => number) | undefined }
const narrow: Narrow = { callback: (n: number): number => n };
const wide: Wide = narrow;
wide.callback = undefined;
console.log(`${narrow.callback?.(1)}`);
```

```a
// Optional invocation does not break a captured strong cycle.
interface Holder { callback: (() => number) | undefined }
function run(): void {
	const holder: Holder = { callback: undefined };
	holder.callback = (): number => holder.callback === undefined ? 0 : 1;
}
run();
```

No proposal to accept an unbound method value is included. `obj.method?.()` is a
call-site receiver reference, while `const detached = obj.method` is a different
program with the existing unbound-method policy. Generic unsupported callable
signatures, unchecked callable assertions, incompatible mutable relations,
strong cycles and invalid checked views stay refused or NotYet by their existing
rules. This unit does not claim their root counts as retired.

Observed on the base: the three policy probes above fail respectively with
`adamic/no-unchecked-cast`, `adamic/invariant-mutable`, and
`adamic/cycle-capable`. Logs: `/tmp/scout-optional-policy.log` and
`/tmp/scout-optional-policy-3.log`. No policy exception was built.

A source-line classification of the 136 optional-call roots finds 34 lines with
an optional callable token, compared with 29 with optional `get`, 21 with
optional `has`, and 11 each with optional `find` or `forEach`. These are observed
line-shape counts, not an AST-level proof that all 34 share one supported ABI.
The selected compiler lesson is the callable guard, the largest single shape;
propagated receiver guards and unsupported signatures remain separate barriers.

## Source reductions and base outcomes

The nine [fixtures](step-18/fixtures/) are reductions of pinned TypeScript source,
with the original file and line in each header. `cross-call.a` is a derived
reassignment variant of the source's assignment/call-result chain, rather than a
literal unchanged reduction. `number.a` replaces SyntaxKind's enum tag with a
number but retains its required parent slot. The independent source execution
uses Node, not Adamic's JavaScript emitter. All finish with exit 0, no stderr.
The base compiler returns NotYet with no C emitted for every fixture; full
observations are retained in [baseline.json](step-18/baseline.json).

| Fixture | TypeScript witness | Node stdout (escaped) | Base exact NotYet reason |
|---|---|---|---|
| [call-result.a](step-18/fixtures/call-result.a) | moduleNameResolver.ts:3103, getPackageJsonInfo(...)?.contents.packageJsonContent. | `absent\nexports\n` | `an optional chain longer than one step` |
| [cross-call.a](step-18/fixtures/cross-call.a) | the assignment/call-result chain at moduleNameResolver.ts:3103. | `exports1\nexports1\n` | `an optional chain longer than one step` |
| [element.a](step-18/fixtures/element.a) | TypeScript checker.ts:7440, labeledElementDeclarations?.[i]. | `absent\nfirst\nabsent\n` | `?.[] on a value` |
| [function.a](step-18/fixtures/function.a) | TypeScript checker.ts:8066, cleanup?.(). | `cleanup\n` | `a call through ?. (an optional call)` |
| [method.a](step-18/fixtures/method.a) | TypeScript builder.ts:528, chain.repopulateInfo?.(). | `absent\nfilled1\n` | `a call through ?. (an optional call)` |
| [number.a](step-18/fixtures/number.a) | utilities.ts:8102, switch (parent?.kind). | `read\n` | `?. to a number, which would be number \| undefined` |
| [receiver-call.a](step-18/fixtures/receiver-call.a) | builder.ts:1852, state.seenEmittedFiles?.get(affectedSourceFile.resolvedPath). | `-1\n2\n` | `a call through ?. (an optional call)` |
| [size.a](step-18/fixtures/size.a) | builder.ts:705, !state.affectedFilesPendingEmit?.size. | `true true false\n` | `optional chaining to .size on a value` |
| [two-guards.a](step-18/fixtures/two-guards.a) | performance.ts:189, system?.cpuProfilingEnabled?.(). | `false false true\n` | `a call through ?. (an optional call)` |

`TestStep18SourceBaselines` holds every source reduction to the recorded Node
output, including the gaps. `TestStep18RecordedGaps` requires the exact current
reason and source position, rather than letting a different failure mask the
optional-syntax barrier. The ordinary oracle registry records each as not yet;
when a shape lands, its registration changes to a full Node/native/JavaScript
comparison and it leaves the gap test, without rewriting the base observation.

Fixture-unit validation:

- `go test ./internal/oracle -run 'TestStep18|TestNativeAgreesWithNode/docs/step-18/fixtures' -count=1 -timeout 10m` passed, including a restored run after both mutants. Logs: `/tmp/scout-optional-fixtures.log`, `/tmp/scout-optional-fixtures-restored.log`.
- Changing the source's `cleanup` output to `cleanup-mutant` failed `TestStep18SourceBaselines/function.a`, exit 0 with different stdout. Log: `/tmp/scout-optional-baseline-mutant.log`.
- Adding a valid array-of-arrays concat before the optional call failed `TestStep18RecordedGaps/function.a`: the exact reason/position check rejected the masking concat NotYet. Log: `/tmp/scout-optional-gap-mutant.log`.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts` passed after `npm ci --prefix stage3/api` installed pinned `@types/node` 25.3.3. The first run failed on missing Node type declarations, not on these fixtures. Logs: `/tmp/scout-optional-counts-3.log`, `/tmp/scout-optional-counts-3-retry.log`, `/tmp/scout-optional-api-setup.log`.
- No new fixture is counted yet because all nine are NotYet. The required generated refresh also moves the existing `logical_and_reference_maybe.a` row to registry order and removes the stale unregistered `17_binder_flow.a` row; measured allocation values do not change.

## Implemented closure guards

The implementation supports `f?.()` and ordinary closure fields such as
`obj.callback?.()`, including a callable selected by a factory call. It saves the
chosen closure in an owned local before arguments execute, checks both undefined
and null, and invokes only in the present arm. Both expression and discarded-call
paths use that lowering. Signature resolution strips nullishness; result fitting
adds the optional result representation after invocation. Existing argument-count,
default, rest and spread handling stays in the ordinary closure-call path.

Five fixtures now run through both backends: `function.a`, `method.a`,
`runtime-values.a`, `runtime-order.a`, and `runtime-arguments.a`. The filename
`method.a` comes from the source's property name; its value is an arrow closure,
not a receiver-bearing method. The new runtime fixtures are derived adversarial
variants of the TypeScript shapes named in their headers, not unchanged excerpts.
Their independent Node observations and current compiler outcomes are in
[runtime-observations.json](step-18/runtime-observations.json). Base observations
remain unchanged. The other nine fixtures retain exact NotYet checks.

The runtime cases cover absent and present closures, null, falsy returns,
number/boolean/reference/closure results, skipped arguments and spreads,
selection through a factory, binding and field reassignment during arguments,
stale checker narrowing, absent array elements, defaults, rest and argument-count
metadata. Dynamic strings and captured closures exercise ownership and release.
`runtime-methods.a` demands preservation of `this` while an argument replaces the
receiver. `runtime-cross-chain.a` crosses a call that clears the guarded binding;
its absent continuation must skip that call's argument. Both remain NotYet.

Observed during implementation: treating a class method property as a saved
plain closure produced failures in both backends. Existing `Property.Method`
metadata serves invocation, not a first-class bound callable. Consequently,
receiver-bearing methods are explicitly NotYet, including literal/class methods
hidden behind structurally compatible callback views. This corrects the initial
method-thunk assumption in the design. A future representation must own both the
selected callable and receiver across argument evaluation; re-reading either is
insufficient. This is an implementation boundary, not a new refusal policy.

Other boundaries remain NotYet: prior optional-chain guards, computed callees,
library calls without their intrinsic implementation, multiple callable
signatures, erased never-rest markers, observed void returns, and observed results
whose checker type has narrowed away undefined. The latter includes references:
a nullish runtime callable must not produce undefined through a required-result
representation. No Refused policy exception was implemented. General chains must
carry a separate short-circuit state: an undefined intermediate on a present
path must not accidentally acquire the absent path's permission to skip access.

### Mutation evidence

Each mutation below was run separately and restored before the final checks.
Compiler mutations reached the runtime comparison or the exact boundary check;
none was credited merely for failing C compilation.

| Mutation | Check that failed | Evidence |
|---|---|---|
| Source prints `cleanup-mutant` | Source baseline, stdout mismatch | `/tmp/scout-optional-baseline-mutant.log` |
| Earlier concat masks the optional-call gap | Exact gap reason/position | `/tmp/scout-optional-gap-mutant.log` |
| Re-select callable instead of reading its snapshot | runtime-order, stdout mismatch in both backends | `/tmp/scout-optional-mutant-reselect.log` |
| Omit null guard | runtime-values, JavaScript backend exit 70 versus Node 0 | `/tmp/scout-optional-mutant-null.log` |
| Evaluate arguments before the guard | runtime-arguments, stdout mismatch in both backends | `/tmp/scout-optional-mutant-eager.log` |
| Replace present numeric result with zero | runtime-values, stdout mismatch in both backends | `/tmp/scout-optional-mutant-result.log` |
| Remove optional-call presence-read handling | runtime-order, both backends panic on stale narrowing, exit 70 versus Node 0 | `/tmp/scout-optional-mutant-narrowing.log` |
| Remove receiver-method barrier | Exact literal-method and erased-prototype-view boundaries | `/tmp/scout-optional-mutant-receiver.log` |
| Use optional expression return descriptor | **Survived** runtime-values; not credited as killed | `/tmp/scout-optional-mutant-return-abi.log` |

An initial narrowing-mutant invocation matched no tests and was discarded; the
recorded invocation used the complete registered fixture path and failed both
backends. Eight mutations were caught, including six compiler mutations; the
return-descriptor mutation survived and limits the ABI claim.

### Validation and environment

Scoped commands, with all test output redirected to the named logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep18|TestNativeAgreesWithNode/docs/step-18/fixtures|TestNativeAgreesWithNode/internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a' -count=1 -timeout 10m > /tmp/scout-optional-final-oracle.log 2>&1
go test ./internal/lower -run 'TestStep18OptionalCallableBoundaries|TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestAMethodReadAsAValueIsRefused|TestAMutableLocationSeenWiderIsRefused' -count=1 > /tmp/scout-optional-final-lower.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/scout-optional-counts-4.log 2>&1
go test ./internal/lower -run 'TestOptionalFunctionValueRelation|TestOverloadedShorthandFunctionValueStaysNotYet' -count=1 > /tmp/scout-optional-final-relations.log 2>&1
```

The uncached oracle passed (2.066s), exercising Node, native and JavaScript
comparisons, sanitizer/leak runs, release checks and allocation checks for the
supported rows. Scoped lowering passed (3.561s); optional relation and overload regressions
passed (0.338s). The required counts refresh passed (49.141s),
adding five rows with balanced allocations and frees and zero leaks. No full
package suite or full gate was run. Backend emitters and the protected files were not edited.

Setup succeeded with the required GOPROXY setting. Reported timing lines: Go
ready 0.070s; Node ready 0.073s; clang ready 0.560s; markdown install step 1.044s,
ready 1.246s; submodules ready 17.781s; build ready 199.695s; test binaries
deferred 199.796s; build cache warm 199.797s; done 199.824s. `nproc` was 5,
with cgroup `cpu.max=400000 100000`. Tool versions were Go 1.27.1, Node 24.19.0,
and clang 20.1.8. Setup omitted the pinned Node declarations; the documented
`npm ci --prefix stage3/api` workaround installed them without changing its
package manifests. No cohere code was copied.

## Census retirement on the candidate

The final full census exited 0 with the no-output guard enabled: 79 compiler
files plus the metadata record, identical file coverage and checker metadata to
the base. Full before/after hashes, normalized remaining roots and every
changed position are in [retired-roots.json](step-18/retired-roots.json).

| Exact base reason | Before | After | Disappeared exact roots |
|---|---:|---:|---:|
| `a call through ?. (an optional call)` | 136 | 103 | 33 |
| `an optional chain longer than one step` | 11 | 11 | 0 |
| `?. to a number, which would be number \| undefined` | 3 | 3 | 0 |
| `optional chaining to .size on a value` | 20 | 20 | 0 |
| `?.[] on a value` | 6 | 6 | 0 |

Conservative retirement credit is **27 optional-call roots with no replacement
barrier on the same source line**. Five more advance to an ordinary read or
spread barrier; they are progress through the optional-call lesson, not accepted
programs. One is a reclassification to the new receiver-method NotYet and gets
no retirement credit. Thus 32 callable-guard barriers disappear, while the total
of step-18 syntax barriers falls from 176 to 144 after counting the new method
barrier. None of these counts proves that its whole declaration or file lowers.
No successful hidden-byte retirement is claimed from the historical ranking.

The six positions requiring that distinction are:

- `checker.ts:6910:21`: `an optional method call requiring a saved receiver`.
- `emitter.ts:652:9`: exposes `reading commonSourceDirectory` at column 40.
- `expressionToTypeNode.ts:214:13`: exposes `reading onExitNewScope`.
- `resolutionCache.ts:1094:9`: exposes `reading result` at column 121.
- `tsbuildPublic.ts:1144:23`: exposes `reading program`.
- `tsbuildPublic.ts:2292:5`: exposes `a call spreading elements that cannot be packed` at column 81.

Examples of the 27 credited local roots: `checker.ts:8066` (cleanup),
`emitter.ts:647` (emit notification), `resolutionCache.ts:506` (compiler-host
callback). The artifact lists all positions rather than extrapolating from these
witnesses. These observations belong to a checker-rejected compiler-directory
measurement, not a tsc CLI bootstrap claim.

Final census reproduction uses the restored production compiler:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/scout-optional-final-overlay > /tmp/scout-optional-final-overlay.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/scout-optional-final-overlay/overlay.json -o /workspace/scratch/scout-optional-final-census ./stage3/census/latent/tool > /tmp/scout-optional-final-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/scout-optional-final-census /workspace/scratch/scout-optional-adapted/src/compiler /workspace/scratch/scout-optional-final.jsonl > /tmp/scout-optional-final-census.log 2>&1
```

## Follow-up for step 18 (#vx8qdwg): return descriptor

Rebased the four Scout commits onto `origin/compiler/area-next-fixtures`
`4885cec5`. Delivery continues on `codex/scout-optional-calls-next`: publishing
the rebased commits under a new owned branch preserves the no-force-push rule.

[runtime-return-descriptor.a](step-18/fixtures/runtime-return-descriptor.a)
reduces checker.ts:8066's optional cleanup call through a void callback view.
TypeScript permits a value-returning function where its result is discarded.
The selected function returns 41 and 42, while the view requires both results
to be discarded. Node prints `2\n`, exit 0, and both backends agree, with native
sanitizer, release and leak checks passing. No language policy changed.

The exact earlier mutant replaces the signature return type with the optional
expression type in `callClosure`. It interprets a discarded numeric result as
an owned reference. This fixture catches it at runtime: ASan reports a SEGV in
`adamic_release`, native exit 1 versus Node exit 0; the release executable also
fails. JavaScript remains correct. This is a runtime failure, not a C warning.
Thus the earlier survival was a missing void-view adversary, not evidence that
the return descriptor is immaterial in general. Scalar packed widening still
shares its physical slot, which explains the original fixture's blind spot.

Commands and observed results:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep18SourceBaselines/runtime-return-descriptor.a|TestNativeAgreesWithNode/docs/step-18/fixtures/runtime-return-descriptor.a' -count=1 > /tmp/scout-18-descriptor-good.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/docs/step-18/fixtures/runtime-return-descriptor.a' -count=1 > /tmp/scout-18-descriptor-mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep18|TestNativeAgreesWithNode/docs/step-18/fixtures' -count=1 > /tmp/scout-18-descriptor-restored.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/scout-18-descriptor-counts.log 2>&1
```

Good run passed (0.675s); mutant failed (0.511s); restored scoped fixtures passed
(1.827s); required counts refresh passed (43.145s). This fixture-only push retires
**zero additional census roots**; compiler lowering is unchanged. Setup succeeded,
with `nproc=5`, submodules ready 0.100s, clang ready 0.258s, build ready 44.692s,
test binaries deferred 44.816s, cache warm 44.817s, and done 44.847s.

## Follow-up: receiver-bearing optional methods

The saved-receiver barrier is now implemented for represented user methods,
including class prototype entries, literal methods and callback fields behind
readonly structural views. The earlier receiver deferral is superseded by this
unit. `CallClosure.Optional` keeps its original `Returns` ABI separately from
`OptionalResult`. Native selects the receiver and its own callback or prototype
entry before evaluating arguments, retains the selected callback, and evaluates
arguments only when the selected callable exists. An own undefined callback
shadows prototype lookup. JavaScript snapshots the same object and function
before arguments, then invokes a receiver-bearing function with that object.
Missing optional methods short-circuit. Defaults, omitted arguments, rest
arguments and argument counts retain the existing call ABI.

The method fixtures reduce checker.ts:6910's report callback and existing tsc
optional-call argument shapes. `runtime-bound-methods.a` alternates absent,
class and literal methods. `runtime-bound-method-arguments.a` checks skipped
spreads, defaults and `arguments.length`. `runtime-bound-method-selection.a`
clears an own callback during argument evaluation and then calls a class method
through the same readonly view. The existing `runtime-methods.a` replaces the
receiver during an argument and must invoke the original receiver.

`runtime-discarded-reference.a` exposed an additional leak: a void view may hide
an owned reference return. Native now identifies the actual selected closure
code or method thunk and releases its discarded reference result, while keeping
numeric void-view results out of the release path. This uses the program's
represented function set rather than treating every discarded packed slot as a
pointer. Allocation and release counts are balanced in the required counts refresh;
the counts table records each fixture separately.

Language refusals are unchanged. `runtime-bound-method-replacement.a` contains
`original.report = (value: string): string => ...` for a declared literal method.
It remains Refused by the existing unbound-method rule; changing that is a
proposal for @system_adamic, not part of this implementation. Checked or
uninitialized field contracts without a represented invocation path remain
NotYet. Detached methods, overloads, erased never-rest signatures, observed void
results and optional results narrowed to present retain their earlier barriers.

Six mutants were run and restored; each failed after clean C compilation:

| Mutant | Fixture | Catcher |
|---|---|---|
| Pass NULL instead of the selected receiver | runtime-methods | UBSan null member access, native exit 1 versus Node 0 |
| Ignore prototype entry presence | runtime-bound-methods | Native output differs from Node |
| Reselect the callback after arguments | runtime-bound-method-selection | JavaScript TypeError, exit 70 versus Node 0 |
| Evaluate spreads before the presence guard | runtime-bound-method-arguments | Native spread count and output differ |
| Omit the selected callback retain | runtime-bound-method-selection | ASan heap use after free |
| Drop discarded reference cleanup | runtime-discarded-reference | LeakSanitizer: 209 bytes in three allocations |

Validation commands (all output redirected to the named logs):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestStep18|TestNativeAgreesWithNode/docs/step-18/fixtures' -count=1 > /tmp/scout-18-methods-final-oracle.log 2>&1
go test ./internal/flow -run TestStep18OptionalMethodPaths -count=1 > /tmp/scout-18-methods-flow.log 2>&1
go test ./internal/lower -run 'TestStep18OptionalCallableBoundaries|TestIteratorViewsCannotEraseReceivers|TestLiteralMethodViewsDoNotLoseThis|TestRepresentedMethodReplacementIsNotYet|TestDestructuredMethodsCannotLoadOwnSlots|TestOptionalFunctionValueRelation|TestAMethodReadAsAValueIsRefused' -count=1 > /tmp/scout-18-methods-lower-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/scout-18-methods-counts.log 2>&1
```

Oracle passed (4.464s), flow passed (0.308s), lower boundaries passed, counts
refresh passed (53.050s). Every supported fixture is held to Node in both
backends, native sanitizer and release builds, and the leak check. Mutant logs
are `/tmp/scout-18-mutant-method-{binding,presence,selection,eager,retain,discard}.log`.

The rebased before/after full census each contains 79 compiler files and identical
checker-rejection metadata, and both no-output guards exit 0. Normalized exact
roots are unchanged by the rebase: optional calls 103, longer chains 11, size 20,
indexing 6, required numeric reads 3, saved-receiver methods 1. This method unit
retires **one additional root**, `checker.ts:6910:21`, with no replacement barrier
on that source line. It removes the saved-receiver group (1 to 0); the remaining
103 optional-call roots are unchanged. See [receiver-roots.json](step-18/receiver-roots.json)
for hashes and the exact comparison. No whole-file acceptance or hidden-byte
retirement is claimed.

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/scout-18-method-overlay > /tmp/scout-18-method-overlay.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/scout-18-method-overlay/overlay.json -o /workspace/scratch/scout-18-method-census ./stage3/census/latent/tool > /tmp/scout-18-method-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/scout-18-method-census /workspace/scratch/scout-optional-adapted/src/compiler /workspace/scratch/scout-18-method.jsonl > /tmp/scout-18-method-census.log 2>&1
```
