# `@typescript-eslint/prefer-optional-chain`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **238** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Enforce using concise optional chain expressions instead of chained logical ands, negated logical ors, or empty objects

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

238 in the tree. Showing the first few.

**`app/(os-layout)/_components/TasksCenter.tsx:876`**

```
lastCommentWakeOutcome && lastCommentWakeOutcome.taskId === selectedTaskId
```

> Prefer using an optional chain expression instead, as it's more concise and easier to read

**`app/(os-layout)/_components/kingdom/OsKingdomView.tsx:245`**

```
selectedMember && selectedMember.parent && selectedMemberLookup
```

> Prefer using an optional chain expression instead, as it's more concise and easier to read

**`app/(os-layout)/_components/kingdom/OsKingdomView.tsx:256`**

```
if(!snapshot || !snapshot.team) {
```

> Prefer using an optional chain expression instead, as it's more concise and easier to read

**`app/(os-layout)/_components/profile/OsProfileDateRangeResolve.ts:35`**

```
if(!match || match.days === null) return null;
```

> Prefer using an optional chain expression instead, as it's more concise and easier to read

**`app/(os-layout)/_components/select/TaskTitleFocusContext.tsx:61`**

```
if(!pending || pending.taskId !== taskId) return null;
```

> Prefer using an optional chain expression instead, as it's more concise and easier to read

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


## Sized and declined 2026-09-06, with the decomposition premise measured and rejected

This rule was dispatched twice: once whole, once as "port the judgment, decline the fixer" on the
premise that its 1,807-line utils package is mostly the repair. Both were stopped before any Go was
written, the second by measuring the premise rather than accepting it.

**The premise is false.** Partitioning `analyzeChain.ts` by function boundary rather than grepping
it, the fixer is `getReportDescriptor` at 323 lines of 1,807, which is 18 percent. Controlled:
`gatherLogicalOperands.ts` and `checkNullishAndReport.ts` carry zero fixer references. Declining the
fixer therefore leaves about 1,713 lines of judgment, past the ~600-line calibration before the
corpus is counted at all.

    prefer-optional-chain.ts                230
    gatherLogicalOperands.ts                492    8 nullish comparison kinds
    compareNodes.ts                         413    structural node comparison
    checkNullishAndReport.ts + options       52
    analyzeChain.ts, judgment portion       526
    judgment total                        1,713

**The corpus is the stronger argument and it inverts what the line counts suggest.** 739 cases
across 14 files, 18,173 lines, and 729 of the 739 carry an `output`. The rule is about 98 percent
repair-asserting, so the judgment half is not a smaller corpus: it is the same corpus with its
most-asserted field discarded, along with 605 `suggestions:` assertions. Porting the half whose
coverage you are throwing away is the worst available shape.

Generalised in the porting standard: read the `output` ratio before accepting any judgment/fixer
split. A corpus that is overwhelmingly output-asserting cannot be halved along that line however the
source happens to divide.

**A third cost, specific to this rule.** It uses `isTypeFlagSet` 14 times alongside
`unionConstituents` 8 times, deliberately decomposing unions and testing flags. Our shelf's
`checking.IsTypeFlagSet` is the own-flags reading, so every one of those 14 sites is the
same-name-different-reading trap and needs a local variant with its own measurement. On this rule
that is a substantial share of the work rather than a footnote.

**What is dispatchable.** `gatherLogicalOperands.ts` alone: 492 lines, 8 nullish comparison kinds,
zero fixer references, consumed by both `analyzeChain` and the rule surface. Ported with its own
probe corpus and no rule attached it is a real deliverable, and it makes the remaining ~1,200 lines
a single agent's job. That is a different task from either dispatch above and wants a fresh budget.
