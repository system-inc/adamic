# `new-cap`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **61** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require constructor names to begin with a capital letter

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

61 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/code-quality/lint/utilities/CreateLintRule.ts:10`**

```
export const createLintRule = EsLintUtilities.RuleCreator(
```

> A function with a name starting with an uppercase letter should only be used as a constructor

**`libraries/structure/libraries/nexus/source/numbers/Number.ts:32`**

```
const numberGroupingSeparators = Intl.NumberFormat().format(probeNumber).replace(/\d/g, '');
```

> A function with a name starting with an uppercase letter should only be used as a constructor

**`libraries/structure/libraries/nexus/source/time/TimeZones.ts:61`**

```
return Intl.DateTimeFormat().resolvedOptions().timeZone;
```

> A function with a name starting with an uppercase letter should only be used as a constructor

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:62`**

```
const string = new stringConstructor('test');
```

> A constructor name should not start with a lowercase letter

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:63`**

```
const number = new numberConstructor(42);
```

> A constructor name should not start with a lowercase letter

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

