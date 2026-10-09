# `@typescript-eslint/no-confusing-void-expression`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **226** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Require expressions of type void to appear in statement position

## Where cohere is deliberately quieter than ESLint

Upstream asks whether an expression's type has `TypeFlags.VoidLike`, which is void OR undefined, so an
honestly `undefined` value in a non-statement position reports as if it were nothing. cohere asks for
`TypeFlags.Void` alone. `void` says there is no result and nothing should look at it; `undefined` is a
value somebody returned on purpose, and the rule's name and every message are about void. The
fixers' `canFix` guards keep VoidLike, since they are upstream's rewrite-safety test and run only
after a report.

Upstream's corpus writes no undefined-typed expression, so none of its 110 imported cases moves. The
real sites that went silent: `libraries/structure/libraries/nexus/source/security/random/Random.test.ts:6`
(`const result = arrayGetRandom([])`, `T` infers to `never`, the call types as `undefined`) and
`libraries/structure/libraries/nexus/source/coordination/TrackedPromise.test.ts:132` (`await` of a
`Promise<undefined>`). 21 findings before, 19 after. `RandomSleep.test.ts:250`, an `await` of a
`Promise<void>`, is a real void and still reports.

Fixtures: `TestNoConfusingVoidExpressionJudgesVoidNotUndefined` in
`no_confusing_void_expression_test.go`, six silent rows (both real sites, and an undefined call in
statement, assigned, arrow and return position) and four reporting controls with a real `void`. It
replaced `TestNoConfusingVoidExpressionCoversUndefinedReturns`, which pinned upstream's breadth.
ESLint asks the same question through Nexus's `NexusTypeScriptEsLintPlugin`, whose wrapper drops a
report whose constrained type carries no Void flag (Nexus 4e3395d), so the engines agree on both sites
(#cn8sthd).

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

226 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomGraphCanvas.tsx:213`**

```
return () => cancelAnimationFrame(frame);
```

> Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function

**`app/(os-layout)/_components/kingdom/OsKingdomGraphCanvas.tsx:301`**

```
return () => cancelAnimationFrame(frame);
```

> Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:123`**

```
onChange={(event) => setQuery(event.target.value)}
```

> Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:133`**

```
onClick={() => setActiveType(null)}
```

> Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:144`**

```
onClick={() => setActiveType(activeType === 'Person' ? null : 'Person')}
```

> Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

