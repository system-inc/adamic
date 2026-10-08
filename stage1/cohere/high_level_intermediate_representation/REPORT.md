# Unit 2 report — full construction and static-components; clone native stop

Roadmap step 28. This stopped unit is on `stage1-hir/wip`; the watched plan branch
is not pushed. The lint-area merge remains `c7dab335` from `ad7bd066`; cohere stays
pinned at `7945d102a6c18dd36adf9114a758ce646e8b2359`. HIR made no parser change.

## Coverage

| Certificate | Native | Node |
| --- | --- | --- |
| Original construction functions | **1,442 / 1,465** | **1,442 / 1,465** |
| Admitted non-Flow originals | **1,442 / 1,442** | **1,442 / 1,442** |
| Construction path probes | **72 / 72** | **72 / 72** |
| Combined constructed observations | **1,514 / 1,537** | **1,514 / 1,537** |
| Static-components upstream cases | **22 / 22** | **22 / 22** |
| Static-components owned witnesses | **3 / 3** | **3 / 3** |
| Pending CloneFunction observations | **0 — compiler refusal** | **1,514 / 1,537** |

Every admitted graph is byte-identical to Go hir-v1, through direct construction
and cached ForFunction. Static-components compares complete findings to the
unchanged Go rule; emitted JavaScript matches too. The 23 Flow graphs and 40 Flow
fixtures remain catalogued exclusions. All 45 original Go test skips retain
their per-group out-of-rule-reach / later-unit classification in SKIPPED.md and
the checked census summary. No admitted failure was dropped.

Construction now covers methods/accessors and computed names, closures and
capture writes, all statement/control-flow/JSX paths in the admitted corpus,
optional chains, typed/default/rest parameters, patterns and destructuring,
catch patterns, casts and unsupported targets, debugger/meta/unsupported nodes.
Patterns have their own checked arena; no cyclic object links are introduced.
Go's Lower→Clone→Construct versus ForFunction→Clone invocation distinction is
recorded independently as cloneBeforeSSA, preserving both pattern views.

## Mutants and local validation

Static-components' creation mutant **fails against Go on native and Node**.
The HIR semantic matrix has **79 / 79 mutants caught** on native and Node,
combining the full run with the repaired-anchor complete-corpus reruns. FunctionIndex and PatternIndex off-by-one mutants stop at checked
reads on both runtimes. Private minting and cross-brand misuse are rejected;
foreign arena handles also fail. The ForFunction cache-hit mutant is caught on
native and Node. Pending CloneFunction's alias mutant fails on Node at
`BlockIndex belongs to another arena`; native clone is not certified.

The rule certificate is reproducible with its owned `certify.sh`, which overlays
its owned test beside the generic lint harness. It captures every upstream case
for the exact descriptor name and uses the harness's native checker transcript
for Node/emitted-JS replay, including the mutant. No shared registration list or
parser code was patched. Two old object-mutant anchors moved with object method
lowering; their targeted complete-corpus rerun passes after reanchoring.

## Rulings and new language stop

The 09:44 optional-boolean ruling (a) is applied: an optional record holds required
boolean presence/value fields. Every site is marked `gap 1 (GAPS.md)`. The original
proving program and exact selector-style test stay in GAPS.md / gaps_test.go,
failing when step 17 (#qgr7mm4) lowers the field. Go dumps identify **74 / 1,465**
original graphs requiring that field and **1,391** without it; the retained impact
JSON names every key and original test call. The authorized workaround restores
the whole native construction build.

Unit 2 stops on the **isolated native CloneFunction** gap, GAPS.md gap 2.
Go `clone.go:96` copies `Instruction.Value: InstructionValue`, through the deep
copy at `inline_remap.go:271`. Adamic `clone.ts` copies `ValueType` records and
owned arrays into fresh checked arenas. The compiler refuses object spread in
the same program as the required private-index static constructor objects:

```
adamic: clone_main.ts: stage 0 can't lower spreading in a program with static constructor objects yet
```

The three-line proving program is `testdata/static-constructor-spread-gap.a`;
Node prints `0\n`, and its exact `lower.NotYet.What` is
`spreading in a program with static constructor objects`. The gap test passes on
that refusal and fails as soon as it closes. No workaround is applied.
The pending clone is isolated behind clone_main.ts / clone_coverage.ts, with no
import from ordinary construction or the rule. Its Node count is separate from
the native = Node construction certificate. Unit 2 is not declared complete.

## Allocations

The same byte-identical **12-function** native probe:

| Metric | Current |
| --- | ---: |
| Allocations / frees | **4,161 / 4,161** |
| Delta from original 3,212 | **+949** |
| Delta from shared-index 3,869 | **+292** |
| Delta from typed checkpoint 4,124 | **+37** |
| Retains / releases | 16,250 / 14,835 |
| Peak live / regions | 237 / 0 |

Known causes of earlier growth are boxed concrete indices, canonical handle
arrays and owner-bound SSA callbacks. This slice additionally creates pattern
node and handle arrays per function. No isolated share of the +37 is claimed;
the brand remains enforced. Full current admitted corpus: **4,300,331 allocations
and frees**, retains 19,518,706, releases 18,665,838, peak 11,969, regions 0.
Its denominator differs from older full-corpus measurements, so that total is
not presented as a like-for-like delta.

## Next work and parallel lanes

After the ruling/fix, certify CloneFunction on native and rerun its integrated
certificate. Static-components itself is green. PLAN.md now names, for units
3–10, exact Go entry points, input/output checkpoint states, exclusive owned
directories and shared-file ownership. Units 3, 4, 6–10 can start isolated Go-input
work without earlier Adamic algorithms. Unit 5 and unit 4's projection require
the promised mutation_aliasing area import, absent in this checkout. A shared
replay owner must first provide the decoder and missing pass-state sidecars:
today's printer has `scopes -`, not a complete later-pass replay schema.
Only the composed unit 10 pipeline genuinely needs the earlier Adamic passes.

Retained certificates and the aggregate mutant receipt are in `validation/unit2-step28/`.
Final metadata verification also preserves source handles on conditional/logical/optional
temporaries, needed for exact static-components creation spans; hir-v1 is unchanged.
