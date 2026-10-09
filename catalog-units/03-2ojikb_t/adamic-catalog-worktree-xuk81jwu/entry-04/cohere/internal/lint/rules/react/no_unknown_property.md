# `react/no-unknown-property`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **18** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow usage of unknown DOM property

## Why this recommendation

An unknown DOM property is silently dropped by React at runtime, making this a correctness rule rather than a stylistic one, and 18 auto-fixable sites is trivial.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Maybe** to **Yes**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

18 in the tree. Showing the first few.

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:16`**

```
fill-rule="evenodd"
```

> Unknown property 'fill-rule' found, use 'fillRule' instead

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:17`**

```
clip-rule="evenodd"
```

> Unknown property 'clip-rule' found, use 'clipRule' instead

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:22`**

```
fill-rule="evenodd"
```

> Unknown property 'fill-rule' found, use 'fillRule' instead

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:23`**

```
clip-rule="evenodd"
```

> Unknown property 'clip-rule' found, use 'clipRule' instead

**`libraries/structure/assets/icons/commerce/MastercardIcon.tsx:28`**

```
fill-rule="evenodd"
```

> Unknown property 'fill-rule' found, use 'fillRule' instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

