# `@typescript-eslint/class-methods-use-this`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **194** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce that class methods utilize `this`

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

194 in the tree. Showing the first few.

**`libraries/structure/command-line/Structure.ts:1603`**

```
override get identifier(): string {
```

> Expected 'this' to be used by class getter 'identifier'

**`libraries/structure/command-line/Structure.ts:1647`**

```
override helpHeader(): string {
```

> Expected 'this' to be used by class method 'helpHeader'

**`libraries/structure/command-line/Structure.ts:1656`**

```
override helpCategories(): Record<string, string[]> {
```

> Expected 'this' to be used by class method 'helpCategories'

**`libraries/structure/command-line/Structure.ts:1687`**

```
protected override formatFooterDuration(elapsedInMilliseconds: number): string {
```

> Expected 'this' to be used by class method 'formatFooterDuration'

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:462`**

```
method() {
```

> Expected 'this' to be used by class method 'method'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

