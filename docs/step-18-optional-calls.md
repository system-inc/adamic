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
| NotYet: `?. to a number, which would be number | undefined` | 3 | 1964 | 3 | `checker.ts:10416`; `checker.ts:10445`; `utilities.ts:8102` |
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
| `obj.method?.(args)` | Evaluate the receiver and select the callable once. A method reference remains bound to its receiver through the existing method thunk. Save that callable before evaluating arguments, then use the same conditional as a free callable. Optionality is on the callable, not on `obj`. | Library callable values retain their own lowering contracts. Earlier optional-chain propagation needs the general chain design below. |
| `a?.[i]` | Save `a`, guard it, and evaluate `i` only in the present arm. Keep the existing array, string or tuple access semantics, including out-of-range undefined; widen the result after access. | Arrays and strings with optional access remain NotYet in this unit. Tuple support already present is not credited to this work. |
| `get()?.field` and `get()?.field.more` | Save the call result and guard that value. A continuous chain carries an explicit short-circuit exit through all following ordinary accesses and calls. Parentheses end a continuous chain: `(get()?.field).more` must still fail on an absent field. | General multi-step propagation remains NotYet. Existing one-step reads retain their current coverage. |
| `obj?.method?.(args)` | Separate guards for receiver absence and selected callable absence, with one shared undefined exit. Each receiver, selector and argument has its source evaluation position. | Two guards and propagated optional flags remain NotYet here. |
| Chains crossing a call that reassigns a binding | Saved values are owned snapshots; later operations read the source binding anew only where JavaScript does. A saved receiver or function survives a reassignment during an argument. Checker narrowing across calls is not proof of runtime presence. | General chain lowering must retain runtime defined/readiness checks for unguarded later accesses; it must never reuse a stale narrowing or stale alias as the chain value. |

The callable's signature return type and the optional expression's type are two
different types. For example `(() => number) | undefined` is a nullable closure,
its call returns a plain number, and `f?.()` returns a maybe-number. Passing the
maybe-number ABI to a function that returns a number silently corrupts the result.
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
