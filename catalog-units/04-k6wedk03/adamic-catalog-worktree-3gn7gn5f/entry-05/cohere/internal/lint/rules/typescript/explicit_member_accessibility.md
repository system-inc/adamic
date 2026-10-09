# `@typescript-eslint/explicit-member-accessibility`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **1942** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require explicit accessibility modifiers on class properties and methods

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

1942 in the tree. Showing the first few.

**`app/(os-layout)/_hooks/useTaskDeleteRequest.ts:26`**

```
terminalDescendantCount: number;
```

> Missing accessibility modifier on class property terminalDescendantCount

**`app/(os-layout)/_hooks/useTaskDeleteRequest.ts:27`**

```
constructor(terminalDescendantCount: number) {
```

> Missing accessibility modifier on method definition constructor

**`libraries/structure/command-line/Structure.ts:1597`**

```
title = 'Structure';
```

> Missing accessibility modifier on class property title

**`libraries/structure/command-line/Structure.ts:1603`**

```
override get identifier(): string {
```

> Missing accessibility modifier on get property accessor identifier

**`libraries/structure/command-line/Structure.ts:1612`**

```
override aliases: Record<string, string> = {
```

> Missing accessibility modifier on class property aliases

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

