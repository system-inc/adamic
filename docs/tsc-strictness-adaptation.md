# Meter optional-property adaptation

`cmd/adamic-meter --adapt` now follows the type-import adaptation with an
in-memory optional-property declaration adaptation. It resolves declarations
through the checker, applies edits to the existing source overlay, and rechecks
the whole program until no further eligible declaration is found. Source files
on disk are unchanged. No language options or lowering rules were changed. The subsequent
[return adaptation and next-layer survey](tsc-strictness-next-layer.md) has a
separate report.

```ts
// Before
interface Slot { value?: string; callback?: () => string; }
// After
interface Slot {
    value?: (string) | undefined;
    callback?: (() => string) | undefined;
}
```

The parentheses matter: adding undefined to a callback's return type would not
represent a possibly undefined callback. Optional interface method signatures
are converted to optional callable properties, preserving their parameters,
type parameters, and return type. Live class methods are excluded: converting
one to a field changes prototype placement and potentially its `this` behavior.

The rule is driven by TS2412/TS2375/TS2379 diagnostics and the actual declaration
symbols. It handles direct property writes, contextual literals, returns,
arguments, and structural views. It requires an explicitly typed optional
property in a named root file, an undefined source constituent, and compatibility
of every other source constituent with the declared value type. Required slots,
external declarations, unresolved/unknown types, inferred slots, arbitrary
generic receiver contracts, live methods/accessors, and overloaded method
signatures are excluded. A bare indexed read is not a seed; an existing explicit
missing-value branch may be, as in `items.length ? items[0] : undefined`.
Shared declarations can remove other diagnostics incidentally; that does not
establish the safety of an indexed read or another caller invariant.

Edits preserve `?`, assignments, object keys, and runtime expressions. Nested
declarations compose using insertions at original byte offsets; diagnostic
columns are interpreted as UTF-16. The JSON adaptation entry records both
`declarations_changed` and the net reduction in the three exact-optional codes.
The text report counts removed diagnostics and entries reaching lowering.

The resolver builds its own tooling-only checker program to inspect rejected
source; it never passes that program to lowering. Its options match load's
fixed options. `LoadOverlay` remains the authoritative checker gate, including
the Adamic prelude, before an entry can reach lowering.

## TypeScript 6.0.3 measurement

Same input as the [survey](tsc-strictness.md):
`050880ce59e30b356b686bd3144efe24f875ebc8`, all 77 roots under `src/compiler`.
[Complete final meter report](tsc-strictness/meter-adapted.json).

| Measure | Import adaptation only | Both adaptations |
| --- | ---: | ---: |
| Checker diagnostics | 2,744 | 2,070 |
| TS2412 | 617 | 6 |
| TS2375 | 76 | 11 |
| TS2379 | 31 | 14 |
| Exact-optional total | 724 | 31 |
| Entries reaching lowering | 0 | 0 |

The new adaptation changes **415 declarations** and removes **693 exact-optional
diagnostics**. The existing import rule still removes **3,718 TS1484** findings.
The total checker reduction is **674**, rather than 693, because consumers are
rechecked and other diagnostic counts change:

| Other changed code | Before | After |
| --- | ---: | ---: |
| TS2345 | 688 | 693 |
| TS2322 | 103 | 102 |
| TS2366 | 1 | 0 |
| TS2430 | 0 | 10 |
| TS2320 | 0 | 6 |

The new interface diagnostics expose relationships that need consistent
representation across declarations. For example, `SignatureDeclarationBase`
now allows present undefined in `name`, while `NamedDeclaration` does not;
`PerDirectoryResolutionCache` and `NonRelativeNameResolutionCache` disagree about
`isReadonly`. The adapter reports these failures rather than widening unrelated
base declarations without write evidence or lowering a rejected program.

Leading remaining reasons:

| Reason | Count |
| --- | ---: |
| TS2345: incompatible argument | 693 |
| TS18048: possibly undefined value | 354 |
| TS7030: missing return on some paths | 251 |
| TS2532: possibly undefined object | 224 |
| TS1294: syntax incompatible with erasableSyntaxOnly | 180 |
| TS2322: incompatible assignment | 102 |
| TS7029: switch fallthrough | 83 |
| TS2591: missing Node globals/types | 54 |

**No entry reaches lowering.** Entries without diagnostics in their own files
are also tried independently with the final overlay; their dependency closures
still fail the checker. There is consequently no first Refused or NotYet result
for this corpus. A separate existing meter fixture proves reachability reporting
with two entries: one first Refused and one first NotYet.

Of the original 96 M sample candidates, **95 no longer have their sampled
exact-optional diagnostic**. E087 (`resolutionCache.ts:1386`, clearing
`resolution.files`) is left because its receiver is a type parameter. An
arbitrary specialization may narrow that property. This is the additional
owner/specialization review required by the original survey, not evidence that
all 96 could safely be automated from their source lines alone. Remaining
exact-optional errors also include generic contracts, nested structural
incompatibilities, and positions requiring stronger evidence. No checked unwrap
or missing-value default was introduced.

## Validation and mutants

Commands, with every test run redirected to a log:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-meter --adapt --json /tmp/adamic-tsc-strictness/typescript/src/compiler > /tmp/adamic-tsc-strictness/meter-optional-final.json 2> /tmp/adamic-tsc-strictness/meter-optional-final.stderr
go test ./cmd/adamic-meter ./internal/load -count=1 > /tmp/adamic-tsc-strictness/optional-tests-final.log 2>&1
go vet ./... > /tmp/adamic-tsc-strictness/optional-vet-final.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/adamic-tsc-strictness/optional-gate-final.log 2>&1
```

The meter run exits successfully; its checker-code counts sum to 2,070.
The diagnostic exporter independently produces 2,070 diagnostics and the same
415/693 adaptation counts. All touched-package tests, vet, and the complete
worker gate pass, including the native/Node oracle. This is the ordinary worker
gate with oracle observation caching permitted, not the uncached integration gate.

Fixtures cover runtime key presence and enumeration against Node, callable union
precedence, structural union views, forwarded options, generic and wrong-value
rejection, existing missing-value branches, nested edits, ownership boundaries,
UTF-16 columns, `.a` files, idempotence, source-file preservation, and the newly
exposed error after an `in` guard. The Node oracle uses Node 24's TypeScript
stripping to execute original and adapted source; no npm install is required by
these package tests.

The `.a` fixture exposed an existing `LoadOverlay` bug: alias reads bypassed the
overlay and read disk. The small `internal/load/source_fs.go` fix honors both the
synthetic `.a.ts` name and the caller's `.a` name before reading disk. A loader
regression proves the overlay is checked and the disk source remains unchanged.

Actual implementation mutants were run through Go file overlays, without
changing the working tree:

| Mutant | Fixture and observed failure |
| --- | --- |
| Replace an undefined write with `delete slot.p` | `TestAdaptOptionalPropertiesPreservesRuntimeAndDisk`: adapted Node output changes from `true true p,callback` to `false false callback`; exit 1 |
| Apply a callable-property rewrite to a live prototype method, replacing it with an initialized instance field | `TestOptionalAdaptationPreservesLiveMethodPlacement`: Node prints `true false` instead of `false true`; exit 1 |
| Restore disk-first `.a` alias reads | `TestOverlayPrecedesDiskForAdamicAliases/.a`: checker reads the disk's wrong string value and reports TS2322 instead of accepting the numeric overlay; exit 1 |

Both meaning-changing adapter mutants typecheck; they fail the Node runtime
comparison, rather than a parser, compiler, or clang warning. The loader mutant
fails its dedicated source-view assertion.

Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 16s,
total 16s. `nproc` reports 5, with a cgroup quota of 4 CPUs.

The original survey's exporter now defaults explicitly to import-only adaptation
so its 200-row audit remains reproducible. Set `ADAMIC_SURVEY_OPTIONAL=1` to export
the new result using the same overlay instructions. The historical survey and
sample counts are retained rather than replaced by this follow-up measurement.
