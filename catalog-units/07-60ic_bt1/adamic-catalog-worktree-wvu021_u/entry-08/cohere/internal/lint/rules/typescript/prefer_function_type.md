# `@typescript-eslint/prefer-function-type`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **3** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce using function types instead of interfaces with call signatures

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

3 in the tree. Showing the first few.

**`libraries/structure/source/components/color/hooks/useEyeDropper.ts:12`**

```
new (): {
```

> Interface only has a call signature, you should use a function type instead

**`modules/art/ArtTerminal.ts:351`**

```
(line: string): string | undefined;
```

> Interface only has a call signature, you should use a function type instead

**`modules/phi/social/PhiSocialTerminal.ts:418`**

```
(line: string): string | undefined;
```

> Interface only has a call signature, you should use a function type instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

