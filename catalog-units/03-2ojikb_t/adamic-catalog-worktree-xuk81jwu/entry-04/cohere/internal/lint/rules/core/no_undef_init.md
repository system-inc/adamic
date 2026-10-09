# `no-undef-init`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **17** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow initializing variables to `undefined`

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

17 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/http/Cookies.ts:67`**

```
let cookie: string | undefined = undefined;
```

> It's not necessary to initialize 'cookie: string | undefined' to undefined

**`libraries/structure/libraries/nexus/source/protocols/http/HttpError.ts:51`**

```
let cause: BaseErrorDataType | undefined = undefined;
```

> It's not necessary to initialize 'cause: BaseErrorDataType | undefined' to undefined

**`libraries/structure/source/components/buttons/AnimatedButton.tsx:371`**

```
let animatedIcon: NonLinkButtonProperties['icon'] = undefined;
```

> It's not necessary to initialize 'animatedIcon: NonLinkButtonProperties['icon']' to undefined

**`libraries/structure/source/components/buttons/AnimatedButton.tsx:372`**

```
let animatedIconLeft: NonLinkButtonProperties['iconLeft'] = undefined;
```

> It's not necessary to initialize 'animatedIconLeft: NonLinkButtonProperties['iconLeft']' to undefined

**`libraries/structure/source/components/buttons/AnimatedButton.tsx:373`**

```
let animatedIconRight: NonLinkButtonProperties['iconRight'] = undefined;
```

> It's not necessary to initialize 'animatedIconRight: NonLinkButtonProperties['iconRight']' to undefined

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.

