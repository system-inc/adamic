# `init-declarations`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **476** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require or disallow initialization in variable declarations

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

476 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:109`**

```
let lowerBound: number;
```

> Variable 'lowerBound' should be initialized on declaration

**`app/(os-layout)/_components/TasksCenter.tsx:110`**

```
let upperBound: number | null;
```

> Variable 'upperBound' should be initialized on declaration

**`app/(os-layout)/_components/drag/TasksDragContext.tsx:376`**

```
let reparentToTaskId: string | undefined;
```

> Variable 'reparentToTaskId' should be initialized on declaration

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:114`**

```
let firstChildX: number | undefined;
```

> Variable 'firstChildX' should be initialized on declaration

**`app/(os-layout)/_components/kingdom/OsKingdomLayout.ts:125`**

```
let x: number;
```

> Variable 'x' should be initialized on declaration

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


## Ported? No, and the reason is measured

**Declined 2026-09-06.** Not ported as a bare core rule, and the line item should be read as
already satisfied rather than outstanding.

### The algorithm is already in the tree, under the namespaced name

`@typescript-eslint/init-declarations` is registered, enabled at `CohereSettings.json:468`, and
lives at `internal/lint/rules/typescript/init_declarations.go` (376 lines). It carries the whole
algorithm inline rather than delegating, because a rule package may not import another rule
package and there is nowhere else it could live.

Upstream's extension is a **pure wrapper**. Its source opens with

    const baseRule = getESLintCoreRule('init-declarations');

and its `create` is `baseRule.create(getBaseContextOverride())` plus two TypeScript-specific
adjustments: a report-LOCATION override so a type annotation is not underlined, and a skip for
`declare` declarations and declared namespaces. It cannot report anything on its own; every
finding comes from the core rule underneath. That is the standard's case 1 in
"A core/extension pair is often one job", and a bare port would be a second implementation of
code already present.

### Measured three ways, because the structural argument alone is not enough

`no-invalid-this` was also a documented wrapper case and porting it anyway turned out to be
correct, because the two disagreed on six shapes. So this was measured rather than assumed.

**The two upstream rules agree on core's whole corpus.** Both driven through the ESLint Linter
interface under `sourceType: module` with the typescript-eslint parser, over all 81 cases
extracted from core's own corpus:

    81/81 agree. core reported 31 findings, the extension 31.

The harness refuses to print a result if either side reports zero across the corpus, since two
silent rules agree vacuously.

**And OUR shipped port answers core's corpus identically.** Driven over the same 81 cases:

    26 cases report, 55 clean, 31 findings total

The same 31. Our namespaced port already is the core algorithm on core's own inputs, so a bare
`init-declarations` would add no verdict this tree does not already produce.

    go test -count=1 -run TestShippedPortAnswersCoresCorpus ./internal/init_declarations_probe/

### The corpus makes two RuleTester.run calls, which is worth recording

The extraction captured 81 cases from **2** `run()` calls. A single-slot interceptor keeps only
the last and silently loses most of the corpus, which the standard names as a real defect that
has shipped before. Any future work here should confirm both calls are captured.

### What would change this

If the extension's report-location override or its `declare` skip were ever found to suppress a
finding core produces, the core rule would become genuinely absent rather than redundant. Neither
does on this corpus: the override moves a span and the skip only removes `declare`, which core has
no concept of.
