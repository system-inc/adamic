# `@typescript-eslint/no-unnecessary-boolean-literal-compare`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **33** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow unnecessary equality comparisons against boolean literals

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

33 in the tree. Showing the first few.

**`app/(os-layout)/finance/_components/FinanceKpiCard.tsx:82`**

```
moneyTrendCents !== undefined ? moneyTrendCents > 0 : (properties.trend ?? '').trim().startsWith('-') === false;
```

> This expression unnecessarily compares a boolean value to a boolean instead of using it directly

**`libraries/structure/libraries/nexus/code-quality/lint/rules/ConsistencyRequireConstantCasingRule.ts:1101`**

```
if(node.parent.declare === true) return;
```

> This expression unnecessarily compares a boolean value to a boolean instead of using it directly

**`libraries/structure/libraries/nexus/source/errors/Assert.test.ts:43`**

```
expect(() => assert(isNaN(Number('hello')) === false)).toThrow('Assertion failed');
```

> This expression unnecessarily compares a boolean value to a boolean instead of using it directly

**`libraries/structure/libraries/nexus/source/errors/Assert.test.ts:164`**

```
expect(() => assert(user.profile.settings.notifications === true)).not.toThrow();
```

> This expression unnecessarily compares a boolean value to a boolean instead of using it directly

**`libraries/structure/libraries/nexus/source/errors/Assert.test.ts:165`**

```
expect(() => assert(user.profile.settings.notifications === false)).toThrow();
```

> This expression unnecessarily compares a boolean value to a boolean instead of using it directly

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

