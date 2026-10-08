# Hidden boundary 01: overload parameter

This unit fixes the supplied readonly witness's native argument-layout crash.
It does not remove the assigned census refusal. Roadmap step 30 remains blocked
at this head; step 05 measures zero revealed bytes for the largest interval.

Base: `origin/compiler/area-next-fixtures`,
`dcdbb9098f77f30ad41790c56df1bd63ad462b63`.
Delivery branch: `codex/hidden-01-overload-parameter`.
No other worker branch was merged. No main or area branch was pushed.

## Observation and soundness

The exact historical replay reproduces on the base and after this change:

```text
transformers/declarations.ts:619:54: Adamic 0.1 refuses overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem
```

The actual source declares:

```typescript
function visitBindingElement<T extends Node>(elem: T): T;
function visitBindingElement(elem: ArrayBindingElement): ArrayBindingElement {
    // The omitted-expression arm returns elem; the other arm updates a binding element.
}
```

The overload admits every Node subtype. The implementation only handles binding
elements, and its replacement result does not establish identity of an arbitrary
T subtype. The supplied readonly number witness has a compatible contract; it
is not a reduced reproducer of this universal generic promise. Removing the
parameter check would trust an incompatible implementation. Validated callback
specialization would need to prove the actual callback domain and its promised
result. That proof is not built here, so the historical refusal stays.

The readonly witness initially passed lowering, then native interpreted its
number field as a boxed number-or-string pointer and crashed in `adamic_retain`.
Node returned `7\n`, exit 0, empty stderr. The final focused oracle holds source
Node, generated JavaScript, ASan/UBSan native, release native and leak checks to
that observation.

Calls to nongeneric top-level overload implementations now fit fresh object
literal fields to the implementation's representation. Nested fresh literals
can be fitted recursively. This changes the allocation's fields rather than
copying an existing object, so it preserves object identity and evaluation order.
An existing object with different field storage is NotYet; copying it or
rewriting its slots would break identity or aliases. Nested and rest overloads
requiring different storage are also NotYet. Generic argument mapping remains
on its existing specialization path; this unit adds no generic escape proof or
callback specialization.

## Fixtures and checks

Positive: `internal/oracle/testdata/hidden_boundary_overload_parameter.a`.
Negative fixtures under `internal/oracle/refusals/`:

- `hidden_boundary_overload_parameter.a`: implementation narrowed to literal 7,
  while the overload admits every number. Node prints 8; the checker rejects
  the contract with TS2394.
- `hidden_boundary_overload_generic.a`: every Node subtype admitted but only
  BindingElement implemented. Node prints 0; lowering retains the exact named
  overload parameter Refused.

Lowering tests pin the existing-object and nested-layout NotYet stops, and rerun
the strict contravariance, generic constraint, mutable parameter and result tests.

Commands, with output redirected to the named scratch logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(hidden_boundary_overload_parameter|census_overload_contracts|census_append_overload).a$|TestHiddenBoundaryOverloadParameterRefused' -count=1 > /tmp/hidden01-oracle.log 2>&1
go test ./internal/lower -run 'TestHiddenBoundaryOverloadLayoutStops|TestCensusOverloadRelation' -count=1 > /tmp/hidden01-lower.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/hidden01-counts.log 2>&1
```

All pass. No whole-package tests or full gate were run.

Independent mutants, restored before the final green run:

| Mutant | Intended catcher and observation |
|---|---|
| Keep the narrow literal field instead of storing its converted value | Positive Node oracle: native sanitizer crash in adamic_retain, exit 1 versus Node exit 0. No compilation diagnostic killed it. |
| Narrow the implementation field to literal 7 | Positive oracle fails Load with TS2394. The separate negative fixtures remain rejected. |
| Bypass the existing-object layout stop | Alias lowering test fails because lowering returns nil error. |
| Bypass the nested-layout stop | Nested lowering test fails because lowering returns nil error. |

Logs: `/tmp/hidden01-mutant-layout.log`, `/tmp/hidden01-mutant-narrow.log`,
`/tmp/hidden01-mutant-alias.log`, `/tmp/hidden01-mutant-nested.log`.

Counts changes are exhaustive:

- New witness: allocations 4, frees 4, retains 1, releases 4, peak 4, regions 0.
- `logical_and_reference_maybe.a` moves to its registry order; all numbers stay
  8, 8, 11, 22, 4, 0.
- The stale `stage3/fixtures/taste/17_binder_flow.a` row is removed. Its registry
  already marks it stopped on the base, so counts do not run it. This unit does
  not change its status.

## Pinned region measurement

Adapted source pin: `388096e6a83a4e9d287fb827f793c599ba1bf0ad`.
The pin's `stage3/apply.sh` prepared `/tmp/hidden-adapted`; all 82 source files
match the pin's RESULT.json hashes. Assigned file SHA256:
`3b69b0e68abb0f4a0745d5dd45651f43c9c336fe5b72883fe8d4dde1f6c4d6ed`.

| Largest assigned interval | Before hidden intersection | After hidden intersection | Revealed |
|---|---:|---:|---:|
| declarations.ts [68180, 81805) | 13,625 | 13,625 | 0 |

Both censuses load and register the entire hash-pinned project. A scratch-only
filter in `latentFullSelected` attempts all 75 independent units of the assigned
file. Frozen `census_small.go` overlays select the base and fixed compiler.
The pin's `hidden.py.calculate` performs boundary union minus independently
examined spans against the pin's stock AST catalogue.

The enclosing Boundary [7221, 95509) covers the entire assigned interval in both
runs. Other files' skipped-dependency records can only add blocked spans there;
they cannot expose bytes. Independent examined spans for this interval come
from the assigned file's units. Thus attempting every unit in this file is
sufficient for this intersection. This is not a whole-corpus measurement.
Unused whole-corpus runs were stopped after the assigned-file runs completed.
The 20,524-byte group total is not reported as revealed bytes.

Remaining head, with no new downstream boundary exposed:
`transformers/declarations.ts:619:54`, the same overload parameter Refused.

The completed before and after file census artifacts are byte-identical, SHA256
`5a14c2ed9838843c7e4435e236bc127362f862e7fde533875a949ab21e5485e0`.
The measured fixed compiler file matches the delivery file, SHA256
`ceaa722781afc49214960f1ec46efc6036cf870145c3c82fb45104396e0348df`.

Artifacts: `/tmp/hidden01-{before,after}-file.jsonl`,
`/tmp/hidden01-{before,after}-file-hidden.json`,
`/tmp/hidden01-measurement.json`, `/tmp/hidden01-source-hashes.json`,
`/tmp/hidden01-replay-{before,after}.log`.
The scratch overlays and binaries are `/tmp/hidden01-{before,after}-file-overlay`
and `/tmp/hidden01-{before,after}-file-census`.

## Setup

GOPROXY was set to `https://proxy.golang.org|direct` before `bash cloud/setup.sh`.
Cumulative setup timings in seconds: submodules 0.102, Node 0.157, Go 0.176,
markdown dependencies 0.341, clang 0.614, Go build 49.773, deferred test binaries
50.000, warm cache 50.002, done 50.030. `nproc` is 5; CPU quota is
`400000 100000`. Source environment: `/workspace/adamic-tools/env.sh`.

Setup succeeded. The initial counts run exposed missing @types/node 25.3.3.
`npm ci --prefix stage3/api` installed the pinned dependency; counts then passed.
The initial witness crash and the initial missing-dependency counts failures
remain recorded in the work observations. No full tsc executable, callback-domain
proof, generic storage conversion, or full-corpus byte delta is claimed.
