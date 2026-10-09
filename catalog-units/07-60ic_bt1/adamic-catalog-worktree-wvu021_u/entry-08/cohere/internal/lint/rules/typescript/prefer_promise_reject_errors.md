# `@typescript-eslint/prefer-promise-reject-errors`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **7** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Require using Error objects as Promise rejection reasons

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

7 in the tree. Showing the first few.

**`modules/samsung/frame-tv/FrameTvApi.ts:452`**

```
reject(error);
```

> Expected the Promise rejection reason to be an Error

**`modules/samsung/frame-tv/FrameTvApi.ts:465`**

```
reject(error);
```

> Expected the Promise rejection reason to be an Error

**`modules/samsung/frame-tv/FrameTvApi.ts:591`**

```
reject(error);
```

> Expected the Promise rejection reason to be an Error

**`modules/samsung/frame-tv/FrameTvApi.ts:597`**

```
reject(error);
```

> Expected the Promise rejection reason to be an Error

**`modules/samsung/frame-tv/FrameTvApi.ts:601`**

```
reject(error);
```

> Expected the Promise rejection reason to be an Error

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

